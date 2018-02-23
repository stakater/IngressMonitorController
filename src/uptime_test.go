package main

import (
	"testing"
)

func TestAddRemoveMonitorWithCorrectValues(t *testing.T) {
	service := UpTimeMonitorService{}
	apiKey := "u544483-b3647f3e973b66417071a555"
	service.Setup(apiKey, "https://api.uptimerobot.com/v2/", "0544483_0_0-2628365_0_0-2633263_0_0")

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
