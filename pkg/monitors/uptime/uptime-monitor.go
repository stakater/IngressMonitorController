package uptime

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/http"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

type UpTimeMonitorService struct {
	apiKey        string
	url           string
	alertContacts string
}

func (monitor *UpTimeMonitorService) Setup(p config.Provider) {
	monitor.apiKey = p.ApiKey
	monitor.url = p.ApiURL
	monitor.alertContacts = p.AlertContacts
}

func (monitor *UpTimeMonitorService) GetByName(name string) (*models.Monitor, error) {

	monitors := monitor.GetAll()

	if monitors != nil {
		for _, monitor := range monitors {
			if monitor.Name == name {
				return &monitor, nil
			}
		}
	}

	errorString := name + " not found"
	log.Println(errorString)
	return nil, errors.New(errorString)
}

func (monitor *UpTimeMonitorService) GetAll() []models.Monitor {

	action := "checks/"

	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey

	response := client.GetUrl(headers, "")

	if response.StatusCode == 200 {

		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(response.Bytes, &f)

		return UptimeMonitorMonitorsToBaseMonitorsMapper(f.Monitors)

	}

	log.Println("GetAllMonitors Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	return nil

}

func (monitor *UpTimeMonitorService) Add(m models.Monitor) {

	action := "checks/add-api/"
	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey

	body := make(map[string]interface{})
	body["msp_script"] = ""
	body["name"] = m.Name
	body["msp_address"] = m.URL

	if val, ok := m.Annotations["uptime.monitor.stakater.com/interval"]; ok {
		body["msp_interval"] = val
	}

	if val, ok := m.Annotations["uptime.monitor.stakater.com/locations"]; ok {
		body["locations"] = strings.Split(val, ",")
	}

	if val, ok := m.Annotations["uptime.monitor.stakater.com/contacts"]; ok {
		body["contact_groups"] = strings.Split(val, ",")
	}
	bod, err := json.Marshal(body)
	if err != nil {
		jsonbody := string(bod)
		response := client.PostUrl(headers, jsonbody)

		if response.StatusCode == 200 {
			var f UptimeMonitorNewMonitorResponse
			json.Unmarshal(response.Bytes, &f)

			if f.Stat == "ok" {
				log.Println("Monitor Added: " + m.Name)
			} else {
				log.Println("Monitor couldn't be added: " + m.Name)
			}
		} else {
			log.Printf("AddMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
		}
	} else {
		log.Println(err.Error())
	}
}

func (monitor *UpTimeMonitorService) Update(m models.Monitor) {
	targetmonitor, err := monitor.GetByName(m.Name)

	action := "checks/" + targetmonitor.ID + "/"
	if err != nil {
		client := http.CreateHttpClient(monitor.url + action)

		headers := make(map[string]string)
		headers["Authorization"] = "Token " + monitor.apiKey

		body := make(map[string]interface{})
		body["msp_script"] = ""
		body["name"] = m.Name
		body["msp_address"] = m.URL

		if val, ok := m.Annotations["uptime.monitor.stakater.com/interval"]; ok {
			body["msp_interval"] = val
		}

		if val, ok := m.Annotations["uptime.monitor.stakater.com/locations"]; ok {
			body["locations"] = strings.Split(val, ",")
		}

		if val, ok := m.Annotations["uptime.monitor.stakater.com/contacts"]; ok {
			body["contact_groups"] = strings.Split(val, ",")
		}
		bod, err := json.Marshal(body)
		if err != nil {
			jsonbody := string(bod)
			response := client.PostUrl(headers, jsonbody)

			if response.StatusCode == 200 {
				var f UptimeMonitorStatusMonitorResponse
				json.Unmarshal(response.Bytes, &f)

				if f.Stat == "ok" {
					log.Println("Monitor Updated: " + m.Name)
				} else {
					log.Println("Monitor couldn't be updated: " + m.Name)
				}
			} else {
				log.Println("UpdateMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
			}
		}
	} else {
		log.Println("Monitor " + m.Name + " does not exist. Create it first")
	}
}

func (monitor *UpTimeMonitorService) Remove(m models.Monitor) {
	targetmonitor, err := monitor.GetByName(m.Name)

	if err != nil {

		action := "checks/" + targetmonitor.ID + "/"

		client := http.CreateHttpClient(monitor.url + action)

		headers := make(map[string]string)
		headers["Authorization"] = "Token " + monitor.apiKey

		response := client.DeleteUrl(headers, "")

		if response.StatusCode == 200 {
			var f UptimeMonitorStatusMonitorResponse
			json.Unmarshal(response.Bytes, &f)

			if f.Stat == "ok" {
				log.Println("Monitor Removed: " + m.Name)
			} else {
				log.Println("Monitor couldn't be removed: " + m.Name)
			}
		} else {
			log.Println("RemoveMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
		}
	} else {

		log.Println("Monitor " + m.Name + " does not exist. Hence cannot be deleted")
	}
}
