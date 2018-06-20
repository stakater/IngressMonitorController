package monitors

import "testing"

func TestMonitorServiceProxyOfTypeWithCorrectType(t *testing.T) {
	monitorType := "UptimeRobot"
	uptime := (&MonitorServiceProxy{}).OfType(monitorType)

	if uptime.monitorType != monitorType {
		t.Error("Monitor type is not the same")
	}
}

func TestMonitorServiceProxyOfTypeWithWrongType(t *testing.T) {
	assertPanic(t, func() {
		monitorType := "Testing"
		uptime := (&MonitorServiceProxy{}).OfType(monitorType)

		if uptime.monitorType != monitorType {
			t.Error("Monitor type is not the same")
		}
	})
}
