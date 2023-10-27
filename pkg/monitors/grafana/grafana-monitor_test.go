package grafana

import (
	"testing"
	"time"

	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
)

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()
	t.Errorf("Insta fail")

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
	if mRes[0].Name != m.Name || mRes[0].URL != m.URL {
		t.Error("URL and name should be the same", mRes[0], m)
	}
	service.Remove(mRes[0])

	time.Sleep(5 * time.Second)

	monitor, err := service.GetByName(mRes[0].Name)

	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}
