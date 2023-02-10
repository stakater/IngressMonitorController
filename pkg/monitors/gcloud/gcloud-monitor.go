package gcloud

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	monitoredres "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

var log = logf.Log.WithName("gcloud-monitor")

type MonitorService struct {
	client    *monitoring.UptimeCheckClient
	projectID string
	ctx       context.Context
}

func (monitor *MonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO: Retrieve oldMonitor config and compare it here
	return false
}

func (service *MonitorService) Setup(provider config.Provider) {
	service.ctx = context.Background()

	client, err := monitoring.NewUptimeCheckClient(service.ctx, option.WithCredentialsJSON([]byte(provider.ApiKey)))

	if err != nil {
		log.Info("Error Seting Up Monitor Service: " + err.Error())
	} else {
		service.client = client
	}
	service.projectID = provider.GcloudConfig.ProjectID
}

func (service *MonitorService) GetByName(name string) (monitor *models.Monitor, err error) {
	uptimeCheckConfigsIterator := service.client.ListUptimeCheckConfigs(service.ctx, &monitoringpb.ListUptimeCheckConfigsRequest{
		Parent: "projects/" + service.projectID,
	})

	for {
		uptimeCheckConfig, err := uptimeCheckConfigsIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error Locating Monitor: %s", err.Error())
		}
		if uptimeCheckConfig.DisplayName == name {
			localMonitor := transformToMonitor(uptimeCheckConfig)
			return &localMonitor, nil
		}
	}

	return nil, fmt.Errorf("Unable to locate monitor with name %v", name)
}

func (service *MonitorService) GetAll() (monitors []models.Monitor) {
	uptimeCheckConfigsIterator := service.client.ListUptimeCheckConfigs(service.ctx, &monitoringpb.ListUptimeCheckConfigsRequest{
		Parent: "projects/" + service.projectID,
	})

	monitors = []models.Monitor{}
	for {
		uptimeCheckConfig, err := uptimeCheckConfigsIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Info("Error received while listing checks: " + err.Error())
			return nil
		}
		monitors = append(monitors, transformToMonitor(uptimeCheckConfig))
	}

	return monitors
}

func (service *MonitorService) Add(monitor models.Monitor) {
	url, err := url.Parse(monitor.URL)
	if err != nil {
		log.Info("Error Adding Monitor: " + err.Error())
		return
	}

	portString := url.Port()
	var port int
	if portString == "" {
		if url.Scheme == "http" {
			port = 80
		} else if url.Scheme == "https" {
			port = 443
		} else {
			log.Info("Error Adding Monitor: unknown protocol " + url.Scheme)
			return
		}
	} else {
		port, err = strconv.Atoi(portString)
		if err != nil {
			log.Info("Error Adding Monitor: " + err.Error())
			return
		}
	}

	projectID := service.projectID
	providerConfig, _ := monitor.Config.(*endpointmonitorv1alpha1.GCloudConfig)
	if providerConfig != nil && len(providerConfig.ProjectId) != 0 {
		projectID = providerConfig.ProjectId
	}

	_, err = service.client.CreateUptimeCheckConfig(service.ctx, &monitoringpb.CreateUptimeCheckConfigRequest{
		Parent: "projects/" + projectID,
		UptimeCheckConfig: &monitoringpb.UptimeCheckConfig{
			DisplayName: monitor.Name,
			Resource: &monitoringpb.UptimeCheckConfig_MonitoredResource{
				MonitoredResource: &monitoredres.MonitoredResource{
					Type: "uptime_url",
					Labels: map[string]string{
						"host": url.Hostname(),
					},
				},
			},
			CheckRequestType: &monitoringpb.UptimeCheckConfig_HttpCheck_{
				HttpCheck: &monitoringpb.UptimeCheckConfig_HttpCheck{
					Path:   url.Path,
					Port:   int32(port),
					UseSsl: url.Scheme == "https",
				},
			},
		},
	})
	if err != nil {
		log.Info("Error Adding Monitor: " + err.Error())
		return
	}

	log.Info("Added monitor for: " + monitor.Name)
}

func (service *MonitorService) Update(monitor models.Monitor) {
	uptimeCheckConfig, err := service.client.GetUptimeCheckConfig(service.ctx, &monitoringpb.GetUptimeCheckConfigRequest{Name: monitor.ID})
	if err != nil {
		log.Info("Error updating Monitor: " + err.Error())
	}

	url, err := url.Parse(monitor.URL)
	if err != nil {
		log.Info("Error Adding Monitor: " + err.Error())
		return
	}

	if uptimeCheckConfig.GetMonitoredResource().Labels["host"] != url.Hostname() {
		log.Info("Error Adding Monitor: URL Host is immutable")
		return
	}

	portString := url.Port()
	var port int
	if portString == "" {
		if url.Scheme == "http" {
			port = 80
		} else if url.Scheme == "https" {
			port = 443
		} else {
			log.Info("Error Adding Monitor: unknown protocol " + url.Scheme)
			return
		}
	} else {
		port, err = strconv.Atoi(portString)
		if err != nil {
			log.Info("Error Adding Monitor: " + err.Error())
			return
		}
	}

	uptimeCheckConfig.DisplayName = monitor.Name
	uptimeCheckConfig.GetHttpCheck().Port = int32(port)
	uptimeCheckConfig.GetHttpCheck().Path = url.Path

	uptimeCheckConfig, err = service.client.UpdateUptimeCheckConfig(service.ctx, &monitoringpb.UpdateUptimeCheckConfigRequest{
		UptimeCheckConfig: uptimeCheckConfig,
	})
	if err != nil {
		log.Info("Error Adding Monitor: " + err.Error())
		return
	}

	log.Info(fmt.Sprintf("Updated Monitor: %v", uptimeCheckConfig))
}

func (service *MonitorService) Remove(monitor models.Monitor) {
	err := service.client.DeleteUptimeCheckConfig(service.ctx, &monitoringpb.DeleteUptimeCheckConfigRequest{
		Name: monitor.ID,
	})
	if err != nil {
		log.Info("Error deleting Monitor: " + err.Error())
		return
	}
	log.Info("Deleted Monitor: " + monitor.Name)
}

func transformToMonitor(uptimeCheckConfig *monitoringpb.UptimeCheckConfig) (monitor models.Monitor) {
	isSsl := uptimeCheckConfig.GetHttpCheck().UseSsl
	path := uptimeCheckConfig.GetHttpCheck().Path
	port := uptimeCheckConfig.GetHttpCheck().Port
	host := uptimeCheckConfig.GetMonitoredResource().Labels["host"]

	var scheme string
	if isSsl {
		scheme = "https"
	} else {
		scheme = "http"
	}

	if (isSsl && port != 443) || (!isSsl && port != 80) {
		host = host + ":" + strconv.FormatInt(int64(port), 10)
	}

	url := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	return models.Monitor{
		URL:  url.String(),
		Name: uptimeCheckConfig.DisplayName,
		ID:   uptimeCheckConfig.Name,
	}
}
