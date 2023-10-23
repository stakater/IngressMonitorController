package grafana

import (
	"context"
	"fmt"
	"net/http"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"

	"github.com/grafana/synthetic-monitoring-agent/pkg/pb/synthetic_monitoring"
	smapi "github.com/grafana/synthetic-monitoring-api-go-client"

	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

var log = logf.Log.WithName("gcloud-monitor")

type GrafanaMonitorService struct {
	apiKey   string
	baseURL  string
	client   http.Client
	ctx      context.Context
	smClient *smapi.Client                // Synthetic Monitoring client
	tenant   *synthetic_monitoring.Tenant // Tenant ID for Synthetic Monitoring
}

func (service *GrafanaMonitorService) Setup(provider config.Provider) {
	service.ctx = context.Background()
	service.apiKey = provider.ApiKey
	service.client = http.Client{}
	service.baseURL = provider.ApiURL
	client := smapi.NewClient(service.baseURL, service.apiKey, http.DefaultClient)
	tenant, err := client.GetTenant(service.ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot get tennant", err)
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to initialize Synthetic Monitoring client", err)
		return
	}
	service.smClient = client
	service.tenant = tenant
}

func (service *GrafanaMonitorService) GetAll() (monitors []models.Monitor) {
	// Using the synthetic monitoring library to list all checks
	checks, err := service.smClient.ListChecks(service.ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting monitors", err)
		return nil
	}

	fmt.Fprintln(os.Stderr, "Result of List checks", checks)
	for _, check := range checks {
		monitors = append(monitors, models.Monitor{
			Name: check.Job,
			URL:  check.Target,
			ID:   fmt.Sprintf("%v", check.Id),
		})
	}
	return monitors
}

func (service *GrafanaMonitorService) CreateSyntheticCheck(monitor models.Monitor) (*synthetic_monitoring.Check, error) {
	probes, err := service.smClient.ListProbes(service.ctx)
	if err != nil {
		return nil, fmt.Errorf("Error listing probes %v", err)
	}

	probeIDs := make([]int64, len(probes))
	for i, p := range probes {
		probeIDs[i] = p.Id
	}

	// Creating a new Check object
	return &synthetic_monitoring.Check{
		Target:    monitor.URL,
		Job:       monitor.Name,
		Frequency: 10000,
		Timeout:   2000,
		Enabled:   true,
		Probes:    probeIDs,
		Settings: synthetic_monitoring.CheckSettings{
			Http: &synthetic_monitoring.HttpSettings{
				IpVersion: synthetic_monitoring.IpVersion_V4,
			},
		},
		BasicMetricsOnly: true,
	}, nil
}

// Add adds a new monitor to Grafana Synthetic Monitoring service
func (service *GrafanaMonitorService) Add(monitor models.Monitor) {
	newCheck, err := service.CreateSyntheticCheck(monitor)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create synthetic check", err)
		return
	}

	// Using the synthetic monitoring client to add the new check
	createdCheck, err := service.smClient.AddCheck(service.ctx, *newCheck)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to add new monitor", err)
		return
	}

	fmt.Fprintln(os.Stderr, "Successfully added new monitor", "monitorID", createdCheck.Id)
}

func (service *GrafanaMonitorService) Update(monitor models.Monitor) {
	newCheck, err := service.CreateSyntheticCheck(monitor)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create synthetic check", err)
		return
	}

	// Using the synthetic monitoring client to update the old check
	createdCheck, err := service.smClient.UpdateCheck(service.ctx, *newCheck)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to add new monitor", err)
		return
	}

	fmt.Fprintln(os.Stderr, "Successfully updated monitor", "monitorID", createdCheck.Id)
}

func (service *GrafanaMonitorService) GetByName(name string) (*models.Monitor, error) {
	monitors := service.GetAll()
	for _, m := range monitors {
		if m.Name == name {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("Unable to locate monitor with name %v", name)
}

func (service *GrafanaMonitorService) Remove(monitor models.Monitor) {
	// Convert string to base64 int
	Id, err := strconv.ParseInt(monitor.ID, 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse int", monitor.ID)
		return
	}
	service.smClient.DeleteCheck(service.ctx, Id)
}

func (service *GrafanaMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO this is not good.
	mResOld, errOld := service.GetByName(oldMonitor.Name)
	mResNew, errNew := service.GetByName(oldMonitor.Name)
	if errNew != nil || errOld != nil {
		fmt.Fprintln(os.Stderr, "Failed to get monitors", errOld, errNew)
	}
	return mResOld.Name == mResNew.Name && mResOld.URL == mResNew.URL
}
