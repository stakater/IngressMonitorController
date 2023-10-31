package grafana

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
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

	monitor, err := service.GetByName(m.Name)

	if err != nil {
		t.Error("Monitor should've been found", monitor, err)
	}
	service.Remove(mRes[0])

	monitor, err = service.GetByName(m.Name)

	if monitor != nil {
		t.Error("Cleanup of Monitor was unsuccessful", monitor, err)
	}
}
func TestUpdateMonitorWithCorrectValues(t *testing.T) {
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
		t.Error("Monitor should not exist since it was updated", monitor, err)
	}
	service.Remove(mRes2[0])

	monitor, err = service.GetByName(m2.Name)

	if monitor != nil {
		t.Error("Cleanup of Monitor was unsuccessful", monitor, err)
	}
}

func TestEqualModules(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := GrafanaMonitorService{}
	provider := util.GetProviderWithName(config, "Grafana")
	if provider == nil {
		return
	}
	service.Setup(*provider)
	m1 := models.Monitor{
		Name: "test",
		URL:  "https://google.com",
		ID:   "0",
		Config: &endpointmonitorv1alpha1.GrafanaConfig{
			TenantId: 5,
		},
	}
	m2 := models.Monitor{
		Name: "test",
		URL:  "https://google.com",
		ID:   "0",
		Config: &endpointmonitorv1alpha1.GrafanaConfig{
			TenantId: 5,
		},
	}
	m3 := models.Monitor{
		Name: "test",
		URL:  "https://google.com",
		ID:   "0",
		Config: &endpointmonitorv1alpha1.GrafanaConfig{
			TenantId: 2,
		},
	}
	m4 := models.Monitor{
		Name: "test",
		URL:  "https://google.com",
		Config: &endpointmonitorv1alpha1.GrafanaConfig{
			TenantId: 5,
		},
	}
	if !service.Equal(m1, m2) {
		t.Error("Monitor should be the same", m1, m2)
	}
	if service.Equal(m1, m3) {
		t.Error("Monitor should not be the same", m1, m3)
	}
	if service.Equal(m1, m4) {
		t.Error("Monitor should not be the same", m1, m4)
	}
}
