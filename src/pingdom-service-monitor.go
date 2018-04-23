package main

import (
	"errors"
	"fmt"
	"github.com/russellcardullo/go-pingdom/pingdom"
	"log"
	"strconv"
)

type PingdomService struct {
	apiKey        string
	username      string
	password      string
	url           string
	alertContacts string
}

type Monitor struct {
	url  string
	name string
	id   string
}

func (monitor *PingdomService) GetAll() []Monitor {
	var monitors []Monitor
	client := pingdom.NewClient(monitor.username, monitor.password, monitor.apiKey)
	checks, err := client.Checks.List()
	if err != nil {
		log.Println("error recieved listing checks: " + err.Error())
	}
	for _, mon := range checks {
		newMon := Monitor{
			url:  mon.Hostname,
			id:   fmt.Sprintf("%v", mon.ID),
			name: mon.Name,
		}
		monitors = append(monitors, newMon)
	}
	return monitors
}

func (monitor *PingdomService) Add(m Monitor) {
	client := pingdom.NewClient(monitor.username, monitor.password, monitor.apiKey)
	newCheck := pingdom.HttpCheck{Name: m.name, Hostname: m.url, Resolution: 1, SendNotificationWhenDown: 3, Encryption: true, Paused: false}
	resp, err := client.Checks.Create(&newCheck)
	if err != nil {
		log.Println("error adding service:", err.Error())
	}
	log.Println("Add service response:", resp)
}

func (monitor *PingdomService) Update(m Monitor) {
	client := pingdom.NewClient(monitor.username, monitor.password, monitor.apiKey)
	updatedCheck := pingdom.HttpCheck{Name: m.name, Hostname: m.url, Resolution: 1, SendNotificationWhenDown: 3, Encryption: true, Paused: false}
	int_id, _ := strconv.Atoi(m.id)
	resp, err := client.Checks.Update(int_id, &updatedCheck)
	if err != nil {
		log.Println("Error updating service:", err.Error())
	}
	log.Println("Update service response:", resp)
}

func (monitor *PingdomService) GetByName(name string) (*Monitor, error) {
	var match *Monitor
	all_monitors := monitor.GetAll()
	for _, mon := range all_monitors {
		if mon.name == name {
			match = &mon
		}
	}
	if match != nil {
		return match, nil
	} else {
		return match, errors.New(fmt.Sprintf("Unable to locate service with name %v", name))
	}
}

func (monitor *PingdomService) Remove(m Monitor) {
	client := pingdom.NewClient(monitor.username, monitor.password, monitor.apiKey)
	int_id, _ := strconv.Atoi(m.id)
	client.Checks.Delete(int_id)
}

func (monitor *PingdomService) Setup(apiKey string, url string, alertContacts string) {
	monitor.apiKey = apiKey
	monitor.url = url
	monitor.alertContacts = alertContacts
}
