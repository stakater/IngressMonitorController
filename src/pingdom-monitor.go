package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

const (
	PingdomResolutionAnnotation               = "monitor.stakater.com/pingdom/resolution"
	PingdomSendNotificationWhenDownAnnotation = "monitor.stakater.com/pingdom/send-notification-when-down"
	PingdomPausedAnnotation                   = "monitor.stakater.com/pingdom/paused"
	PingdomNotifyWhenBackUpAnnotation         = "monitor.stakater.com/pingdom/notify-when-back-up"
)

// PingdomMonitorService interfaces with MonitorService
type PingdomMonitorService struct {
	apiKey        string
	url           string
	alertContacts string
	username      string
	password      string
	client        *pingdom.Client
}

func (service *PingdomMonitorService) Setup(p Provider) {
	service.apiKey = p.ApiKey
	service.url = p.ApiURL
	service.alertContacts = p.AlertContacts
	service.username = p.Username
	service.password = p.Password
	service.client = pingdom.NewClient(service.username, service.password, service.apiKey)
}

func (service *PingdomMonitorService) GetByName(name string) (*Monitor, error) {
	var match *Monitor

	monitors := service.GetAll()
	for _, mon := range monitors {
		if mon.name == name {
			match = &mon
		}
	}

	if match == nil {
		return match, fmt.Errorf("Unable to locate monitor with name %v", name)
	}

	return match, nil
}

func (service *PingdomMonitorService) GetAll() []Monitor {
	var monitors []Monitor

	checks, err := service.client.Checks.List()
	if err != nil {
		log.Println("Error recevied while listing checks: ", err.Error())
		return nil
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

func (service *PingdomMonitorService) Add(m Monitor) {
	httpCheck := service.createHttpCheck(m)

	_, err := service.client.Checks.Create(&httpCheck)
	if err != nil {
		log.Println("Error Adding Monitor: ", err.Error())
	} else {
		log.Println("Added monitor for: ", m.name)
	}
}

func (service *PingdomMonitorService) Update(m Monitor) {
	httpCheck := service.createHttpCheck(m)
	monitorID, _ := strconv.Atoi(m.id)

	resp, err := service.client.Checks.Update(monitorID, &httpCheck)
	if err != nil {
		log.Println("Error updating Monitor: ", err.Error())
	} else {
		log.Println("Updated Monitor: ", resp)
	}
}

func (service *PingdomMonitorService) Remove(m Monitor) {
	monitorID, _ := strconv.Atoi(m.id)

	resp, err := service.client.Checks.Delete(monitorID)
	if err != nil {
		log.Println("Error deleting Monitor: ", err.Error())
	} else {
		log.Println("Delete Monitor: ", resp)
	}
}

func (service *PingdomMonitorService) createHttpCheck(monitor Monitor) pingdom.HttpCheck {
	httpCheck := pingdom.HttpCheck{}
	url, err := url.Parse(service.url)
	if err != nil {
		log.Println("Unable to parse the URL: ", service.url)
	}

	if url.Scheme == "https" {
		httpCheck.Encryption = true
	} else {
		httpCheck.Encryption = false
	}

	httpCheck.Hostname = url.Host
	httpCheck.Url = url.Path
	httpCheck.Name = monitor.name

	userIdsStringArray := strings.Split(service.alertContacts, "-")

	if userIds, err := sliceAtoi(userIdsStringArray); err != nil {
		log.Println(err.Error())
	} else {
		httpCheck.UserIds = userIds
	}

	service.addAnnotationConfigToHttpCheck(&httpCheck, monitor.annotations)

	return httpCheck
}

func (service *PingdomMonitorService) addAnnotationConfigToHttpCheck(httpCheck *pingdom.HttpCheck, annotations map[string]string) {
	// Read known annotations, try to map them to pingdom configs
	// set some default values if we can't find them

	if value, ok := annotations[PingdomNotifyWhenBackUpAnnotation]; ok {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			httpCheck.NotifyWhenBackup = boolValue
		}
	}

	if value, ok := annotations[PingdomPausedAnnotation]; ok {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			httpCheck.Paused = boolValue
		}
	}

	if value, ok := annotations[PingdomResolutionAnnotation]; ok {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			httpCheck.Resolution = intValue
		} else {
			log.Println("Error decoding input into an integer")
			httpCheck.Resolution = 1
		}
	} else {
		httpCheck.Resolution = 1
	}

	if value, ok := annotations[PingdomSendNotificationWhenDownAnnotation]; ok {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			httpCheck.SendNotificationWhenDown = intValue
		} else {
			log.Println("Error decoding input into an integer")
			httpCheck.SendNotificationWhenDown = 3
		}
	} else {
		httpCheck.SendNotificationWhenDown = 3
	}

}
