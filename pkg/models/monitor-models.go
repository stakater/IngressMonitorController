package models

type Monitor struct {
	URL    string
	Name   string
	ID     string
	Config interface{}
}

func NewMonitor(monitorName string, id string, monitorUrl string, config interface{}) Monitor {
	return Monitor{
		Name:   monitorName,
		ID:			id,
		URL:    monitorUrl,
		Config: config,
	}
}
