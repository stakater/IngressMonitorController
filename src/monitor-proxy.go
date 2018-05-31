package main

import "log"

type MonitorServiceProxy struct {
	monitorType string
	monitor     MonitorService
}

func (mp *MonitorServiceProxy) OfType(mType string) MonitorServiceProxy {
	mp.monitorType = mType
	switch mType {
	case "UptimeRobot":
		mp.monitor = &UpTimeMonitorService{}
	case "Pingdom":
		mp.monitor = &PingdomMonitorService{}
	default:
		log.Panic("No such provider found: ", mType)
	}
	return *mp
}

func (mp *MonitorServiceProxy) Setup(p Provider) {
	mp.monitor.Setup(p)
}

func (mp *MonitorServiceProxy) GetAll() []Monitor {
	return mp.monitor.GetAll()
}

func (mp *MonitorServiceProxy) GetByName(name string) (*Monitor, error) {
	return mp.monitor.GetByName(name)
}

func (mp *MonitorServiceProxy) Add(m Monitor) {
	mp.monitor.Add(m)
}

func (mp *MonitorServiceProxy) Update(m Monitor) {
	mp.monitor.Update(m)
}

func (mp *MonitorServiceProxy) Remove(m Monitor) {
	mp.monitor.Remove(m)
}
