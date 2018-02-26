package main

import (
	"testing"
)

func TestAddRemoveMonitorWithCorrectValues(t *testing.T) {
	config := getControllerConfig()

	service := UpTimeMonitorService{}
	apiKey := config.Providers[0].ApiKey
	alertContacts := config.Providers[0].AlertContacts
	url := config.Providers[0].ApiURL
	service.Setup(apiKey, url, alertContacts)

	m := Monitor{name: "google-test", url: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.name != mRes.name || mRes.url != m.url {
		t.Error("URL and name should be the same")
	}
	service.Remove(*mRes)
}
