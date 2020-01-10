// Package AppInsightsMonitor adds Azure AppInsights webtest support in IngressMonitorController
package appinsights

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/appinsights/mgmt/2015-05-01/insights"
	insightsAlert "github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

func TestAppinsightsMonitorService_createWebTest(t *testing.T) {

	location := "westeurope"
	webtestName := "foo"
	name := "foo-appinsights"
	resourceGroup := "demoRG"
	geoLocation := "us-tx-sn1-azr"
	frequency := int32(300)
	isEnabled := true
	isRetryEnabled := true
	tag := make(map[string]*string)
	tagValue := "Resource"
	tag["hidden-link:/subscriptions/99cb99da-9cf9-9999-9999-9eacc5d36a65/resourceGroups/demoRG/providers/microsoft.insights/components/foo-appinsights"] = &tagValue
	webTestConfig := "<WebTest xmlns=\"http://microsoft.com/schemas/VisualStudio/TeamTest/2010\" Name=\"\" Enabled=\"true\" Timeout=\"120\" Description=\"foo webtest is created by Ingress Monitor controller\" StopOnError=\"true\"><Items><Request Method=\"GET\" Version=\"1.1\" Url=\"https://microsoft.com\" ThinkTime=\"\" Timeout=\"0\" Encoding=\"utf-8\" ExpectedHttpStatusCode=\"200\" ExpectedResponseUrl=\"\" IgnoreHttpStatusCode=\"false\"></Request></Items></WebTest>"

	type fields struct {
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
	type args struct {
		monitor models.Monitor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   insights.WebTest
	}{
		{
			name: webtestName,
			fields: fields{
				name:           name,
				location:       location,
				emailToOwners:  true,
				subscriptionID: "99cb99da-9cf9-9999-9999-9eacc5d36a65",
				resourceGroup:  resourceGroup,
				geoLocation:    []interface{}{geoLocation},
				emailAction:    []string{"mail@cizer.dev"},
				webhookAction:  "https://webhook.io",
				ctx:            context.Background(),
			},
			args: args{
				monitor: models.Monitor{
					URL:  "https://microsoft.com",
					Name: "foo",
					Annotations: map[string]string{
						"AppInsightsStatusCodeAnnotation":   "200",
						"AppInsightsFrequency":              string(frequency),
						"AppInsightsRetryEnabledAnnotation": "true",
					},
					ID: "",
				},
			},
			want: insights.WebTest{
				Name:     &webtestName,
				Location: &location,
				Kind:     insights.Ping, // forcing type of webtest to 'ping',this could be as replace with provider configuration
				WebTestProperties: &insights.WebTestProperties{
					SyntheticMonitorID: &webtestName,
					WebTestName:        &webtestName,
					WebTestKind:        insights.Ping,
					RetryEnabled:       &isRetryEnabled,
					Enabled:            &isEnabled,
					Frequency:          &frequency,
					Locations: &[]insights.WebTestGeolocation{
						{Location: &geoLocation},
					},
					Configuration: &insights.WebTestPropertiesConfiguration{
						WebTest: &webTestConfig,
					},
				},
				Tags: tag,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aiService := &AppinsightsMonitorService{
				insightsClient:   tt.fields.insightsClient,
				alertrulesClient: tt.fields.alertrulesClient,
				name:             tt.fields.name,
				location:         tt.fields.location,
				resourceGroup:    tt.fields.resourceGroup,
				geoLocation:      tt.fields.geoLocation,
				emailAction:      tt.fields.emailAction,
				webhookAction:    tt.fields.webhookAction,
				emailToOwners:    tt.fields.emailToOwners,
				subscriptionID:   tt.fields.subscriptionID,
				ctx:              tt.fields.ctx,
			}
			if got := aiService.createWebTest(tt.args.monitor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppinsightsMonitorService.createWebTest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppinsightsMonitorService_createAlertRuleResource(t *testing.T) {

	location := "westeurope"
	name := "foo-appinsights"
	failedLocationCount := int32(1)
	period := "PT5M"
	webtestName := "foo"
	alertName := fmt.Sprintf("%s-alert", webtestName)
	resourceGroup := "demoRG"
	geoLocation := []interface{}{"us-tx-sn1-azr"}
	frequency := int32(300)
	isEnabled := true
	subID := "99cb99da-9cf9-9999-9999-9eacc5d36a65"
	emailAction := []string{"mail@cizer.dev"}
	webhookAction := "https://webhook.io"
	emailToOwners := false
	description := fmt.Sprintf("%s-alert alert is created using Ingress Monitor Controller", webtestName)
	resourceUri := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/webtests/%s", subID, resourceGroup, webtestName)
	tag := make(map[string]*string)
	tagValue := "Resource"
	tag["hidden-link:/subscriptions/99cb99da-9cf9-9999-9999-9eacc5d36a65/resourceGroups/demoRG/providers/microsoft.insights/components/foo-appinsights"] = &tagValue
	tag["hidden-link:/subscriptions/99cb99da-9cf9-9999-9999-9eacc5d36a65/resourceGroups/demoRG/providers/microsoft.insights/webtests/foo"] = &tagValue

	actions := []insightsAlert.BasicRuleAction{
		insightsAlert.RuleEmailAction{
			SendToServiceOwners: &emailToOwners,
			CustomEmails:        &(emailAction),
		},
		insightsAlert.RuleWebhookAction{
			ServiceURI: &webhookAction,
		},
	}

	type fields struct {
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
	type args struct {
		monitor models.Monitor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   insightsAlert.AlertRuleResource
	}{
		{
			name: webtestName,
			fields: fields{
				name:           name,
				location:       location,
				emailToOwners:  emailToOwners,
				subscriptionID: subID,
				resourceGroup:  resourceGroup,
				geoLocation:    geoLocation,
				emailAction:    emailAction,
				webhookAction:  webhookAction,
				ctx:            context.Background(),
			},
			args: args{
				monitor: models.Monitor{
					URL:  "https://microsoft.com",
					Name: "foo",
					Annotations: map[string]string{
						"AppInsightsStatusCodeAnnotation":   "200",
						"AppInsightsFrequency":              string(frequency),
						"AppInsightsRetryEnabledAnnotation": "true",
					},
					ID: "",
				},
			},
			want: insightsAlert.AlertRuleResource{
				Name:     &webtestName,
				Location: &location,
				ID:       &resourceUri,
				AlertRule: &insightsAlert.AlertRule{
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
				},
				Tags: tag,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aiService := &AppinsightsMonitorService{
				insightsClient:   tt.fields.insightsClient,
				alertrulesClient: tt.fields.alertrulesClient,
				name:             tt.fields.name,
				location:         tt.fields.location,
				resourceGroup:    tt.fields.resourceGroup,
				geoLocation:      tt.fields.geoLocation,
				emailAction:      tt.fields.emailAction,
				webhookAction:    tt.fields.webhookAction,
				emailToOwners:    tt.fields.emailToOwners,
				subscriptionID:   tt.fields.subscriptionID,
				ctx:              tt.fields.ctx,
			}
			if got := aiService.createAlertRuleResource(tt.args.monitor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%+v", got)
				t.Errorf("%+v", tt.want)
				t.Errorf("AppinsightsMonitorService.createAlertRuleResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
