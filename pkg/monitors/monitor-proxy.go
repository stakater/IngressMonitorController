package monitors

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors/statuscake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("monitors")

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
	case "StatusCake":
		mp.monitor = &statuscake.StatusCakeMonitorService{}
	default:
		panic("No such provider found: " + mType)
	}
	return *mp
}

func (mp *MonitorServiceProxy) ExtractConfig(spec endpointmonitorv1alpha1.EndpointMonitorSpec) interface{} {
	var config interface{}

	switch mp.monitorType {
	case "StatusCake":
		config = spec.StatusCakeConfig
	default:
		return config
	}
	return config
}

func (mp *MonitorServiceProxy) Setup(p config.Provider) {
	mp.monitor.Setup(p)
}

func (mp *MonitorServiceProxy) GetAll() ([]models.Monitor, error) {
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
