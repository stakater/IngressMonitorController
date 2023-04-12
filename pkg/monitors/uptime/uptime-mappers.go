package uptime

import (
	"strconv"
	"strings"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

func UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitor UptimeMonitorMonitor) *models.Monitor {
	var m models.Monitor

	m.Name = uptimeMonitor.Name
	m.URL = uptimeMonitor.MspAddress
	m.ID = strconv.Itoa(uptimeMonitor.PK)

	var providerConfig endpointmonitorv1alpha1.UptimeConfig
	providerConfig.Interval = uptimeMonitor.MspInterval
	providerConfig.CheckType = uptimeMonitor.CheckType
	providerConfig.Contacts = strings.Join(uptimeMonitor.ContactGroups, ",")
	providerConfig.Locations = strings.Join(uptimeMonitor.Locations, ",")
	providerConfig.Tags = strings.Join(uptimeMonitor.Tags, ",")
	m.Config = &providerConfig
	return &m
}

func UptimeMonitorMonitorsToBaseMonitorsMapper(uptimeMonitors []UptimeMonitorMonitor) []models.Monitor {
	var monitors []models.Monitor

	for index := 0; index < len(uptimeMonitors); index++ {
		monitors = append(monitors, *UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitors[index]))
	}

	return monitors
}
