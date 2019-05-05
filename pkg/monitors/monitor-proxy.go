package monitors

import (
	log "github.com/sirupsen/logrus"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors/pingdom"
	"github.com/stakater/IngressMonitorController/pkg/monitors/statuscake"
	"github.com/stakater/IngressMonitorController/pkg/monitors/updown"
	"github.com/stakater/IngressMonitorController/pkg/monitors/uptime"
	"github.com/stakater/IngressMonitorController/pkg/monitors/uptimerobot"
)

type MonitorServiceProxy struct {
	monitorType string
	monitor     MonitorService
}

func (mp *MonitorServiceProxy) OfType(mType string) MonitorServiceProxy {
	mp.monitorType = mType
	switch mType {
	case "UptimeRobot":
		mp.monitor = &uptimerobot.UpTimeMonitorService{}
	case "Pingdom":
		mp.monitor = &pingdom.PingdomMonitorService{}
	case "StatusCake":
		mp.monitor = &statuscake.StatusCakeMonitorService{}
	case "Uptime":
		mp.monitor = &uptime.UpTimeMonitorService{}
	case "Updown":
		mp.monitor = &updown.UpdownMonitorService{}
	default:
		log.Panic("No such provider found: ", mType)
	}
	return *mp
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

func (mp *MonitorServiceProxy) Update(m models.Monitor) {
	mp.monitor.Update(m)
}

func (mp *MonitorServiceProxy) Remove(m models.Monitor) {
	mp.monitor.Remove(m)
}
