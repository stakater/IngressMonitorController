package main

import "strconv"

//StatusCakeMonitorMonitorToBaseMonitorMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeMonitor StatusCakeMonitorMonitor) *Monitor {
	var m Monitor
	m.name = statuscakeMonitor.WebsiteName
	m.url = statuscakeMonitor.WebsiteURL
	m.id = strconv.Itoa(statuscakeMonitor.TestID)
	return &m
}

//StatusCakeMonitorMonitorsToBaseMonitorsMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorsToBaseMonitorsMapper(statuscakeMonitors []StatusCakeMonitorMonitor) []Monitor {
	var monitors []Monitor
	for index := 0; index < len(statuscakeMonitors); index++ {
		monitors = append(monitors, *StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeMonitors[index]))
	}
	return monitors
}
