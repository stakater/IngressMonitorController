package controllers

import (
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"
)

func findMonitorByName(monitorService *monitors.MonitorServiceProxy, monitorName string) *models.Monitor {

	monitor, _ := monitorService.GetByName(monitorName)
	// Monitor Exists
	if monitor != nil {
		return monitor
	}
	return nil
}

// findMonitorServiceThatContainsMonitor iterates over all monitor services and returns the one that contains the monitor
func findMonitorServicesThatContainsMonitor(monitorServices []*monitors.MonitorServiceProxy, monitorName string) []*monitors.MonitorServiceProxy {
	var targetMonitorServices []*monitors.MonitorServiceProxy
	for _, monitorService := range monitorServices {
		monitor, _ := monitorService.GetByName(monitorName)
		if monitor != nil {
			targetMonitorServices = append(targetMonitorServices, monitorService)
		}
	}
	return targetMonitorServices
}
