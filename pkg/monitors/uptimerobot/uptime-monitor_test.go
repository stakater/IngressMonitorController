package uptimerobot

import (
	"testing"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Error("URL and name should be the same")
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Error("URL and name should be the same")
	}

	mRes.URL = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("URL and name should be the same")
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

    var annotations = map[string]string {
        "uptimerobot.monitor.stakater.com/interval": "600",
    }

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name || mRes.URL != m.URL || "600" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("URL, name and interval should be the same")
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

    var annotations = map[string]string {
        "uptimerobot.monitor.stakater.com/interval": "600",
    }

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name || mRes.URL != m.URL || "600" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("URL, name and interval should be the same")
	}

	mRes.URL = "https://facebook.com"
    annotations["uptimerobot.monitor.stakater.com/interval"] = "900"
	mRes.Annotations = annotations

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" || "900" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("URL and interval should be the same.")
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithIncorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	config.Providers[0].ApiKey = "dummy-api-key"
	service.Setup(config.Providers[0])

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}

	if mRes != nil {
		t.Error("Monitor should not be added")
	}
}
