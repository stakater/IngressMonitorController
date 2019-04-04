package pingdom

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/russellcardullo/go-pingdom/pingdom"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

const (
	PingdomResolutionAnnotation               = "pingdom.monitor.stakater.com/resolution"
	PingdomSendNotificationWhenDownAnnotation = "pingdom.monitor.stakater.com/send-notification-when-down"
	PingdomPausedAnnotation                   = "pingdom.monitor.stakater.com/paused"
	PingdomNotifyWhenBackUpAnnotation         = "pingdom.monitor.stakater.com/notify-when-back-up"
	PingdomRequestHeadersAnnotation           = "pingdom.monitor.stakater.com/request-headers"
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

func (service *PingdomMonitorService) Setup(p config.Provider) {
	service.apiKey = p.ApiKey
	service.url = p.ApiURL
	service.alertContacts = p.AlertContacts
	service.username = p.Username
	service.password = p.Password
	service.client = pingdom.NewClient(service.username, service.password, service.apiKey)
}

func (service *PingdomMonitorService) GetByName(name string) (*models.Monitor, error) {
	var match *models.Monitor

	monitors := service.GetAll()
	for _, mon := range monitors {
		if mon.Name == name {
			match = &mon
		}
	}

	if match == nil {
		return match, fmt.Errorf("Unable to locate monitor with name %v", name)
	}

	return match, nil
}

func (service *PingdomMonitorService) GetAll() []models.Monitor {
	var monitors []models.Monitor

	checks, err := service.client.Checks.List()
	if err != nil {
		log.Println("Error recevied while listing checks: ", err.Error())
		return nil
	}
	for _, mon := range checks {
		newMon := models.Monitor{
			URL:  mon.Hostname,
			ID:   fmt.Sprintf("%v", mon.ID),
			Name: mon.Name,
		}
		monitors = append(monitors, newMon)
	}

	return monitors
}

func (service *PingdomMonitorService) Add(m models.Monitor) {
	httpCheck := service.createHttpCheck(m)

	_, err := service.client.Checks.Create(&httpCheck)
	if err != nil {
		log.Println("Error Adding Monitor: ", err.Error())
	} else {
		log.Println("Added monitor for: ", m.Name)
	}
}

func (service *PingdomMonitorService) Update(m models.Monitor) {
	httpCheck := service.createHttpCheck(m)
	monitorID, _ := strconv.Atoi(m.ID)

	resp, err := service.client.Checks.Update(monitorID, &httpCheck)
	if err != nil {
		log.Println("Error updating Monitor: ", err.Error())
	} else {
		log.Println("Updated Monitor: ", resp)
	}
}

func (service *PingdomMonitorService) Remove(m models.Monitor) {
	monitorID, _ := strconv.Atoi(m.ID)

	resp, err := service.client.Checks.Delete(monitorID)
	if err != nil {
		log.Println("Error deleting Monitor: ", err.Error())
	} else {
		log.Println("Delete Monitor: ", resp)
	}
}

func (service *PingdomMonitorService) createHttpCheck(monitor models.Monitor) pingdom.HttpCheck {
	httpCheck := pingdom.HttpCheck{}
	url, err := url.Parse(monitor.URL)
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
	httpCheck.Name = monitor.Name

	userIdsStringArray := strings.Split(service.alertContacts, "-")

	if userIds, err := util.SliceAtoi(userIdsStringArray); err != nil {
		log.Println(err.Error())
	} else {
		httpCheck.UserIds = userIds
	}

	service.addAnnotationConfigToHttpCheck(&httpCheck, monitor.Annotations)

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

	if value, ok := annotations[PingdomRequestHeadersAnnotation]; ok {

		httpCheck.RequestHeaders = make(map[string]string)
		err := json.Unmarshal([]byte(value), &httpCheck.RequestHeaders)
		if err != nil {
			log.Println("Error Converting from string to JSON object")
		}
	}
}
