package uptimerobot

import (
	"strconv"

	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

func UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitor UptimeMonitorMonitor) *models.Monitor {
	var m models.Monitor

	m.Name = uptimeMonitor.FriendlyName
	m.URL = uptimeMonitor.URL
	m.ID = strconv.Itoa(uptimeMonitor.ID)

	var annotations = map[string]string{
		"uptimerobot.monitor.stakater.com/interval": strconv.Itoa(uptimeMonitor.Interval),
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

func UptimeStatusPageToBaseStatusPageMapper(uptimePublicStatusPage UptimePublicStatusPage) *UpTimeStatusPage {
	var s UpTimeStatusPage

	s.Name = uptimePublicStatusPage.FriendlyName
	s.Monitors = util.SliceItoa(uptimePublicStatusPage.Monitors)
	s.ID = strconv.Itoa(uptimePublicStatusPage.ID)

	return &s
}
