package main

import (
	"testing"
	"reflect"
)

func TestConfigWithCorrectValues(t *testing.T){
	correctConfig := Config{Providers: []Provider{Provider{Name: "UptimeRobot", ApiKey: "657a68d9ashdyasjdklkskuasd", ApiURL: "https://api.uptimerobot.com/v2/", AlertContacts: "0544483_0_0-2628365_0_0-2633263_0_0"}}, EnableMonitorDeletion: true}
	config := ReadConfig("test-config.yaml")

	if ! reflect.DeepEqual(config, correctConfig) {
		t.Error("Marshalled config and correct config do not match")
	}
}

func TestConfigWithIncorrectProviderValues(t *testing.T){
	incorrectConifg := Config{Providers: []Provider{Provider{Name: "UptimeRobot2", ApiKey: "abc", ApiURL: "https://api.uptimerobot.com/v2/", AlertContacts: "0544483_0_0-2628365_0_0-2633263_0_0"}}, EnableMonitorDeletion: true}
	config := ReadConfig("test-config.yaml")

	if reflect.DeepEqual(config, incorrectConifg) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithIncorrectEnableFlag(t *testing.T){
	incorrectConifg := Config{Providers: []Provider{Provider{Name: "UptimeRobot", ApiKey: "657a68d9ashdyasjdklkskuasd", ApiURL: "https://api.uptimerobot.com/v2/", AlertContacts: "0544483_0_0-2628365_0_0-2633263_0_0"}}, EnableMonitorDeletion: false}
	config := ReadConfig("test-config.yaml")

	if reflect.DeepEqual(config, incorrectConifg) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithoutProvider(t *testing.T){
	incorrectConifg := Config{Providers: []Provider{}, EnableMonitorDeletion: false}
	config := ReadConfig("test-config.yaml")

	if reflect.DeepEqual(config, incorrectConifg) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithoutEnabledFlag(t *testing.T){
	incorrectConifg := Config{Providers: []Provider{Provider{Name: "UptimeRobot", ApiKey: "657a68d9ashdyasjdklkskuasd", ApiURL: "https://api.uptimerobot.com/v2/", AlertContacts: "0544483_0_0-2628365_0_0-2633263_0_0"}}}
	config := ReadConfig("test-config.yaml")

	if reflect.DeepEqual(config, incorrectConifg) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}

func TestConfigWithEmptyConfig(t *testing.T){
	incorrectConifg := Config{}
	config := ReadConfig("test-config.yaml")

	if reflect.DeepEqual(config, incorrectConifg) {
		t.Error("Marshalled config and incorrect config match, should not match")
	}
}