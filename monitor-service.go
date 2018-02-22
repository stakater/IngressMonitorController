package main

type MonitorService interface {
	GetAll() []Monitor
	Add(m Monitor)
	GetByName(name string) (*Monitor, error)
	// Remove()
	Setup(apiKey string, url string, alertContacts string)
}
