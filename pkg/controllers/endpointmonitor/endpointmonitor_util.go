package endpointmonitor

import (
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"
)

func findMonitorByName(monitorService monitors.MonitorServiceProxy, monitorName string) *models.Monitor {

	monitor, _ := monitorService.GetByName(monitorName)
	// Monitor Exists
	if monitor != nil {
		return monitor
	}
	return nil
}
