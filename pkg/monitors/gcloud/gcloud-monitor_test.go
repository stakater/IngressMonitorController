package gcloud

import (
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := MonitorService{}
	provider := util.GetProviderWithName(config, "gcloud")
	if provider == nil {
		return
	}

	service.Setup(*provider)
	m := models.Monitor{Name: "google-test", URL: "https://google1.com/"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Error("URL and name should be the same")
	}

	var inList = false
	for _, monitorInList := range service.GetAll() {
		if monitorInList.Name != m.Name {
			continue
		}
		inList = true
	}
	if !inList {
		t.Error("Monitor should've been in list ")
	}

	service.Remove(*mRes)
	monitor, err := service.GetByName(mRes.Name)

	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := MonitorService{}

	provider := util.GetProviderWithName(config, "gcloud")
	if provider == nil {
		return
	}
	service.Setup(*provider)

	m := models.Monitor{Name: "google-test", URL: "https://google.com/"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Error("URL and name should be the same")
	}

	mRes.Name = "google-test2"
	mRes.URL = "https://google.com/test"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test2")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://google.com/test" {
		t.Error("URL and name should be the same")
	}

	service.Remove(*mRes)

	monitor, err := service.GetByName(mRes.Name)

	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}
