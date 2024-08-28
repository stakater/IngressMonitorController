package grafana

import (
	"reflect"
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// To allow normal logging to be printed if tests fails
	// Dev mode is an extra feature to make output more readable
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := GrafanaMonitorService{}
	provider := util.GetProviderWithName(config, "Grafana")
	if provider == nil {
		return
	}
	service.Setup(*provider)

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: &endpointmonitorv1alpha1.GrafanaConfig{
		Frequency:        20000,
		Probes:           []string{"Singapore"},
		AlertSensitivity: "low",
	}}

	preExistingMonitor, _ := service.GetByName(m.Name)

	if preExistingMonitor != nil {
		service.Remove(*preExistingMonitor)
	}

	previousResources := len(service.GetAll())
	service.Add(m)

	mRes := service.GetAll()

	if len(mRes) == previousResources {
		t.Errorf("Found empty response for Monitor. Name: %s and URL: %s", m.Name, m.URL)
	}
	if len(mRes) > previousResources+1 {
		t.Errorf("Found too many response for Monitor, %v, after add.", len(mRes))
	}

	monitor, err := service.GetByName(m.Name)

	if err != nil {
		t.Error("Monitor should've been found", monitor, err)
	}
	monitorConfig, _ := monitor.Config.(*endpointmonitorv1alpha1.GrafanaConfig)
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.GrafanaConfig)

	if monitor.Name != m.Name || monitor.URL != m.URL || monitorConfig.Frequency != providerConfig.Frequency || reflect.DeepEqual(monitorConfig.Probes, providerConfig.Probes) || monitorConfig.AlertSensitivity != providerConfig.AlertSensitivity {
		t.Error("URL, name, frequency, probes and alertSensitivity should be the same", monitor, m)
	}
	service.Remove(*monitor)

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
	preExistingMonitor, _ := service.GetByName(m.Name)

	if preExistingMonitor != nil {
		service.Remove(*preExistingMonitor)
	}

	previousResources := len(service.GetAll())
	service.Add(m)

	mRes := service.GetAll()

	if len(mRes) == previousResources {
		t.Errorf("Found empty response for Monitor. Name: %s and URL: %s", m.Name, m.URL)
	}
	if len(mRes) > previousResources+1 {
		t.Errorf("Found too many response for Monitor, %v, after add.", len(mRes))
	}
	monitor, err := service.GetByName(m.Name)
	if err != nil || monitor == nil {
		t.Error("Monitor should've been found", monitor, err)
	}
	if monitor.Name != m.Name || monitor.URL != m.URL {
		t.Error("URL and name should be the same", monitor, m)
	}
	m2 := models.Monitor{Name: "stakater-test", URL: "https://stakater.com", ID: monitor.ID, Config: monitor.Config}
	service.Update(m2)

	mRes2 := service.GetAll()

	if len(mRes2) == previousResources {
		t.Errorf("Found empty response for Monitor. Name: %s and URL: %s", m2.Name, m2.URL)
	}
	if len(mRes2) > previousResources+1 {
		t.Errorf("Found too many response for Monitor, %v, after update.", len(mRes2))
	}

	monitor1, _ := service.GetByName(m.Name)
	if monitor1 != nil {
		t.Error("Monitor should not exist since it was updated", monitor, err)
	}
	monitor2, err := service.GetByName(m2.Name)
	if err != nil {
		t.Error("Monitor should've been found", monitor, err)
	}
	if monitor2.Name != m2.Name || monitor2.URL != m2.URL {
		t.Error("URL and name should be the same", monitor2, m2)
	}
	service.Remove(*monitor2)

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
