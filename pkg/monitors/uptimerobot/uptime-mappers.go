package uptimerobot

import (
	"strconv"
	"strings"

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

	alertContacts := make([]string,0)
	if uptimeMonitor.AlertContacts != nil {
		for _, alertContact := range uptimeMonitor.AlertContacts {
			contact := alertContact.ID + "_" + strconv.Itoa(alertContact.threshold) + "_" + strconv.Itoa(alertContact.recurrence)
			alertContacts = append(alertContacts, contact)
		}
		annotations["uptimerobot.monitor.stakater.com/alert-contacts"] = strings.Join(alertContacts,"-")
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
