// Package AppInsightsMonitor adds Azure AppInsights webtest support in IngressMonitorController
package appinsights

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/google/go-cmp/cmp"
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// To allow normal logging to be printed if tests fails
	// Dev mode is an extra feature to make output more readable
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

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
	type args struct {
		monitor models.Monitor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   armapplicationinsights.WebTest
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
				emailAction:    []*string{to.Ptr("mail@cizer.dev")},
				webhookAction:  "https://webhook.io",
				ctx:            context.Background(),
			},
			args: args{
				monitor: models.Monitor{
					URL:  "https://microsoft.com",
					Name: "foo",
					Config: endpointmonitorv1alpha1.AppInsightsConfig{
						StatusCode:  200,
						Frequency:   300,
						RetryEnable: true,
					},
					ID: "",
				},
			},
			want: armapplicationinsights.WebTest{
				Name:     &webtestName,
				Location: &location,
				Kind:     to.Ptr(armapplicationinsights.WebTestKindPing), // forcing type of webtest to 'ping',this could be as replace with provider configuration
				Properties: &armapplicationinsights.WebTestProperties{
					SyntheticMonitorID: &webtestName,
					WebTestName:        &webtestName,
					WebTestKind:        to.Ptr(armapplicationinsights.WebTestKindPing),
					RetryEnabled:       &isRetryEnabled,
					Enabled:            &isEnabled,
					Frequency:          &frequency,
					Locations: []*armapplicationinsights.WebTestGeolocation{
						{Location: &geoLocation},
					},
					Configuration: &armapplicationinsights.WebTestPropertiesConfiguration{
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
			if got := aiService.createWebTest(tt.args.monitor); !cmp.Equal(got, tt.want) {
				t.Errorf("AppinsightsMonitorService.createWebTest() = %v", cmp.Diff(got, tt.want))
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
	isEnabled := true
	subID := "99cb99da-9cf9-9999-9999-9eacc5d36a65"
	emailAction := []*string{to.Ptr("mail@cizer.dev")}
	webhookAction := "https://webhook.io"
	emailToOwners := false
	description := fmt.Sprintf("%s-alert alert is created using Ingress Monitor Controller", webtestName)
	resourceUri := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/webtests/%s", subID, resourceGroup, webtestName)
	tag := make(map[string]*string)
	tagValue := "Resource"
	tag["hidden-link:/subscriptions/99cb99da-9cf9-9999-9999-9eacc5d36a65/resourceGroups/demoRG/providers/microsoft.insights/components/foo-appinsights"] = &tagValue
	tag["hidden-link:/subscriptions/99cb99da-9cf9-9999-9999-9eacc5d36a65/resourceGroups/demoRG/providers/microsoft.insights/webtests/foo"] = &tagValue

	actions := []armmonitor.RuleActionClassification{
		&armmonitor.RuleEmailAction{
			SendToServiceOwners: &emailToOwners,
			CustomEmails:        emailAction,
		},
		&armmonitor.RuleWebhookAction{
			ServiceURI: &webhookAction,
		},
	}

	type fields struct {
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
	type args struct {
		monitor models.Monitor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   armmonitor.AlertRuleResource
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
					Config: endpointmonitorv1alpha1.AppInsightsConfig{
						StatusCode:  200,
						Frequency:   300,
						RetryEnable: true,
					},
					ID: "",
				},
			},
			want: armmonitor.AlertRuleResource{
				Name:     &webtestName,
				Location: &location,
				ID:       &resourceUri,
				Properties: &armmonitor.AlertRule{
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
			if got := aiService.createAlertRuleResource(tt.args.monitor); !cmp.Equal(got, tt.want) {
				t.Errorf("AppinsightsMonitorService.createAlertRuleResource() = %s", cmp.Diff(got, tt.want))
			}
		})
	}
}
