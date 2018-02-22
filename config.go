package main

type Config struct {
	providers             []Provider
	enableMonitorDeletion bool
}

type Provider struct {
	name          string
	apiKey        string
	apiURL        string
	alertContacts string
}
