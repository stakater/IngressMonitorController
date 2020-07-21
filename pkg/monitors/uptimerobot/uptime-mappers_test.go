package uptimerobot

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

func TestUptimeMonitorMonitorToBaseMonitorMapper(t *testing.T) {
	uptimeMonitorObject := UptimeMonitorMonitor{FriendlyName: "Test Monitor", ID: 124, URL: "https://stakater.com", Interval: 900}

	monitorObject := UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitorObject)

	providerConfig, _ := monitorObject.Config.(endpointmonitorv1alpha1.UptimeRobotConfig)

	if monitorObject.ID != strconv.Itoa(uptimeMonitorObject.ID) || monitorObject.Name != uptimeMonitorObject.FriendlyName || monitorObject.URL != uptimeMonitorObject.URL || 900 != providerConfig.Interval {
		t.Error("Mapper did not map the values correctly")
	}
}

func TestUptimeMonitorMonitorsToBaseMonitorsMapper(t *testing.T) {
	uptimeMonitorObject1 := UptimeMonitorMonitor{FriendlyName: "Test Monitor 1", ID: 124, URL: "https://stakater.com", Interval: 900}
	uptimeMonitorObject2 := UptimeMonitorMonitor{FriendlyName: "Test Monitor 2", ID: 125, URL: "https://stackator.com", Interval: 600}

	config1 := endpointmonitorv1alpha1.UptimeRobotConfig{
		Interval: 900,
	}

	config2 := endpointmonitorv1alpha1.UptimeRobotConfig{
		Interval: 600,
	}

	correctMonitors := []models.Monitor{{Name: "Test Monitor 1", ID: "124", URL: "https://stakater.com", Config: config1}, {Name: "Test Monitor 2", ID: "125", URL: "https://stackator.com", Config: config2}}

	var uptimeMonitors []UptimeMonitorMonitor
	uptimeMonitors = append(uptimeMonitors, uptimeMonitorObject1)
	uptimeMonitors = append(uptimeMonitors, uptimeMonitorObject2)

	monitors := UptimeMonitorMonitorsToBaseMonitorsMapper(uptimeMonitors)

	for index := 0; index < len(monitors); index++ {
		if !reflect.DeepEqual(correctMonitors[index], monitors[index]) {
			t.Error("Monitor object should be the same")
		}
	}
}

func TestUptimeStatusPageToBaseStatusPageMapper(t *testing.T) {
	uptimePublicStatusPageObject := UptimePublicStatusPage{FriendlyName: "Test Status Page", ID: 124, Monitors: []int{1234, 5678}}

	uptimeStatusPageObject := UptimeStatusPageToBaseStatusPageMapper(uptimePublicStatusPageObject)

	if uptimeStatusPageObject.ID != "124" {
		t.Error("Mapper did not map ID correctly, expected: 124, but was: " + uptimeStatusPageObject.ID)
	}
	if uptimeStatusPageObject.Name != uptimePublicStatusPageObject.FriendlyName {
		t.Error("Mapper did not map the name correctly, expected: " + uptimePublicStatusPageObject.FriendlyName + ", but was: " + uptimeStatusPageObject.Name)
	}
	if !util.ContainsString(uptimeStatusPageObject.Monitors, "1234") || !util.ContainsString(uptimeStatusPageObject.Monitors, "5678") {
		t.Error("Mapper the monitors array correctly, expected: 1234-5678, but got: " + strings.Join(uptimeStatusPageObject.Monitors, "-"))
	}
}
