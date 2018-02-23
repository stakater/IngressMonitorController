package main

type MonitorService interface {
	GetAll() []Monitor
	Add(m Monitor)
	Update(m Monitor)
	GetByName(name string) (*Monitor, error)
	Remove(m Monitor)
	Setup(apiKey string, url string, alertContacts string)
}
