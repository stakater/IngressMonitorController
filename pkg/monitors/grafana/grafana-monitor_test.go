package grafana

import (
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	"testing"
)

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := GrafanaMonitorService{}
	provider := util.GetProviderWithName(config, "Grafana")
	if provider == nil {
		return
	}
	service.Setup(*provider)
	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes := service.GetAll()

	if len(mRes) == 0 {
		t.Errorf("Found empty response for Monitor. Name: %s and URL: %s", m.Name, m.URL)
	}
	if len(mRes) > 1 {
		t.Errorf("Found too many response for Monitor, %v, after add.", len(mRes))
	}
	if mRes[0].Name != m.Name || mRes[0].URL != m.URL {
		t.Error("URL and name should be the same", mRes[0], m)
	}
	m2 := models.Monitor{Name: "stakater-test", URL: "https://stakater.com", ID: mRes[0].ID, Config: mRes[0].Config}
	service.Update(m2)

	mRes2 := service.GetAll()

	if len(mRes2) == 0 {
		t.Errorf("Found empty response for Monitor. Name: %s and URL: %s", m2.Name, m2.URL)
	}
	if len(mRes2) > 1 {
		t.Errorf("Found too many response for Monitor, %v, after update.", len(mRes2))
	}
	if mRes2[0].Name != m2.Name || mRes2[0].URL != m2.URL {
		t.Error("URL and name should be the same", mRes2[0], m2)
	}

	monitor, err := service.GetByName(m.Name)

	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
	service.Remove(mRes2[0])

	monitor, err = service.GetByName(m2.Name)

	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}
