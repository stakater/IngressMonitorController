package uptime

import (
	log "github.com/sirupsen/logrus"
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

func TestGetAllMonitors(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "Uptime")
	if provider == nil {
		panic("Failed to find provider")
	}
	// If test Config is passed skip the test
	if provider.ApiKey == "API_KEY" {
		return
	}
	service.Setup(*provider)
	monitors := service.GetAll()
	log.Println(monitors)

	if len(monitors) == 0 {
		t.Log("No Monitors Exist")
	}
	if nil == monitors {
		t.Error("Error: " + "GetAll request Failed")
	}
}

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "Uptime")
	if nil == provider {
		panic("Failed to find provider")
	}
	// If test Config is passed skip the test
	if provider.ApiKey == "API_KEY" {
		return
	}
	service.Setup(*provider)

	monitorConfig := endpointmonitorv1alpha1.UptimeConfig{
		Interval:  5,
		Locations: "US-Central",
		Contacts:  "Default",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: monitorConfig}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "Uptime")
	if provider == nil {
		panic("Failed to find provider")
	}
	// If test Config is passed skip the test
	if provider.ApiKey == "API_KEY" {
		return
	}
	service.Setup(*provider)

	monitorConfig := endpointmonitorv1alpha1.UptimeConfig{
		Interval:  5,
		Locations: "US-Central",
		Contacts:  "Default",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: monitorConfig}

	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The initial name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The initial URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}

	mRes.Name = "google-test-update"
	mRes.URL = "https://facebook.com"
	monitorConfig.Locations = "US-East"
	monitorConfig.Contacts = "Default"
	monitorConfig.Interval = 10

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test-update")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The URL should have been updated, expected: https://facebook.com, but was: " + mRes.URL)
	}
	if mRes.Name != "google-test-update" {
		t.Error("The URL should have been updated, expected: google-test-update, but was: " + mRes.Name)
	}
	// TODO: These annotations are removed so ensure that these values are returned with API call and populated in
	// Monitor.Config and then compare this from there
	// if mRes.Annotations["uptime.monitor.stakater.com/locations"] != "US-East" {
	// 	t.Error("The URL should have been updated, expected: US-East, but was: " + mRes.Annotations["uptime.monitor.stakater.com/locations"])
	// }
	// if mRes.Annotations["uptime.monitor.stakater.com/interval"] != "10" {
	// 	t.Error("The URL should have been updated, expected: 10, but was: " + mRes.Annotations["uptime.monitor.stakater.com/interval"])
	// }

	service.Remove(*mRes)
}

func TestAddMonitorWithIncorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}

	provider := util.GetProviderWithName(config, "Uptime")
	if provider == nil {
		panic("Failed to find provider")
	}
	// If test Config is passed skip the test
	if provider.ApiKey == "API_KEY" {
		return
	}
	service.Setup(*provider)

	monitorConfig := endpointmonitorv1alpha1.UptimeConfig{
		Interval:  900,
		Locations: "US-Central",
		Contacts:  "Default",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: monitorConfig}

	service.Add(m)

	_, err := service.GetByName("google-test")

	if err == nil {
		t.Error("google-test should not have existed")
	}
}
