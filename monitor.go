package main

type MonitorService interface {
	GetAll()
	GetByName()
	Add()
	Remove()
	Authorize(apiKey string)
}

type Monitor struct {
	url  string
	name string
}
