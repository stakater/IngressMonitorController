package statuscake

import (
	statuscake "github.com/StatusCakeDev/statuscake-go"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

// StatusCakeMonitorMonitorToBaseMonitorMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeData StatusCakeMonitorData) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.WebsiteName
	m.URL = statuscakeData.WebsiteURL
	m.ID = statuscakeData.TestID
	return &m
}

// StatusCakeApiResponseDataToBaseMonitorMapper function to map Statuscake Uptime Test Response to Monitor
func StatusCakeApiResponseDataToBaseMonitorMapper(statuscakeData statuscake.UptimeTestResponse) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.Data.Name
	m.URL = statuscakeData.Data.WebsiteURL
	m.ID = statuscakeData.Data.ID
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
