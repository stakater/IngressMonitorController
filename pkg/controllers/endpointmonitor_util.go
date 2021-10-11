package controllers

import (
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"
)

func findMonitorByName(monitorService monitors.MonitorServiceProxy, monitorName string) (*models.Monitor, error) {

	monitor, err := monitorService.GetByName(monitorName)
	return monitor, err
}
