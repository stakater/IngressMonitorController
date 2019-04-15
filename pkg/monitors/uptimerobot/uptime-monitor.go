package uptimerobot

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"strconv"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/http"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

type UpTimeMonitorService struct {
	apiKey            string
	url               string
	alertContacts     string
	statusPageService UpTimeStatusPageService
}

func (monitor *UpTimeMonitorService) Setup(p config.Provider) {
	monitor.apiKey = p.ApiKey
	monitor.url = p.ApiURL
	monitor.alertContacts = p.AlertContacts
	monitor.statusPageService = UpTimeStatusPageService{}
	monitor.statusPageService.Setup(p)
}

func (monitor *UpTimeMonitorService) GetByName(name string) (*models.Monitor, error) {
	action := "getMonitors"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&logs=1" + "&search=" + name

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeMonitorGetMonitorsResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}

		if f.Monitors != nil {
			for _, monitor := range f.Monitors {
				if monitor.FriendlyName == name {
					return UptimeMonitorMonitorToBaseMonitorMapper(monitor), nil
				}
			}
		}

		return nil, nil
	}

	errorString := "GetByName Request failed for name: " + name + ". Status Code: " + strconv.Itoa(response.StatusCode)

	log.Println(errorString)
	return nil, errors.New(errorString)
}

func (monitor *UpTimeMonitorService) GetAll() []models.Monitor {

	action := "getMonitors"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&logs=1"

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {

		var f UptimeMonitorGetMonitorsResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}

		return UptimeMonitorMonitorsToBaseMonitorsMapper(f.Monitors)

	}

	log.Println("GetAllMonitors Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	return nil

}

func (monitor *UpTimeMonitorService) Add(m models.Monitor) {
	action := "newMonitor"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&type=1&url=" + url.QueryEscape(m.URL) + "&friendly_name=" + url.QueryEscape(m.Name) + "&alert_contacts=" + monitor.alertContacts

	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/interval"]; ok {
		body += "&interval=" + val
	}
	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/maintenance-windows"]; ok {
		body += "&mwindows=" + val
	}

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeMonitorNewMonitorResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}

		if f.Stat == "ok" {
			log.Println("Monitor Added: " + m.Name)
			monitor.handleStatusPagesAnnotations(m, strconv.Itoa(f.Monitor.ID))
		} else {
			log.Println("Monitor couldn't be added: " + m.Name)
		}
	} else {
		log.Printf("AddMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	}
}

func (monitor *UpTimeMonitorService) Update(m models.Monitor) {
	action := "editMonitor"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&id=" + m.ID + "&friendly_name=" + m.Name + "&url=" + m.URL + "&alert_contacts=" + monitor.alertContacts

	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/interval"]; ok {
		body += "&interval=" + val
	}
	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/maintenance-windows"]; ok {
		body += "&mwindows=" + val
	}

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeMonitorStatusMonitorResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}

		if f.Stat == "ok" {
			log.Println("Monitor Updated: " + m.Name)
			monitor.handleStatusPagesAnnotations(m, strconv.Itoa(f.Monitor.ID))
		} else {
			log.Println("Monitor couldn't be updated: " + m.Name)
		}
	} else {
		log.Println("UpdateMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	}
}

func (monitor *UpTimeMonitorService) Remove(m models.Monitor) {
	action := "deleteMonitor"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&id=" + m.ID

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeMonitorStatusMonitorResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}

		if f.Stat == "ok" {
			log.Println("Monitor Removed: " + m.Name)
		} else {
			log.Println("Monitor couldn't be removed: " + m.Name)
			log.Println(string(body))
		}
	} else {
		log.Println("RemoveMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	}
}

func (monitor *UpTimeMonitorService) handleStatusPagesAnnotations(monitorToAdd models.Monitor, monitorId string) {
	if val, ok := monitorToAdd.Annotations["uptimerobot.monitor.stakater.com/status-pages"]; ok {
		monitor.updateStatusPages(val, models.Monitor{ID: monitorId})
	}
}

func (monitor *UpTimeMonitorService) updateStatusPages(statusPages string, monitorToAdd models.Monitor) {
	statusPage := UpTimeStatusPage{ID: statusPages}
	_, err := monitor.statusPageService.AddMonitorToStatusPage(statusPage, monitorToAdd)
	if err != nil {
		log.Println("Monitor couldn't be added to status page: " + err.Error())
	}
}
