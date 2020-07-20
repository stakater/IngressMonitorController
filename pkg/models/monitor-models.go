package models

type Monitor struct {
	URL    string
	Name   string
	ID     string
	Config interface{}
}

func NewMonitor(monitorName string, monitorUrl string, config interface{}) Monitor {
	return Monitor{
		Name:   monitorName,
		URL:    monitorUrl,
		Config: config,
	}
}
