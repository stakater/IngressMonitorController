package main

import (
	"encoding/json"
	"fmt"
	"github.com/russellcardullo/go-pingdom/pingdom"
	"log"
	"net/url"
	"strconv"
)

// PingdomService interfaces with monitor-proxy
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
	_, err := client.Checks.Create(&checkConfig)
	if err != nil {
		log.Println("error adding service:", err.Error())
	}
	log.Println("Added monitor for: ", m.name)
}

func (monitor *PingdomService) Update(m Monitor) {
	client := pingdom.NewClient(monitor.username, monitor.password, monitor.apiKey)
	checkConfig := m.setupPingdomConfigsFromAnnotations(m.annotations)
	intID, _ := strconv.Atoi(m.id)
	resp, err := client.Checks.Update(intID, &checkConfig)
	if err != nil {
		log.Println("Error updating service:", err.Error())
	}
	log.Println("Update service response:", resp)
}

func (monitor *PingdomService) GetByName(name string) (*Monitor, error) {
	var match *Monitor
	allMonitors := monitor.GetAll()
	for _, mon := range allMonitors {
		if mon.name == name {
			match = &mon
		}
	}
	if match == nil {
		return match, fmt.Errorf("Unable to locate service with name %v", name)
	}
	return match, nil
}

func (monitor *PingdomService) Remove(m Monitor) {
	client := pingdom.NewClient(monitor.username, monitor.password, monitor.apiKey)
	intID, _ := strconv.Atoi(m.id)
	log.Println("Deleting monitoring: ", m.name)
	client.Checks.Delete(intID)
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
	const userIdsAnnotation = "monitor.stakater.com/pingdom-user-ids"
	const pausedAnnotation = "monitor.stakater.com/pingdom-paused"
	const notifyWhenBackUpAnnotation = "monitor.stakater.com/pingdom-notify-when-back-up"

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
	httpCheck.Name = monitor.name

	// Read known annotations, try to map them to pingdom configs
	// set some default values if we can't find them
	if annotations[notifyWhenBackUpAnnotation] != "false" {
		httpCheck.NotifyWhenBackup = true
	}
	if annotations[pausedAnnotation] == "true" {
		httpCheck.Paused = true
	}
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
		httpCheck.SendNotificationWhenDown = 3
	}
	// This feels wrong, by im trying to take a string with a slice of ints
	// and make them an actual slice of ints. oh well.
	if annotations[userIdsAnnotation] != "" {
		var intSlice []int
		if err := json.Unmarshal([]byte(annotations[userIdsAnnotation]), &intSlice); err != nil {
			log.Println(err.Error())
		}
		httpCheck.UserIds = intSlice
	}
	return httpCheck
}
