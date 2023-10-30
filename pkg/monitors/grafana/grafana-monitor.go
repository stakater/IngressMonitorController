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
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

var log = logf.Log.WithName("gcloud-monitor")

const (
	// Default value for monitor configuration
	FrequencyDefaultValue = 10000
)

type GrafanaMonitorService struct {
	apiKey    string
	baseURL   string
	client    http.Client
	ctx       context.Context
	smClient  *smapi.Client                // Synthetic Monitoring client
	tenant    *synthetic_monitoring.Tenant // Tenant ID for Synthetic Monitoring
	frequency int64
}

func (service *GrafanaMonitorService) Setup(provider config.Provider) {
	service.ctx = context.Background()
	service.apiKey = provider.ApiKey
	service.client = http.Client{}
	service.baseURL = provider.ApiURL
	client := smapi.NewClient(service.baseURL, service.apiKey, http.DefaultClient)
	tenant, err := client.GetTenant(service.ctx)
	if err != nil {
		log.Error(err, "Cannot get tennant")
		return
	}
	if err != nil {
		log.Error(err, "Failed to initialize Synthetic Monitoring client")
		return
	}
	service.smClient = client
	service.tenant = tenant
	//CHECK if freq is set
	if provider.GrafanaConfig.Frequency > 0 {
		service.frequency = provider.GrafanaConfig.Frequency
	} else {
		service.frequency = FrequencyDefaultValue
	}
}

func (service *GrafanaMonitorService) GetAll() (monitors []models.Monitor) {
	// Using the synthetic monitoring library to list all checks
	checks, err := service.smClient.ListChecks(service.ctx)
	if err != nil {
		log.Error(err, "Error getting monitors")
		return nil
	}

	for _, check := range checks {
		monitors = append(monitors, models.Monitor{
			Name: check.Job,
			URL:  check.Target,
			ID:   fmt.Sprintf("%v", check.Id),
			Config: &endpointmonitorv1alpha1.GrafanaConfig{
				TenantId: check.TenantId,
			},
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

	var checkId int64
	if len(monitor.ID) > 0 {
		idResult, err := strconv.ParseInt(monitor.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error converting ID %v %v", monitor.ID, err)
		}
		checkId = idResult
	}
	var tentantId int64
	fmt.Fprintln(os.Stderr, "What is config before", monitor.Config)
	grafanaConfig, _ := monitor.Config.(*endpointmonitorv1alpha1.GrafanaConfig)
	fmt.Fprintln(os.Stderr, "What is config after", grafanaConfig)
	if grafanaConfig != nil && &grafanaConfig.TenantId != nil {
		tentantId = grafanaConfig.TenantId
		fmt.Fprintln(os.Stderr, "Sets tenantId", tentantId)
	}
	// Creating a new Check object
	return &synthetic_monitoring.Check{
		Id:        checkId,
		Target:    monitor.URL,
		Job:       monitor.Name,
		Frequency: service.frequency,
		TenantId:  tentantId,
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
		log.Error(err, "Failed to create synthetic check")
		fmt.Fprintln(os.Stderr, "Failed create", err)
		return
	}

	// Using the synthetic monitoring client to add the new check
	fmt.Fprintln(os.Stderr, "Result add check", newCheck)
	createdCheck, err := service.smClient.AddCheck(service.ctx, *newCheck)
	if err != nil {
		log.Error(err, "Failed to add new monitor")
		return
	}

	log.Info(fmt.Sprintf("Successfully added new monitor %v %v", monitor.ID, createdCheck.Id))
}

func (service *GrafanaMonitorService) Update(monitor models.Monitor) {
	newCheck, err := service.CreateSyntheticCheck(monitor)
	if err != nil {
		log.Error(err, "Failed to create synthetic check")
		return
	}
	fmt.Fprintln(os.Stderr, "Result update check", newCheck)
	// Using the synthetic monitoring client to update the old check
	createdCheck, err := service.smClient.UpdateCheck(service.ctx, *newCheck)
	if err != nil {
		log.Error(err, "Failed to update monitor")
		fmt.Fprintln(os.Stderr, "Failed to update monitor", err)
		return
	}

	log.Info(fmt.Sprintf("Successfully updated monitor %v %v", monitor.ID, createdCheck.Id))
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
		log.Info("Failed to parse int", monitor.ID)
		return
	}
	service.smClient.DeleteCheck(service.ctx, Id)
}

func (service *GrafanaMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO Implement Deep equal for config as well
	return oldMonitor.Name == newMonitor.Name && oldMonitor.URL == oldMonitor.URL && oldMonitor.ID == newMonitor.ID
}
