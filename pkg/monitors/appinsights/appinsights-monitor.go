// Package AppInsightsMonitor adds Azure AppInsights webtest support in IngressMonitorController
package appinsights

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"

	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// Default value for monitor configuration
	AppInsightsStatusCodeDefaultValue   = http.StatusOK
	AppInsightsRetryEnabledDefaultValue = true
	AppInsightsFrequencyDefaultValue    = 300
)

var log = logf.Log.WithName("appinsights-monitor")

// Configuration holds appinsights specific configuration
type Configuration struct {
	isRetryEnabled     bool
	expectedStatusCode int
	frequency          int32
}

// AppinsightsMonitorService struct contains parameters required by appinsights go client
type AppinsightsMonitorService struct {
	insightsClient   *armapplicationinsights.WebTestsClient
	alertrulesClient *armmonitor.AlertRulesClient
	name             string
	location         string
	resourceGroup    string
	geoLocation      []interface{}
	emailAction      []*string
	webhookAction    string
	emailToOwners    bool
	subscriptionID   string
	ctx              context.Context
}

type WebTest struct {
	XMLName     xml.Name `xml:"WebTest"`
	Xmlns       string   `xml:"xmlns,attr"`
	Name        string   `xml:"Name,attr"`
	Enabled     bool     `xml:"Enabled,attr"`
	Timeout     string   `xml:"Timeout,attr"`
	Description string   `xml:"Description,attr"`
	StopOnError bool     `xml:"StopOnError,attr"`
	Items       struct {
		Request struct {
			Method                 string `xml:"Method,attr"`
			Version                string `xml:"Version,attr"`
			URL                    string `xml:"Url,attr"`
			ThinkTime              string `xml:"ThinkTime,attr"`
			Timeout                int    `xml:"Timeout,attr"`
			Encoding               string `xml:"Encoding,attr"`
			ExpectedHttpStatusCode int    `xml:"ExpectedHttpStatusCode,attr"`
			ExpectedResponseUrl    string `xml:"ExpectedResponseUrl,attr"`
			IgnoreHttpStatusCode   bool   `xml:"IgnoreHttpStatusCode,attr"`
		} `xml:"Request"`
	} `xml:"Items"`
}

// NewWebTest() initialize WebTest with default values
func NewWebTest() *WebTest {
	w := WebTest{
		XMLName:     xml.Name{Local: "WebTest"},
		Xmlns:       "http://microsoft.com/schemas/VisualStudio/TeamTest/2010",
		Enabled:     true,
		Timeout:     "120",
		StopOnError: true,
	}
	w.Items.Request.Encoding = "utf-8"
	w.Items.Request.Version = "1.1"
	w.Items.Request.Method = "GET"
	w.Items.Request.IgnoreHttpStatusCode = false

	return &w
}

func (monitor *AppinsightsMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO: Retrieve oldMonitor config and compare it here
	return false
}

// Setup method will initialize a appinsights's go client
func (aiService *AppinsightsMonitorService) Setup(provider config.Provider) {

	log.Info("AppInsights Monitor's Setup has been called. Initializing AppInsights Client..")

	aiService.ctx = context.Background()
	aiService.name = provider.AppInsightsConfig.Name
	aiService.location = provider.AppInsightsConfig.Location
	aiService.resourceGroup = provider.AppInsightsConfig.ResourceGroup
	aiService.geoLocation = provider.AppInsightsConfig.GeoLocation
	aiService.emailAction = provider.AppInsightsConfig.EmailAction.CustomEmails
	aiService.emailToOwners = provider.AppInsightsConfig.EmailAction.SendToServiceOwners
	aiService.webhookAction = provider.AppInsightsConfig.WebhookAction.ServiceURI
	aiService.subscriptionID = provider.AppInsightsConfig.SubscriptionId

	// For backward compatibility when subscriptionID was supplied via environment variable
	if aiService.subscriptionID == "" {
		if sid := os.Getenv("AZURE_SUBSCRIPTION_ID"); sid != "" {
			aiService.subscriptionID = sid
		} else {
			log.Error(nil, "Error fetching environment variable")
			os.Exit(1)
		}
	}

	creds, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err, "Error initializing AppInsights Client")
		os.Exit(1)
	}

	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Telemetry: policy.TelemetryOptions{
				ApplicationID: "appInsightsMonitor",
			},
		},
	}

	aiService.insightsClient, err = armapplicationinsights.NewWebTestsClient(aiService.subscriptionID, creds, clientOptions)

	if err != nil {
		log.Error(err, "Error initializing AppInsights Client")
		os.Exit(1)
	}

	log.Info("AppInsights Insights Client has been initialized")

	// initialize monitoring alertrule client only if Email Action or Webhook Action is specified.
	if aiService.isAlertEnabled() {
		aiService.alertrulesClient, err = armmonitor.NewAlertRulesClient(aiService.subscriptionID, creds, clientOptions)
		if err != nil {
			log.Error(err, "Error initializing AppInsights Alertrules Client")
			os.Exit(1)
		}
		log.Info("AppInsights Alertrules Client has been initialized")
	}

	log.Info("AppInsights Monitor has been initialized")
}

// GetAll function will return all monitors (appinsights webtest) object in an array
// GetAll for AppInsights returns all webtest for specific component in a resource group.
func (aiService *AppinsightsMonitorService) GetAll() []models.Monitor {

	log.Info("AppInsight monitor's GetAll method has been called")

	var monitors []models.Monitor

	pager := aiService.insightsClient.NewListByComponentPager(aiService.name, aiService.resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(aiService.ctx)
		if err != nil {
			return monitors
		}

		for _, wt := range page.Value {
			m := models.Monitor{
				Name: *wt.Name,
				ID:   *wt.ID,
				URL:  getURL(*wt.Properties.Configuration.WebTest),
			}
			monitors = append(monitors, m)
		}
	}
	return monitors

}

// GetByName function will return a  monitors (appinsights webtest) object based on the name provided
// GetAll for AppInsights returns a webtest for specific resource group.
func (aiService *AppinsightsMonitorService) GetByName(monitorName string) (*models.Monitor, error) {

	log.Info("AppInsights Monitor's GetByName method has been called")
	webtest, err := aiService.insightsClient.Get(aiService.ctx, aiService.resourceGroup, monitorName, nil)
	if err != nil {
		var re *azcore.ResponseError
		if errors.As(err, &re) {
			if re.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("Application Insights WebTest %s was not found in Resource Group %s", monitorName, aiService.resourceGroup)
			}
		}
		return nil, fmt.Errorf("Error retrieving Application Insights WebTests %s (Resource Group %s): %v", monitorName, aiService.resourceGroup, err)
	}
	return &models.Monitor{
		Name: *webtest.Name,
		URL:  getURL(*webtest.Properties.Configuration.WebTest),
		ID:   *webtest.ID,
	}, nil

}

// Add function method will add a monitor
func (aiService *AppinsightsMonitorService) Add(monitor models.Monitor) {

	log.Info("AppInsights Monitor's Add method has been called")
	log.Info(fmt.Sprintf("Adding Application Insights WebTest '%s' from '%s'", monitor.Name, aiService.name))
	webtest := aiService.createWebTest(monitor)
	_, err := aiService.insightsClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, monitor.Name, webtest, nil)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error adding Application Insights WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err))
	} else {
		log.Info(fmt.Sprintf("Successfully added Application Insights WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup))
		if aiService.isAlertEnabled() {
			log.Info(fmt.Sprintf("Adding alert rule for WebTest '%s' from '%s'", monitor.Name, aiService.name))
			alertName := fmt.Sprintf("%s-alert", monitor.Name)
			webtestAlert := aiService.createAlertRuleResource(monitor)
			_, err := aiService.alertrulesClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, alertName, webtestAlert, nil)
			if err != nil {
				log.Error(err, fmt.Sprintf("Error adding alert rule for WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err))
			}
			log.Info(fmt.Sprintf("Successfully added Alert rule for WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup))
		}
	}

}

// Update method will update a monitor
func (aiService *AppinsightsMonitorService) Update(monitor models.Monitor) {

	log.Info("AppInsights Monitor's Update method has been called")
	log.Info(fmt.Sprintf("Updating Application Insights WebTest '%s' from '%s'", monitor.Name, aiService.name))

	webtest := aiService.createWebTest(monitor)
	_, err := aiService.insightsClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, monitor.Name, webtest, nil)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error updating Application Insights WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err))
	} else {
		log.Info(fmt.Sprintf("Successfully updated Application Insights WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup))
		if aiService.isAlertEnabled() {
			log.Info(fmt.Sprintf("Updating alert rule for WebTest '%s' from '%s'", monitor.Name, aiService.name))
			alertName := fmt.Sprintf("%s-alert", monitor.Name)
			webtestAlert := aiService.createAlertRuleResource(monitor)
			_, err := aiService.alertrulesClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, alertName, webtestAlert, nil)
			if err != nil {
				log.Error(err, fmt.Sprintf("Error updating alert rule for WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err))
			}
			log.Info(fmt.Sprintf("Successfully updating Alert rule for WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup))
		}
	}
}

// Remove method will remove a monitor
func (aiService *AppinsightsMonitorService) Remove(monitor models.Monitor) {

	log.Info("AppInsights Monitor's Remove method has been called")
	log.Info(fmt.Sprintf("Deleting Application Insights WebTest '%s' from '%s'", monitor.Name, aiService.name))
	_, err := aiService.insightsClient.Delete(aiService.ctx, aiService.resourceGroup, monitor.Name, nil)
	if err != nil {
		var re *azcore.ResponseError
		if errors.As(err, &re) {
			if re.StatusCode == http.StatusNotFound {
				log.Error(err, fmt.Sprintf("Application Insights WebTest %s was not found in Resource Group %s", monitor.Name, aiService.resourceGroup))
			}
		}
		log.Error(err, fmt.Sprintf("Error deleting Application Insights WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err))
	} else {
		log.Info(fmt.Sprintf("Successfully removed Application Insights WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup))
		if aiService.isAlertEnabled() {
			log.Info(fmt.Sprintf("Deleting alert rule for WebTest '%s' from '%s'", monitor.Name, aiService.name))
			alertName := fmt.Sprintf("%s-alert", monitor.Name)
			_, err := aiService.alertrulesClient.Delete(aiService.ctx, aiService.resourceGroup, alertName, nil)
			if err != nil {
				var re *azcore.ResponseError
				if errors.As(err, &re) {
					if re.StatusCode == http.StatusNotFound {
						log.Error(err, fmt.Sprintf("WebTest Alert rule %s was not found in Resource Group %s", alertName, aiService.resourceGroup))
					}
				}
				log.Error(err, fmt.Sprintf("Error deleting alert rule for WebTests %s (Resource Group %s): %v", alertName, aiService.resourceGroup, err))
			}
			log.Info(fmt.Sprintf("Successfully removed Alert rule for WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup))
		}
	}
}

// createWebTest forms xml configuration for Appinsights WebTest
func (aiService *AppinsightsMonitorService) createWebTest(monitor models.Monitor) armapplicationinsights.WebTest {

	isEnabled := true
	webtest := NewWebTest()
	configs := getConfiguration(monitor)

	webtest.Description = fmt.Sprintf("%s webtest is created by Ingress Monitor controller", monitor.Name)
	webtest.Items.Request.URL = monitor.URL
	webtest.Items.Request.ExpectedHttpStatusCode = configs.expectedStatusCode

	xmlByte, err := xml.Marshal(webtest)
	if err != nil {
		log.Error(err, "Error encoding XML WebTest Configuration")
	}
	webtestConfig := string(xmlByte)
	pingKind := armapplicationinsights.WebTestKindPing
	return armapplicationinsights.WebTest{
		Name:     &monitor.Name,
		Location: &aiService.location,
		Kind:     &pingKind, // forcing type of webtest to 'ping',this could be as replace with provider configuration
		Properties: &armapplicationinsights.WebTestProperties{
			SyntheticMonitorID: &monitor.Name,
			WebTestName:        &monitor.Name,
			WebTestKind:        &pingKind,
			RetryEnabled:       &configs.isRetryEnabled,
			Enabled:            &isEnabled,
			Frequency:          &configs.frequency,
			Locations:          getGeoLocation(aiService.geoLocation),
			Configuration: &armapplicationinsights.WebTestPropertiesConfiguration{
				WebTest: &webtestConfig,
			},
		},
		Tags: aiService.getTags("webtest", monitor.Name),
	}

}

// createWebTestAlert forms xml configuration for Appinsights WebTest
func (aiService *AppinsightsMonitorService) createAlertRuleResource(monitor models.Monitor) armmonitor.AlertRuleResource {

	isEnabled := aiService.isAlertEnabled()
	failedLocationCount := int32(1)
	period := "PT5M"
	alertName := fmt.Sprintf("%s-alert", monitor.Name)
	description := fmt.Sprintf("%s alert is created using Ingress Monitor Controller", alertName)
	resourceUri := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/webtests/%s", aiService.subscriptionID, aiService.resourceGroup, monitor.Name)

	actions := make([]armmonitor.RuleActionClassification, 0, 2)

	if len(aiService.emailAction) > 0 {
		actions = append(actions, &armmonitor.RuleEmailAction{
			SendToServiceOwners: &aiService.emailToOwners,
			CustomEmails:        aiService.emailAction,
		})
	}

	if aiService.webhookAction != "" {
		actions = append(actions, &armmonitor.RuleWebhookAction{
			ServiceURI: &aiService.webhookAction,
		})
	}

	alertRule := armmonitor.AlertRule{
		Name:        &alertName,
		IsEnabled:   &isEnabled,
		Description: &description,
		Condition: &armmonitor.LocationThresholdRuleCondition{
			DataSource: &armmonitor.RuleMetricDataSource{
				ResourceURI: &resourceUri,
				ODataType:   to.Ptr("Microsoft.Azure.Management.Insights.Models.RuleMetricDataSource"),
				MetricName:  &alertName,
			},
			FailedLocationCount: &failedLocationCount,
			WindowSize:          &period,
			ODataType:           to.Ptr("Microsoft.Azure.Management.Insights.Models.LocationThresholdRuleCondition"),
		},
		Actions: actions,
	}

	return armmonitor.AlertRuleResource{
		Location:   &aiService.location,
		Properties: &alertRule,
		Tags:       aiService.getTags("alert", monitor.Name),
		ID:         &resourceUri,
		Name:       &monitor.Name,
	}

}
