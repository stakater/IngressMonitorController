package models

type Monitor struct {
	URL         string
	Name        string
	ID          string
	Annotations map[string]string
}

func NewMonitor(monitorName string, monitorUrl string) (Monitor) {
	return Monitor{
		Name: monitorName,
		URL:  monitorUrl,
	}
}