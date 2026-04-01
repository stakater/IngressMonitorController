package controllers

import (
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"
)

func findMonitorByName(monitorService *monitors.MonitorServiceProxy, monitorName string) (*models.Monitor, error) {
	return monitorService.GetByName(monitorName)
}

// findMonitorServicesThatContainsMonitor iterates over all monitor services and returns the ones that contain the monitor
func findMonitorServicesThatContainsMonitor(monitorServices []*monitors.MonitorServiceProxy, monitorName string) ([]*monitors.MonitorServiceProxy, error) {
	var targetMonitorServices []*monitors.MonitorServiceProxy
	for _, monitorService := range monitorServices {
		monitor, err := monitorService.GetByName(monitorName)
		if err != nil {
			return nil, err
		}
		if monitor != nil {
			targetMonitorServices = append(targetMonitorServices, monitorService)
		}
	}
	return targetMonitorServices, nil
}
