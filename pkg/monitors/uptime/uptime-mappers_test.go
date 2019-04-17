package uptime

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stakater/IngressMonitorController/pkg/models"
)

func TestUptimeMonitorMonitorToBaseMonitorMapper(t *testing.T) {
	uptimeMonitorObject := UptimeMonitorMonitor{Name: "Test Monitor",
		PK:          124,
		MspAddress:  "https://stakater.com",
		MspInterval: 5,
		CheckType:   "HTTP"}

	monitorObject := UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitorObject)

	if monitorObject.ID != strconv.Itoa(uptimeMonitorObject.PK) ||
		monitorObject.Name != uptimeMonitorObject.Name ||
		monitorObject.URL != uptimeMonitorObject.MspAddress ||
		"5" != monitorObject.Annotations["uptime.monitor.stakater.com/interval"] ||
		"HTTP" != monitorObject.Annotations["uptime.monitor.stakater.com/check_type"] {
		t.Error("Correct: \n",
			uptimeMonitorObject.Name,
			uptimeMonitorObject.PK,
			uptimeMonitorObject.MspAddress,
			uptimeMonitorObject.MspInterval,
			uptimeMonitorObject.CheckType)
		t.Error("Parsed: \n", monitorObject.Name,
			monitorObject.ID,
			monitorObject.URL,
			monitorObject.Annotations["uptime.monitor.stakater.com/interval"],
			monitorObject.Annotations["uptime.monitor.stakater.com/check_type"],
		)
		t.Error("Mapper did not map the values correctly")
	}
}

func TestUptimeMonitorMonitorsToBaseMonitorsMapper(t *testing.T) {
	uptimeMonitorObject1 := UptimeMonitorMonitor{
		Name:          "Test Monitor",
		PK:            124,
		MspAddress:    "https://stakater.com",
		MspInterval:   5,
		CheckType:     "HTTP",
		Locations:     []string{"US-Central"},
		ContactGroups: []string{"Default"}}
	uptimeMonitorObject2 := UptimeMonitorMonitor{
		Name:          "Test Monitor 2",
		PK:            125,
		MspAddress:    "https://facebook.com",
		MspInterval:   10,
		CheckType:     "ICMP",
		Locations:     []string{"US-Central"},
		ContactGroups: []string{"Default"}}

	var annotations1 = map[string]string{
		"uptime.monitor.stakater.com/interval":   "5",
		"uptime.monitor.stakater.com/check_type": "HTTP",
		"uptime.monitor.stakater.com/locations":  "US-Central",
		"uptime.monitor.stakater.com/contacts":   "Default",
	}
	var annotations2 = map[string]string{
		"uptime.monitor.stakater.com/interval":   "10",
		"uptime.monitor.stakater.com/check_type": "ICMP",
		"uptime.monitor.stakater.com/locations":  "US-Central",
		"uptime.monitor.stakater.com/contacts":   "Default",
	}

	correctMonitors := []models.Monitor{
		models.Monitor{
			Name:        "Test Monitor",
			ID:          "124",
			URL:         "https://stakater.com",
			Annotations: annotations1},
		models.Monitor{
			Name:        "Test Monitor 2",
			ID:          "125",
			URL:         "https://facebook.com",
			Annotations: annotations2}}

	var uptimeMonitors []UptimeMonitorMonitor
	uptimeMonitors = append(uptimeMonitors, uptimeMonitorObject1)
	uptimeMonitors = append(uptimeMonitors, uptimeMonitorObject2)

	monitors := UptimeMonitorMonitorsToBaseMonitorsMapper(uptimeMonitors)

	for index := 0; index < len(monitors); index++ {
		if !reflect.DeepEqual(correctMonitors[index], monitors[index]) {
			t.Error("Correct: ", correctMonitors[index], "Parsed", monitors[index])
			t.Error("Monitor object should be the same")
		}
	}
}
