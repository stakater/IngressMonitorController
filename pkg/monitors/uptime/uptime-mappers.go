package uptime

import (
	"strconv"
	"strings"

	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

func UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitor UptimeMonitorMonitor) *models.Monitor {
	var m models.Monitor

	m.Name = uptimeMonitor.Name
	m.URL = uptimeMonitor.MspAddress
	m.ID = strconv.Itoa(uptimeMonitor.PK)

	var providerConfig ingressmonitorv1alpha1.UptimeConfig
	providerConfig.Interval = uptimeMonitor.MspInterval
	providerConfig.CheckType = uptimeMonitor.CheckType
	providerConfig.Contacts = strings.Join(uptimeMonitor.ContactGroups, ",")
	providerConfig.Locations = strings.Join(uptimeMonitor.Locations, ",")

	m.Config = providerConfig
	return &m
}

func UptimeMonitorMonitorsToBaseMonitorsMapper(uptimeMonitors []UptimeMonitorMonitor) []models.Monitor {
	var monitors []models.Monitor

	for index := 0; index < len(uptimeMonitors); index++ {
		monitors = append(monitors, *UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitors[index]))
	}

	return monitors
}
