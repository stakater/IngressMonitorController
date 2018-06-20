package uptimerobot

import (
	"testing"
)

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := getControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

	m := Monitor{name: "google-test", url: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.name != m.name || mRes.url != m.url {
		t.Error("URL and name should be the same")
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := getControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

	m := Monitor{name: "google-test", url: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.name != m.name || mRes.url != m.url {
		t.Error("URL and name should be the same")
	}

	mRes.url = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.url != "https://facebook.com" {
		t.Error("URL and name should be the same")
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithIncorrectValues(t *testing.T) {
	config := getControllerConfig()

	service := UpTimeMonitorService{}
	config.Providers[0].ApiKey = "dummy-api-key"
	service.Setup(config.Providers[0])

	m := Monitor{name: "google-test", url: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}

	if mRes != nil {
		t.Error("Monitor should not be added")
	}
}
