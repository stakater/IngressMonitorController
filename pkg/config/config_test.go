package config

import (
	"reflect"
	"testing"
)

const (
	configFilePath                   = "../../configs/testConfigs/test-config.yaml"
	correctTestConfigName            = "UptimeRobot"
	correctTestAPIKey                = "657a68d9ashdyasjdklkskuasd"
	correctTestAPIURL                = "https://api.uptimerobot.com/v2/"
	correctTestAlertContacts         = "0544483_0_0-2628365_0_0-2633263_0_0"
	correctTestEnableMonitorDeletion = true

	configFilePathPingdom               = "../../configs/testConfigs/test-config-pingdom.yaml"
	correctTestPingdomConfigMulti       = "PingdomMulti"
	correctTestPingdomAPIToken          = "657a68d9ashdyasjdklkskuasd"
	correctTestPingdomAlertContacts     = "0544483_0_0-2628365_0_0-2633263_0_0"
	correctTestPingdomAlertIntegrations = "91166,10924"

	configFilePathUptime           = "../../configs/testConfigs/test-config-uptime.yaml"
	correctTestUptimeConfigName    = "Uptime"
	correctTestUptimeAPIURL        = "https://uptime.com/api/v1/"
	correctTestUptimeAPIKey        = "657a68d9ashdyasjdklkskuasd"
	correctTestUptimeAlertContacts = "Default"

	configFilePathAppInsights        = "../../configs/testConfigs/test-config-appinsights.yaml"
	correctTestAppInsightsConfigName = "AppInsights"
)

func TestConfigWithCorrectValues(t *testing.T) {
	correctConfig := Config{Providers: []Provider{Provider{Name: correctTestPingdomConfigMulti, ApiToken: correctTestPingdomAPIToken,
		AlertContacts: correctTestPingdomAlertContacts, AlertIntegrations: correctTestPingdomAlertIntegrations}},
		EnableMonitorDeletion: correctTestEnableMonitorDeletion, ResyncPeriod: 0}

	config := ReadConfig(configFilePathPingdom)
	if !reflect.DeepEqual(config, correctConfig) {
		t.Error("Marshalled config and correct config do not match")
	}
}

func TestConfigWithIncorrectProviderValues(t *testing.T) {
	incorrectConfig := Config{Providers: []Provider{Provider{Name: "UptimeRobot2", ApiKey: "abc", ApiURL: correctTestAPIURL, AlertContacts: correctTestAlertContacts}}, EnableMonitorDeletion: correctTestEnableMonitorDeletion}
	config := ReadConfig(configFilePath)

	if reflect.DeepEqual(config, incorrectConfig) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithIncorrectEnableFlag(t *testing.T) {
	incorrectConfig := Config{Providers: []Provider{Provider{Name: correctTestConfigName, ApiKey: correctTestAPIKey, ApiURL: correctTestAPIURL, AlertContacts: correctTestAlertContacts}}, EnableMonitorDeletion: false}
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
	incorrectConfig := Config{Providers: []Provider{Provider{Name: correctTestConfigName, ApiKey: correctTestAPIKey, ApiURL: correctTestAPIURL, AlertContacts: correctTestAlertContacts}}}
	config := ReadConfig(configFilePath)

	if reflect.DeepEqual(config, incorrectConfig) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithPingdomMultiAuthEnabledFlag(t *testing.T) {
	correctConfig := Config{Providers: []Provider{Provider{Name: correctTestPingdomConfigMulti, ApiToken: correctTestPingdomAPIToken,
		AlertContacts: correctTestPingdomAlertContacts, AlertIntegrations: correctTestPingdomAlertIntegrations}},
		EnableMonitorDeletion: correctTestEnableMonitorDeletion}
	config := ReadConfig(configFilePathPingdom)
	if !reflect.DeepEqual(config, correctConfig) {
		t.Error("Marshalled config and correct config do not match")
	}
}

func TestConfigWithUptime(t *testing.T) {
	correctConfig := Config{Providers: []Provider{Provider{Name: correctTestUptimeConfigName, ApiKey: correctTestUptimeAPIKey, ApiURL: correctTestUptimeAPIURL, AlertContacts: correctTestUptimeAlertContacts}}, EnableMonitorDeletion: correctTestEnableMonitorDeletion, ResyncPeriod: 300}
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

	correctConfig := Config{Providers: []Provider{Provider{Name: correctTestAppInsightsConfigName, AppInsightsConfig: appinsightConfig}}, EnableMonitorDeletion: correctTestEnableMonitorDeletion}
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
