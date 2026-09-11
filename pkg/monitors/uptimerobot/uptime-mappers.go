package uptimerobot

import (
	"strconv"
	"strings"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
)

func UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitor UptimeMonitorMonitor) *models.Monitor {
	var m models.Monitor

	m.Name = uptimeMonitor.FriendlyName
	m.URL = uptimeMonitor.URL
	m.ID = strconv.Itoa(uptimeMonitor.ID)

	var providerConfig endpointmonitorv1alpha1.UptimeRobotConfig
	providerConfig.Interval = uptimeMonitor.Interval

	// Rebuild the v2-style "id_threshold_recurrence-..." string so Equal()
	// keeps comparing monitors created from CRDs of the same format
	alertContacts := make([]string, 0, len(uptimeMonitor.AssignedAlertContacts))
	for _, alertContact := range uptimeMonitor.AssignedAlertContacts {
		contact := strconv.Itoa(alertContact.AlertContactId) + "_" + strconv.Itoa(alertContact.Threshold) + "_" + strconv.Itoa(alertContact.Recurrence)
		alertContacts = append(alertContacts, contact)
	}
	providerConfig.AlertContacts = strings.Join(alertContacts, "-")

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

func UptimeStatusPageToBaseStatusPageMapper(uptimePublicStatusPage UptimePublicStatusPage) *UpTimeStatusPage {
	var s UpTimeStatusPage

	s.Name = uptimePublicStatusPage.FriendlyName
	s.Monitors = util.SliceItoa(uptimePublicStatusPage.MonitorIds)
	s.ID = strconv.Itoa(uptimePublicStatusPage.ID)

	return &s
}
