package uptime

import (
	"strconv"
	"strings"

	"github.com/stakater/IngressMonitorController/pkg/models"
)

func UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitor UptimeMonitorMonitor) *models.Monitor {
	var m models.Monitor

	m.Name = uptimeMonitor.Name
	m.URL = uptimeMonitor.MspAddress
	m.ID = strconv.Itoa(uptimeMonitor.PK)

	var annotations = map[string]string{
		"uptime.monitor.stakater.com/interval":   strconv.Itoa(uptimeMonitor.MspInterval),
		"uptime.monitor.stakater.com/check_type": uptimeMonitor.CheckType,
		"uptime.monitor.stakater.com/contacts":   strings.Join(uptimeMonitor.ContactGroups, ","),
		"uptime.monitor.stakater.com/locations":  strings.Join(uptimeMonitor.Locations, ","),
	}

	m.Annotations = annotations
	return &m
}

func UptimeMonitorMonitorsToBaseMonitorsMapper(uptimeMonitors []UptimeMonitorMonitor) []models.Monitor {
	var monitors []models.Monitor

	for index := 0; index < len(uptimeMonitors); index++ {
		monitors = append(monitors, *UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitors[index]))
	}

	return monitors
}
