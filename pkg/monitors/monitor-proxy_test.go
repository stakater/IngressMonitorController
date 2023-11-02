package monitors

import (
	"testing"

	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// To allow normal logging to be printed if tests fails
	// Dev mode is an extra feature to make output more readable
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

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
