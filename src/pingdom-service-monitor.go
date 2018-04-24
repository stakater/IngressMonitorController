package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/russellcardullo/go-pingdom/pingdom"
	"log"
	"net/url"
	"strconv"
)

type PingdomService struct {
	apiKey        string
	username      string
	password      string
	url           string
	alertContacts string
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
	checkConfig := m.setupPingdomConfigsFromAnnotations(m.annotations)
	checkConfig.Name = m.name
	checkConfig.Paused = false
	resp, err := client.Checks.Create(&checkConfig)
	if err != nil {
		log.Println("error adding service:", err.Error())
	}
	log.Println("Added monitor for: ", m.name)
}

func (monitor *PingdomService) Update(m Monitor) {
	client := pingdom.NewClient(monitor.username, monitor.password, monitor.apiKey)
	checkConfig := m.setupPingdomConfigsFromAnnotations(m.annotations)
	checkConfig.Name = m.name
	checkConfig.Paused = false
	int_id, _ := strconv.Atoi(m.id)
	resp, err := client.Checks.Update(int_id, &checkConfig)
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
	log.Println("Deleting monitoring: ", m.name)
	client.Checks.Delete(int_id)
}

func (monitor *PingdomService) Setup(apiKey string, url string, alertContacts string, username string, password string) {
	monitor.apiKey = apiKey
	monitor.url = url
	monitor.alertContacts = alertContacts
	monitor.username = username
	monitor.password = password
}

func (monitor *Monitor) setupPingdomConfigsFromAnnotations(annotations map[string]string) (checkConfig pingdom.HttpCheck) {
	const resolutionAnnotation = "monitor.stakater.com/pingdom-resolution"
	const sendNotificationWhenDownAnnotation = "monitor.stakater.com/pingdom-send-notification-when-down"
	const teamIdsAnnotation = "monitor.stakater.com/pingdom-team-ids"

	httpCheck := pingdom.HttpCheck{}
	url, err := url.Parse(monitor.url)
	if err != nil {
		log.Println("Unable to parse the url being requested")
	}
	if url.Scheme == "https" {
		httpCheck.Encryption = true
	} else {
		httpCheck.Encryption = false
	}

	httpCheck.Hostname = url.Host
	httpCheck.Url = url.Path
	// Read known annotations, try to map them to pingdom configs
	// set some default values if we can't find them
	if annotations[resolutionAnnotation] != "" {
		intVal, err := strconv.Atoi(annotations[resolutionAnnotation])
		if err != nil {
			log.Println("Error decoding input into an integer")
		}
		httpCheck.Resolution = intVal
	} else {
		httpCheck.Resolution = 1
	}
	if annotations[sendNotificationWhenDownAnnotation] != "" {

		intVal, err := strconv.Atoi(annotations[sendNotificationWhenDownAnnotation])
		if err != nil {
			log.Println("Error decoding input into an integer")
		}
		httpCheck.SendNotificationWhenDown = intVal
	} else {
		httpCheck.SendNotificationWhenDown = 2
	}
	if annotations[teamIdsAnnotation] != "" {
		var is []int
		if err := json.Unmarshal([]byte(annotations[teamIdsAnnotation]), &is); err != nil {
			panic(err)
		}
		httpCheck.TeamIds = is
	}
	return httpCheck
}
