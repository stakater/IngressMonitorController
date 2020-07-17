package uptimerobot

import (
	"strconv"
	"strings"

	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
)

func UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitor UptimeMonitorMonitor) *models.Monitor {
	var m models.Monitor

	m.Name = uptimeMonitor.FriendlyName
	m.URL = uptimeMonitor.URL
	m.ID = strconv.Itoa(uptimeMonitor.ID)

	var providerConfig ingressmonitorv1alpha1.UptimeRobotConfig
	providerConfig.Interval = uptimeMonitor.Interval

	alertContacts := make([]string, 0)
	if uptimeMonitor.AlertContacts != nil {
		for _, alertContact := range uptimeMonitor.AlertContacts {
			contact := alertContact.ID + "_" + strconv.Itoa(alertContact.Threshold) + "_" + strconv.Itoa(alertContact.Recurrence)
			alertContacts = append(alertContacts, contact)
		}
		providerConfig.AlertContacts = strings.Join(alertContacts, "-")
	}

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

func UptimeStatusPageToBaseStatusPageMapper(uptimePublicStatusPage UptimePublicStatusPage) *UpTimeStatusPage {
	var s UpTimeStatusPage

	s.Name = uptimePublicStatusPage.FriendlyName
	s.Monitors = util.SliceItoa(uptimePublicStatusPage.Monitors)
	s.ID = strconv.Itoa(uptimePublicStatusPage.ID)

	return &s
}
