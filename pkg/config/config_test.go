package config

import (
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

func init() {
	// To allow normal logging to be printed if tests fails
	// Dev mode is an extra feature to make output more readable
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

const (
	configFilePath                   = "../../examples/configs/test-config-uptimerobot.yaml"
	correctTestConfigName            = "UptimeRobot"
	correctTestAPIKey                = "657a68d9ashdyasjdklkskuasd"
	correctTestAPIURL                = "https://api.uptimerobot.com/v2/"
	correctTestAlertContacts         = "0544483_0_0-2628365_0_0-2633263_0_0"
	correctTestEnableMonitorDeletion = true

	configFilePathPingdom               = "../../examples/configs/test-config-pingdom.yaml"
	correctTestPingdomConfigMulti       = "Pingdom"
	correctTestPingdomAPIURL            = "https://api.pingdom.com/api/3.1"
	correctTestPingdomAlertContacts     = "0544483_0_0-2628365_0_0-2633263_0_0"
	correctTestPingdomAlertIntegrations = "91166-10924"
	correctTestPingdomAPIToken          = "657a68d9ashdyasjdklkskuasd"
	correctTestPingdomTeamAlertContacts = "1234567_0_0-2628365_0_0-2633263_0_0"

	configFilePathUptime           = "../../examples/configs/test-config-uptime.yaml"
	correctTestUptimeConfigName    = "Uptime"
	correctTestUptimeAPIURL        = "https://uptime.com/api/v1/"
	correctTestUptimeAPIKey        = "657a68d9ashdyasjdklkskuasd"
	correctTestUptimeAlertContacts = "Default"

	configFilePathAppInsights        = "../../examples/configs/test-config-appinsights.yaml"
	correctTestAppInsightsConfigName = "AppInsights"
)

func TestConfigWithCorrectValues(t *testing.T) {
	correctConfig := Config{Providers: []Provider{{Name: correctTestPingdomConfigMulti, ApiToken: correctTestPingdomAPIToken, ApiURL: correctTestPingdomAPIURL,
		AlertContacts: correctTestPingdomAlertContacts, AlertIntegrations: correctTestPingdomAlertIntegrations, TeamAlertContacts: correctTestPingdomTeamAlertContacts}},
		EnableMonitorDeletion: correctTestEnableMonitorDeletion, ResyncPeriod: 0}

	config := ReadConfig(configFilePathPingdom)
	if !reflect.DeepEqual(config, correctConfig) {
		t.Error("Marshalled config and correct config do not match")
	}
}

func TestConfigWithIncorrectProviderValues(t *testing.T) {
	incorrectConfig := Config{Providers: []Provider{{Name: "UptimeRobot2", ApiKey: "abc", ApiURL: correctTestAPIURL, AlertContacts: correctTestAlertContacts}}, EnableMonitorDeletion: correctTestEnableMonitorDeletion}
	config := ReadConfig(configFilePath)

	if reflect.DeepEqual(config, incorrectConfig) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithIncorrectEnableFlag(t *testing.T) {
	incorrectConfig := Config{Providers: []Provider{{Name: correctTestConfigName, ApiKey: correctTestAPIKey, ApiURL: correctTestAPIURL, AlertContacts: correctTestAlertContacts}}, EnableMonitorDeletion: false}
	config := ReadConfig(configFilePath)

	if reflect.DeepEqual(config, incorrectConfig) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithoutProvider(t *testing.T) {
	incorrectConfig := Config{Providers: []Provider{}, EnableMonitorDeletion: false}
	config := ReadConfig(configFilePath)

	if reflect.DeepEqual(config, incorrectConfig) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithoutEnabledFlag(t *testing.T) {
	incorrectConfig := Config{Providers: []Provider{{Name: correctTestConfigName, ApiKey: correctTestAPIKey, ApiURL: correctTestAPIURL, TeamAlertContacts: correctTestPingdomTeamAlertContacts, AlertContacts: correctTestAlertContacts}}}
	config := ReadConfig(configFilePath)

	if reflect.DeepEqual(config, incorrectConfig) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithPingdom(t *testing.T) {
	correctConfig := Config{Providers: []Provider{{Name: correctTestPingdomConfigMulti, ApiToken: correctTestPingdomAPIToken, ApiURL: correctTestPingdomAPIURL,
		AlertContacts: correctTestPingdomAlertContacts, AlertIntegrations: correctTestPingdomAlertIntegrations, TeamAlertContacts: correctTestPingdomTeamAlertContacts}},
		EnableMonitorDeletion: correctTestEnableMonitorDeletion}
	config := ReadConfig(configFilePathPingdom)
	if !reflect.DeepEqual(config, correctConfig) {
		t.Error("Marshalled config and correct config do not match")
	}
}

func TestConfigWithUptime(t *testing.T) {
	correctConfig := Config{Providers: []Provider{{Name: correctTestUptimeConfigName, ApiKey: correctTestUptimeAPIKey, ApiURL: correctTestUptimeAPIURL, AlertContacts: correctTestUptimeAlertContacts}}, EnableMonitorDeletion: correctTestEnableMonitorDeletion, ResyncPeriod: 300}
	config := ReadConfig(configFilePathUptime)
	if !reflect.DeepEqual(config, correctConfig) {
		t.Error("Marshalled config and correct config do not match")
	}
}

func TestConfigWithAppinsights(t *testing.T) {

	appinsightConfig := AppInsights{
		Name:          "demo-appinsights",
		Location:      "westeurope",
		GeoLocation:   []interface{}{"us-tx-sn1-azr", "emea-nl-ams-azr", "us-fl-mia-edge", "latam-br-gru-edge"},
		ResourceGroup: "demoRG",
		EmailAction: EmailAction{
			SendToServiceOwners: false,
			CustomEmails:        []string{"mail@cizer.dev"},
		},
		WebhookAction: WebhookAction{
			ServiceURI: "https://webhook.io",
		},
	}

	correctConfig := Config{Providers: []Provider{{Name: correctTestAppInsightsConfigName, AppInsightsConfig: appinsightConfig}}, EnableMonitorDeletion: correctTestEnableMonitorDeletion}
	config := ReadConfig(configFilePathAppInsights)
	if !reflect.DeepEqual(config, correctConfig) {
		t.Error("Marshalled config and correct config do not match")
	}
}

func TestConfigWithEmptyConfig(t *testing.T) {
	incorrectConfig := Config{}
	config := ReadConfig(configFilePath)

	if reflect.DeepEqual(config, incorrectConfig) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}
