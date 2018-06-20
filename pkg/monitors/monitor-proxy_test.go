package monitors

import (
	"testing"

	"github.com/stakater/IngressMonitorController/pkg/util"
)

func TestMonitorServiceProxyOfTypeWithCorrectType(t *testing.T) {
	monitorType := "UptimeRobot"
	uptime := (&MonitorServiceProxy{}).OfType(monitorType)

	if uptime.monitorType != monitorType {
		t.Error("Monitor type is not the same")
	}
}

func TestMonitorServiceProxyOfTypeWithWrongType(t *testing.T) {
	util.AssertPanic(t, func() {
		monitorType := "Testing"
		uptime := (&MonitorServiceProxy{}).OfType(monitorType)

		if uptime.monitorType != monitorType {
			t.Error("Monitor type is not the same")
		}
	})
}
