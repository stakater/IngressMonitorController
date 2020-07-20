package gcloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	monitoredres "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

const (
	ProjectIDAnnotation = "gcloud.monitor.stakater.com/project-id"
)

type MonitorService struct {
	client    *monitoring.UptimeCheckClient
	projectID string
	ctx       context.Context
}

func (service *MonitorService) Setup(provider config.Provider) {
	service.ctx = context.Background()

	client, err := monitoring.NewUptimeCheckClient(service.ctx, option.WithCredentialsJSON([]byte(provider.ApiKey)))

	if err != nil {
		log.Println("Error Seting Up Monitor Service: ", err.Error())
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
			log.Println("Error received while listing checks: ", err.Error())
			return nil
		}
		monitors = append(monitors, transformToMonitor(uptimeCheckConfig))
	}

	return monitors
}

func (service *MonitorService) Add(monitor models.Monitor) {
	url, err := url.Parse(monitor.URL)
	if err != nil {
		log.Println("Error Adding Monitor: ", err.Error())
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
			log.Println("Error Adding Monitor: unknown protocol ", url.Scheme)
			return
		}
	} else {
		port, err = strconv.Atoi(portString)
		if err != nil {
			log.Println("Error Adding Monitor: ", err.Error())
			return
		}
	}

	_, err = service.client.CreateUptimeCheckConfig(service.ctx, &monitoringpb.CreateUptimeCheckConfigRequest{
		Parent: "projects/" + service.projectID,
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
		log.Println("Error Adding Monitor: ", err.Error())
		return
	}

	log.Println("Added monitor for: ", monitor.Name)
}

func (service *MonitorService) Update(monitor models.Monitor) {
	uptimeCheckConfig, err := service.client.GetUptimeCheckConfig(service.ctx, &monitoringpb.GetUptimeCheckConfigRequest{Name: monitor.ID})
	if err != nil {
		log.Println("Error updating Monitor: ", err.Error())
	}

	url, err := url.Parse(monitor.URL)
	if err != nil {
		log.Println("Error Adding Monitor: ", err.Error())
		return
	}

	if uptimeCheckConfig.GetMonitoredResource().Labels["host"] != url.Hostname() {
		log.Println("Error Adding Monitor: URL Host is immutable")
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
			log.Println("Error Adding Monitor: unknown protocol ", url.Scheme)
			return
		}
	} else {
		port, err = strconv.Atoi(portString)
		if err != nil {
			log.Println("Error Adding Monitor: ", err.Error())
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
		log.Println("Error Adding Monitor: ", err.Error())
		return
	}

	log.Println("Updated Monitor: ", uptimeCheckConfig)
}

func (service *MonitorService) Remove(monitor models.Monitor) {
	err := service.client.DeleteUptimeCheckConfig(service.ctx, &monitoringpb.DeleteUptimeCheckConfigRequest{
		Name: monitor.ID,
	})
	if err != nil {
		log.Println("Error deleting Monitor: ", err.Error())
		return
	}
	log.Println("Deleted Monitor: ", monitor.Name)
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
		URL:         url.String(),
		Name:        uptimeCheckConfig.DisplayName,
		ID:          uptimeCheckConfig.Name,
		Annotations: map[string]string{},
	}
}
