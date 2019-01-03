package uptimerobot

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stakater/IngressMonitorController/pkg/models"
)

func TestUptimeMonitorMonitorToBaseMonitorMapper(t *testing.T) {
	uptimeMonitorObject := UptimeMonitorMonitor{FriendlyName: "Test Monitor", ID: 124, URL: "https://stakater.com", Interval: 900}

	monitorObject := UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitorObject)

	if monitorObject.ID != strconv.Itoa(uptimeMonitorObject.ID) || monitorObject.Name != uptimeMonitorObject.FriendlyName || monitorObject.URL != uptimeMonitorObject.URL || "900" != monitorObject.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("Mapper did not map the values correctly")
	}
}

func TestUptimeMonitorMonitorsToBaseMonitorsMapper(t *testing.T) {
	uptimeMonitorObject1 := UptimeMonitorMonitor{FriendlyName: "Test Monitor 1", ID: 124, URL: "https://stakater.com", Interval: 900}
	uptimeMonitorObject2 := UptimeMonitorMonitor{FriendlyName: "Test Monitor 2", ID: 125, URL: "https://stackator.com", Interval: 600}

    var annotations1 = map[string]string {
        "uptimerobot.monitor.stakater.com/interval": "900",
    }
    var annotations2 = map[string]string {
        "uptimerobot.monitor.stakater.com/interval": "600",
    }

	correctMonitors := []models.Monitor{models.Monitor{Name: "Test Monitor 1", ID: "124", URL: "https://stakater.com", Annotations: annotations1}, models.Monitor{Name: "Test Monitor 2", ID: "125", URL: "https://stackator.com", Annotations: annotations2}}

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
