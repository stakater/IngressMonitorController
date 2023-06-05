package controllers

import (
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"
)

func findMonitorByName(monitorService monitors.MonitorServiceProxy, monitorName string) (*models.Monitor, error) {

	monitor, err := monitorService.GetByName(monitorName)
	return monitor, err
}
