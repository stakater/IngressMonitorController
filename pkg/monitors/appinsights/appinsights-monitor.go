// Package AppInsightsMonitor adds Azure AppInsights webtest support in IngressMonitorController
package appinsights

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/services/appinsights/mgmt/2015-05-01/insights"
	insightsAlert "github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/apex/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

const (
	// Default value for monitor configuration
	AppInsightsStatusCodeDefaultValue   = http.StatusOK
	AppInsightsRetryEnabledDefaultValue = true
	AppInsightsFrequencyDefaultValue    = 300
)

// Configuration holds appinsights specific configuration
type Configuration struct {
	isRetryEnabled     bool
	expectedStatusCode int
	frequency          int32
}

// AppinsightsMonitorService struct contains parameters required by appinsights go client
type AppinsightsMonitorService struct {
	insightsClient   insights.WebTestsClient
	alertrulesClient insightsAlert.AlertRulesClient
	name             string
	location         string
	resourceGroup    string
	geoLocation      []interface{}
	emailAction      []string
	webhookAction    string
	emailToOwners    bool
	subscriptionID   string
	ctx              context.Context
}

// AzureConfig holds service principle credentials required of auth
type AzureConfig struct {
	Subscription_ID string
	Client_ID       string
	Client_Secret   string
	Tenant_ID       string
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

	var azConfig AzureConfig
	err := envconfig.Process("AZURE", &azConfig)
	if err != nil {
		log.Fatalf("Error fetching environment variable: %s", err.Error())
	}

	aiService.ctx = context.Background()
	aiService.name = provider.AppInsightsConfig.Name
	aiService.location = provider.AppInsightsConfig.Location
	aiService.resourceGroup = provider.AppInsightsConfig.ResourceGroup
	aiService.geoLocation = provider.AppInsightsConfig.GeoLocation
	aiService.emailAction = provider.AppInsightsConfig.EmailAction.CustomEmails
	aiService.emailToOwners = provider.AppInsightsConfig.EmailAction.SendToServiceOwners
	aiService.webhookAction = provider.AppInsightsConfig.WebhookAction.ServiceURI
	aiService.subscriptionID = azConfig.Subscription_ID

	// Generate clientConfig based on Azure Credentials (Service Principle)
	clientConfig := auth.NewClientCredentialsConfig(azConfig.Client_ID, azConfig.Client_Secret, azConfig.Tenant_ID)

	// initialize appinsights client
	err = aiService.insightsClient.AddToUserAgent("appInsightsMonitor")
	if err != nil {
		log.Fatal("Error adding UserAgent in AppInsights Client")
	}

	aiService.insightsClient = insights.NewWebTestsClient(azConfig.Subscription_ID)
	if err != nil {
		log.Fatal("Error initializing AppInsights Client")
	}

	aiService.insightsClient.Authorizer, err = clientConfig.Authorizer()
	if err != nil {
		log.Fatal("Error initializing AppInsights Client")
	}

	log.Info("AppInsights Insights Client has been initialized")

	// initialize monitoring alertrule client only if Email Action or Webhook Action is specified.
	if aiService.isAlertEnabled() {
		aiService.alertrulesClient = insightsAlert.NewAlertRulesClient(azConfig.Subscription_ID)
		aiService.alertrulesClient.Authorizer, err = clientConfig.Authorizer()
		if err != nil {
			log.Fatal("Error initializing AppInsights Alertrules Client")
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

	webtests, err := aiService.insightsClient.ListByComponent(aiService.ctx, aiService.name, aiService.resourceGroup)
	if err != nil {
		if webtests.Response().StatusCode == http.StatusNotFound {
			return monitors
		}
		return monitors
	}
	for _, webtest := range webtests.Values() {

		newMonitor := models.Monitor{
			Name: *webtest.Name,
			URL:  getURL(*webtest.Configuration.WebTest),
			ID:   *webtest.ID,
		}
		monitors = append(monitors, newMonitor)
	}

	return monitors

}

// GetByName function will return a  monitors (appinsights webtest) object based on the name provided
// GetAll for AppInsights returns a webtest for specific resource group.
func (aiService *AppinsightsMonitorService) GetByName(monitorName string) (*models.Monitor, error) {

	log.Info("AppInsights Monitor's GetByName method has been called")
	webtest, err := aiService.insightsClient.Get(aiService.ctx, aiService.resourceGroup, monitorName)
	if err != nil {
		if webtest.Response.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("Application Insights WebTest %s was not found in Resource Group %s", monitorName, aiService.resourceGroup)
		}
		return nil, fmt.Errorf("Error retrieving Application Insights WebTests %s (Resource Group %s): %v", monitorName, aiService.resourceGroup, err)
	}
	return &models.Monitor{
		Name: *webtest.Name,
		URL:  getURL(*webtest.Configuration.WebTest),
		ID:   *webtest.ID,
	}, nil

}

// Add function method will add a monitor
func (aiService *AppinsightsMonitorService) Add(monitor models.Monitor) {

	log.Info("AppInsights Monitor's Add method has been called")
	log.Info("Adding Application Insights WebTest '%s' from '%s'", monitor.Name, aiService.name)
	webtest := aiService.createWebTest(monitor)
	_, err := aiService.insightsClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, monitor.Name, webtest)
	if err != nil {
		log.Errorf("Error adding Application Insights WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err)
	} else {
		log.Info("Successfully added Application Insights WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup)
		if aiService.isAlertEnabled() {
			log.Info("Adding alert rule for WebTest '%s' from '%s'", monitor.Name, aiService.name)
			alertName := fmt.Sprintf("%s-alert", monitor.Name)
			webtestAlert := aiService.createAlertRuleResource(monitor)
			_, err := aiService.alertrulesClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, alertName, webtestAlert)
			if err != nil {
				log.Errorf("Error adding alert rule for WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err)
			}
			log.Info("Successfully added Alert rule for WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup)
		}
	}

}

// Update method will update a monitor
func (aiService *AppinsightsMonitorService) Update(monitor models.Monitor) {

	log.Info("AppInsights Monitor's Update method has been called")
	log.Info("Updating Application Insights WebTest '%s' from '%s'", monitor.Name, aiService.name)

	webtest := aiService.createWebTest(monitor)
	_, err := aiService.insightsClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, monitor.Name, webtest)
	if err != nil {
		log.Errorf("Error updating Application Insights WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err)
	} else {
		log.Info("Successfully updated Application Insights WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup)
		if aiService.isAlertEnabled() {
			log.Info("Updating alert rule for WebTest '%s' from '%s'", monitor.Name, aiService.name)
			alertName := fmt.Sprintf("%s-alert", monitor.Name)
			webtestAlert := aiService.createAlertRuleResource(monitor)
			_, err := aiService.alertrulesClient.CreateOrUpdate(aiService.ctx, aiService.resourceGroup, alertName, webtestAlert)
			if err != nil {
				log.Errorf("Error updating alert rule for WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err)
			}
			log.Info("Successfully updating Alert rule for WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup)
		}
	}
}

// Remove method will remove a monitor
func (aiService *AppinsightsMonitorService) Remove(monitor models.Monitor) {

	log.Info("AppInsights Monitor's Remove method has been called")
	log.Info("Deleting Application Insights WebTest '%s' from '%s'", monitor.Name, aiService.name)
	r, err := aiService.insightsClient.Delete(aiService.ctx, aiService.resourceGroup, monitor.Name)
	if err != nil {
		if r.Response.StatusCode == http.StatusNotFound {
			log.Errorf("Application Insights WebTest %s was not found in Resource Group %s", monitor.Name, aiService.resourceGroup)
		}
		log.Errorf("Error deleting Application Insights WebTests %s (Resource Group %s): %v", monitor.Name, aiService.resourceGroup, err)
	} else {
		log.Info("Successfully removed Application Insights WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup)
		if aiService.isAlertEnabled() {
			log.Info("Deleting alert rule for WebTest '%s' from '%s'", monitor.Name, aiService.name)
			alertName := fmt.Sprintf("%s-alert", monitor.Name)
			r, err := aiService.alertrulesClient.Delete(aiService.ctx, aiService.resourceGroup, alertName)
			if err != nil {
				if r.Response.StatusCode == http.StatusNotFound {
					log.Errorf("WebTest Alert rule %s was not found in Resource Group %s", alertName, aiService.resourceGroup)
				}
				log.Errorf("Error deleting alert rule for WebTests %s (Resource Group %s): %v", alertName, aiService.resourceGroup, err)
			}
			log.Info("Successfully removed Alert rule for WebTest %s (Resource Group %s)", monitor.Name, aiService.resourceGroup)
		}
	}
}

// createWebTest forms xml configuration for Appinsights WebTest
func (aiService *AppinsightsMonitorService) createWebTest(monitor models.Monitor) insights.WebTest {

	isEnabled := true
	webtest := NewWebTest()
	configs := getConfiguration(monitor)

	webtest.Description = fmt.Sprintf("%s webtest is created by Ingress Monitor controller", monitor.Name)
	webtest.Items.Request.URL = monitor.URL
	webtest.Items.Request.ExpectedHttpStatusCode = configs.expectedStatusCode

	xmlByte, err := xml.Marshal(webtest)
	if err != nil {
		log.Error("Error encoding XML WebTest Configuration")
	}
	webtestConfig := string(xmlByte)
	return insights.WebTest{
		Name:     &monitor.Name,
		Location: &aiService.location,
		Kind:     insights.Ping, // forcing type of webtest to 'ping',this could be as replace with provider configuration
		WebTestProperties: &insights.WebTestProperties{
			SyntheticMonitorID: &monitor.Name,
			WebTestName:        &monitor.Name,
			WebTestKind:        insights.Ping,
			RetryEnabled:       &configs.isRetryEnabled,
			Enabled:            &isEnabled,
			Frequency:          &configs.frequency,
			Locations:          getGeoLocation(aiService.geoLocation),
			Configuration: &insights.WebTestPropertiesConfiguration{
				WebTest: &webtestConfig,
			},
		},
		Tags: aiService.getTags("webtest", monitor.Name),
	}

}

// createWebTestAlert forms xml configuration for Appinsights WebTest
func (aiService *AppinsightsMonitorService) createAlertRuleResource(monitor models.Monitor) insightsAlert.AlertRuleResource {

	isEnabled := aiService.isAlertEnabled()
	failedLocationCount := int32(1)
	period := "PT5M"
	alertName := fmt.Sprintf("%s-alert", monitor.Name)
	description := fmt.Sprintf("%s alert is created using Ingress Monitor Controller", alertName)
	resourceUri := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/webtests/%s", aiService.subscriptionID, aiService.resourceGroup, monitor.Name)

	actions := make([]insightsAlert.BasicRuleAction, 0, 2)

	if len(aiService.emailAction) > 0 {
		actions = append(actions, insightsAlert.RuleEmailAction{
			SendToServiceOwners: &aiService.emailToOwners,
			CustomEmails:        &(aiService.emailAction),
		})
	}

	if aiService.webhookAction != "" {
		actions = append(actions, insightsAlert.RuleWebhookAction{
			ServiceURI: &aiService.webhookAction,
		})
	}

	alertRule := insightsAlert.AlertRule{
		Name:        &alertName,
		IsEnabled:   &isEnabled,
		Description: &description,
		Condition: &insightsAlert.LocationThresholdRuleCondition{
			DataSource: insightsAlert.RuleMetricDataSource{
				ResourceURI: &resourceUri,
				OdataType:   insightsAlert.OdataTypeMicrosoftAzureManagementInsightsModelsRuleMetricDataSource,
				MetricName:  &alertName,
			},
			FailedLocationCount: &failedLocationCount,
			WindowSize:          &period,
			OdataType:           insightsAlert.OdataTypeMicrosoftAzureManagementInsightsModelsLocationThresholdRuleCondition,
		},
		Actions: &actions,
	}

	return insightsAlert.AlertRuleResource{
		Name:      &monitor.Name,
		Location:  &aiService.location,
		AlertRule: &alertRule,
		ID:        &resourceUri,
		Tags:      aiService.getTags("alert", monitor.Name),
	}
}
