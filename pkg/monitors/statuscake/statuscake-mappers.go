package statuscake

import (
	"strings"

	statuscake "github.com/StatusCakeDev/statuscake-go"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

// StatusCakeMonitorMonitorToBaseMonitorMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeData StatusCakeMonitorData) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.WebsiteName
	m.URL = statuscakeData.WebsiteURL
	m.ID = statuscakeData.TestID

	var providerConfig endpointmonitorv1alpha1.StatusCakeConfig
	providerConfig.TestTags = strings.Join(statuscakeData.Tags, ",")
	m.Config = &providerConfig
	return &m
}

// StatusCakeApiResponseDataToBaseMonitorMapper function to map Statuscake Uptime Test Response to Monitor
func StatusCakeApiResponseDataToBaseMonitorMapper(statuscakeData statuscake.UptimeTestResponse) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.Data.Name
	m.URL = statuscakeData.Data.WebsiteURL
	m.ID = statuscakeData.Data.ID

	var providerConfig endpointmonitorv1alpha1.StatusCakeConfig
	providerConfig.TestTags = strings.Join(statuscakeData.Data.Tags, ",")
	m.Config = &providerConfig
	return &m
}

// StatusCakeMonitorMonitorsToBaseMonitorsMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorsToBaseMonitorsMapper(statuscakeData []StatusCakeMonitorData) []models.Monitor {
	var monitors []models.Monitor
	for _, payloadData := range statuscakeData {
		monitors = append(monitors, *StatusCakeMonitorMonitorToBaseMonitorMapper(payloadData))
	}
	return monitors
}

// StatusCakeHeartbeatToBaseMonitorMapper maps a StatusCake HeartbeatTestOverview to a Monitor.
// m.URL is the StatusCake-generated push endpoint that clients ping to register a heartbeat.
func StatusCakeHeartbeatToBaseMonitorMapper(hb statuscake.HeartbeatTestOverview) *models.Monitor {
	var m models.Monitor
	m.Name = hb.Name
	m.URL = hb.WebsiteURL
	m.ID = hb.ID

	var providerConfig endpointmonitorv1alpha1.StatusCakeConfig
	providerConfig.TestType = "Heartbeat"
	providerConfig.TestTags = strings.Join(hb.Tags, ",")
	providerConfig.CheckRate = int(hb.Period)
	providerConfig.Paused = hb.Paused
	providerConfig.ContactGroup = strings.Join(hb.ContactGroups, ",")
	m.Config = &providerConfig
	return &m
}

// StatusCakeHeartbeatsToBaseMonitorsMapper maps a slice of HeartbeatTestOverview to Monitors
func StatusCakeHeartbeatsToBaseMonitorsMapper(hbs []statuscake.HeartbeatTestOverview) []models.Monitor {
	var monitors []models.Monitor
	for _, hb := range hbs {
		monitors = append(monitors, *StatusCakeHeartbeatToBaseMonitorMapper(hb))
	}
	return monitors
}
