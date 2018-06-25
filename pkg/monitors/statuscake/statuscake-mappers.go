package statuscake

import (
	"strconv"

	"github.com/stakater/IngressMonitorController/pkg/models"
)

//StatusCakeMonitorMonitorToBaseMonitorMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeMonitor StatusCakeMonitorMonitor) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeMonitor.WebsiteName
	m.URL = statuscakeMonitor.WebsiteURL
	m.ID = strconv.Itoa(statuscakeMonitor.TestID)
	return &m
}

//StatusCakeMonitorMonitorsToBaseMonitorsMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorsToBaseMonitorsMapper(statuscakeMonitors []StatusCakeMonitorMonitor) []models.Monitor {
	var monitors []models.Monitor
	for index := 0; index < len(statuscakeMonitors); index++ {
		monitors = append(monitors, *StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeMonitors[index]))
	}
	return monitors
}
