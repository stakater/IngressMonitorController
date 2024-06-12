package monitors

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/appinsights"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/gcloud"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/grafana"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/pingdom"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/pingdomtransaction"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/statuscake"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/updown"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/uptime"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors/uptimerobot"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("monitors")

const (
	TypeUptimeRobot        = "UptimeRobot"
	TypePingdom            = "Pingdom"
	TypePingdomTransaction = "PingdomTransaction"
	TypeStatusCake         = "StatusCake"
	TypeUptime             = "Uptime"
	TypeUpdown             = "Updown"
	TypeAppInsights        = "AppInsights"
	TypeGCloud             = "gcloud"
	TypeGrafana            = "Grafana"
)

type MonitorServiceProxy struct {
	monitorType string
	monitor     MonitorService
}

func (mp *MonitorServiceProxy) GetType() string {
	return mp.monitorType
}

func (mp *MonitorServiceProxy) OfType(mType string) MonitorServiceProxy {
	mp.monitorType = mType
	switch mType {
	case TypeUptimeRobot:
		mp.monitor = &uptimerobot.UpTimeMonitorService{}
	case TypePingdom:
		mp.monitor = &pingdom.PingdomMonitorService{}
	case TypePingdomTransaction:
		mp.monitor = &pingdomtransaction.PingdomTransactionMonitorService{}
	case TypeStatusCake:
		mp.monitor = &statuscake.StatusCakeMonitorService{}
	case TypeUptime:
		mp.monitor = &uptime.UpTimeMonitorService{}
	case TypeUpdown:
		mp.monitor = &updown.UpdownMonitorService{}
	case TypeAppInsights:
		mp.monitor = &appinsights.AppinsightsMonitorService{}
	case TypeGCloud:
		mp.monitor = &gcloud.MonitorService{}
	case TypeGrafana:
		mp.monitor = &grafana.GrafanaMonitorService{}
	default:
		panic("No such provider found: " + mType)
	}
	return *mp
}

func (mp *MonitorServiceProxy) ExtractConfig(spec endpointmonitorv1alpha1.EndpointMonitorSpec) interface{} {
	var config interface{}
	switch mp.monitorType {
	case TypeUptimeRobot:
		config = spec.UptimeRobotConfig
	case TypePingdom:
		config = spec.PingdomConfig
	case TypePingdomTransaction:
		config = spec.PingdomTransactionConfig
	case TypeStatusCake:
		config = spec.StatusCakeConfig
	case TypeUptime:
		config = spec.UptimeConfig
	case TypeUpdown:
		config = spec.UpdownConfig
	case TypeAppInsights:
		config = spec.AppInsightsConfig
	case TypeGCloud:
		config = spec.GCloudConfig
	case TypeGrafana:
		config = spec.GrafanaConfig
	default:
		return config
	}
	return config
}

func (mp *MonitorServiceProxy) Setup(p config.Provider) {
	mp.monitor.Setup(p)
}

func (mp *MonitorServiceProxy) GetAll() []models.Monitor {
	return mp.monitor.GetAll()
}

func (mp *MonitorServiceProxy) GetByName(name string) (*models.Monitor, error) {
	return mp.monitor.GetByName(name)
}

func (mp *MonitorServiceProxy) Add(m models.Monitor) {
	mp.monitor.Add(m)
}

func (mp *MonitorServiceProxy) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	return mp.monitor.Equal(oldMonitor, newMonitor)
}

func (mp *MonitorServiceProxy) Update(m models.Monitor) {
	mp.monitor.Update(m)
}

func (mp *MonitorServiceProxy) Remove(m models.Monitor) {
	mp.monitor.Remove(m)
}
