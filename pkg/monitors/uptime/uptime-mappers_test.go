package uptime

import (
	"reflect"
	"strconv"
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

func TestUptimeMonitorMonitorToBaseMonitorMapper(t *testing.T) {
	uptimeMonitorObject := UptimeMonitorMonitor{Name: "Test Monitor",
		PK:          124,
		MspAddress:  "https://stakater.com",
		MspInterval: 5,
		CheckType:   "HTTP"}

	monitorObject := UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitorObject)
	providerConfig, _ := monitorObject.Config.(*endpointmonitorv1alpha1.UptimeConfig)

	if monitorObject.ID != strconv.Itoa(uptimeMonitorObject.PK) ||
		monitorObject.Name != uptimeMonitorObject.Name ||
		monitorObject.URL != uptimeMonitorObject.MspAddress ||
		5 != providerConfig.Interval ||
		"HTTP" != providerConfig.CheckType {
		t.Error("Correct: \n",
			uptimeMonitorObject.Name,
			uptimeMonitorObject.PK,
			uptimeMonitorObject.MspAddress,
			uptimeMonitorObject.MspInterval,
			uptimeMonitorObject.CheckType)
		t.Error("Parsed: \n", monitorObject.Name,
			monitorObject.ID,
			monitorObject.URL,
			providerConfig.Interval,
			providerConfig.CheckType,
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

	config1 := &endpointmonitorv1alpha1.UptimeConfig{
		Interval:  5,
		CheckType: "HTTP",
		Locations: "US-Central",
		Contacts:  "Default",
	}
	config2 := &endpointmonitorv1alpha1.UptimeConfig{
		Interval:  10,
		CheckType: "ICMP",
		Locations: "US-Central",
		Contacts:  "Default",
	}

	correctMonitors := []models.Monitor{
		{
			Name:   "Test Monitor",
			ID:     "124",
			URL:    "https://stakater.com",
			Config: config1},
		{
			Name:   "Test Monitor 2",
			ID:     "125",
			URL:    "https://facebook.com",
			Config: config2}}

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
