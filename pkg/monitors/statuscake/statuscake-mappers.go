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
