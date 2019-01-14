package uptimerobot

import (
	"github.com/stakater/IngressMonitorController/pkg/util"
	"strings"
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
	service.Setup(config.Providers[0])

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
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

	mRes.URL = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The URL should have been updated, expected: https://facebook.com, but was: " + mRes.URL)
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithIntervalAnnotations(t *testing.T) {
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
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}
	if "600" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("The interval is incorrect, expected: 600, but was: " + mRes.Annotations["uptimerobot.monitor.stakater.com/interval"])
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorIntervalAnnotations(t *testing.T) {
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
	if mRes.Name != m.Name {
		t.Error("The initial name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The initial URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}
	if "600" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("The initial interval is incorrect: 600, but was: " + mRes.Annotations["uptimerobot.monitor.stakater.com/interval"])
	}

	mRes.URL = "https://facebook.com"
    annotations["uptimerobot.monitor.stakater.com/interval"] = "900"
	mRes.Annotations = annotations

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The updated URL is incorrect, expected: https://facebook.com, but was: " + mRes.URL)
	}
	if "900" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("The updated interval is incorrect, expected: 900, but was: " + mRes.Annotations["uptimerobot.monitor.stakater.com/interval"])
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithStatusPageAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

	statusPageService := UpTimeStatusPageService{}
	statusPageService.Setup(config.Providers[0])

	statusPage := UpTimeStatusPage{Name: "status-page-test"}
	ID, err :=  statusPageService.Add(statusPage)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage.ID = ID

	var annotations = map[string]string {
		"uptimerobot.monitor.stakater.com/status-pages": statusPage.ID,
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
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
	statusPageRes, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if ! util.ContainsString(statusPageRes.Monitors, mRes.ID) {
		t.Error("The status page does not contain the monitor, expected: " + mRes.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}
	service.Remove(*mRes)
	statusPageService.Remove(statusPage)
}

func TestUpdateMonitorIntervalStatusPageAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	service.Setup(config.Providers[0])

	statusPageService := UpTimeStatusPageService{}
	statusPageService.Setup(config.Providers[0])

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
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

	statusPage := UpTimeStatusPage{Name: "status-page-test"}
	ID, err :=  statusPageService.Add(statusPage)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage.ID = ID

	var annotations = map[string]string {
		"uptimerobot.monitor.stakater.com/status-pages": statusPage.ID,
	}

	mRes.URL = "https://facebook.com"
	mRes.Annotations = annotations

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The updated URL is incorrect, expected: https://facebook.com, but was: " + mRes.URL)
	}
	statusPageRes, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if ! util.ContainsString(statusPageRes.Monitors, mRes.ID) {
		t.Error("The status page does not contain the monitor, expected: " + mRes.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}

	service.Remove(*mRes)
	statusPageService.Remove(statusPage)
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
