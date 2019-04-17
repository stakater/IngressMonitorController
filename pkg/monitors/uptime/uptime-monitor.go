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
	headers["Content-Type"] = "application/json"

	response := client.GetUrl(headers, []byte(""))

	if response.StatusCode == 200 {

		var f UptimeMonitorGetMonitorsResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response")
		}
		if f.Count == 0 {
			return []models.Monitor{}
		} else {
			return UptimeMonitorMonitorsToBaseMonitorsMapper(f.Monitors)
		}

	}

	log.Println("GetAllMonitors Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	return nil

}

func (monitor *UpTimeMonitorService) Add(m models.Monitor) {

	action := "checks/add-http/"
	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey
	headers["Content-Type"] = "application/json"
	headers["Accepts"] = "application/json"

	body := make(map[string]interface{})
	body["name"] = m.Name
	body["msp_address"] = m.URL

	if val, ok := m.Annotations["uptime.monitor.stakater.com/interval"]; ok {
		interval, err := strconv.Atoi(val)
		if nil == err {
			body["msp_interval"] = interval
		}
	}

	if val, ok := m.Annotations["uptime.monitor.stakater.com/locations"]; ok {
		body["locations"] = strings.Split(val, ",")
	}

	if val, ok := m.Annotations["uptime.monitor.stakater.com/contacts"]; ok {
		body["contact_groups"] = strings.Split(val, ",")
	}
	jsonBody, err := json.Marshal(body)
	if err == nil {
		log.Println(string(jsonBody))
		response := client.PostUrl(headers, jsonBody)

		if response.StatusCode == 200 {
			var f UptimeMonitorMonitorResponse

			err := json.Unmarshal(response.Bytes, &f)
			if err != nil {
				log.Println("Failed to Unmarshal Response Json Object")
			}

			if f.Errors == false {
				log.Println("Monitor Added: " + m.Name)
			} else {
				log.Print("Monitor couldn't be added: " + m.Name +
					"Response: ")
				log.Println(string(response.Bytes))
			}
		} else {
			log.Printf("AddMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
		}
	} else {
		log.Println(err.Error())
	}
}

func (monitor *UpTimeMonitorService) Update(m models.Monitor) {

	action := "checks/" + m.ID + "/"
	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey
	headers["Content-Type"] = "application/json"

	body := make(map[string]interface{})
	body["name"] = m.Name
	body["msp_address"] = m.URL

	if val, ok := m.Annotations["uptime.monitor.stakater.com/interval"]; ok {
		interval, err := strconv.Atoi(val)
		if nil == err {
			body["msp_interval"] = interval
		}
	}

	if val, ok := m.Annotations["uptime.monitor.stakater.com/locations"]; ok {
		body["locations"] = strings.Split(val, ",")
	}

	if val, ok := m.Annotations["uptime.monitor.stakater.com/contacts"]; ok {
		body["contact_groups"] = strings.Split(val, ",")
	}
	jsonBody, err := json.Marshal(body)
	log.Println(string(jsonBody))
	if err == nil {
		response := client.PutUrl(headers, jsonBody)

		if response.StatusCode == 200 {
			var f UptimeMonitorMonitorResponse
			err := json.Unmarshal(response.Bytes, &f)
			if err != nil {
				log.Println("Failed to Unmarshal Response Json Object")
			}
			if f.Errors == false {
				log.Println("Monitor Updated: " + m.Name)
			} else {
				log.Println("Monitor couldn't be updated: " + m.Name)
			}
		} else {
			log.Println("UpdateMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
		}
	} else {
		log.Println("Failed to Marshal JSON Object")
	}
}

func (monitor *UpTimeMonitorService) Remove(m models.Monitor) {
	action := "checks/" + m.ID + "/"

	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey
	headers["Content-Type"] = "application/json"

	response := client.DeleteUrl(headers, []byte(""))

	if response.StatusCode == 200 {
		var f UptimeMonitorMonitorResponse
		json.Unmarshal(response.Bytes, &f)

		if f.Errors == false {
			log.Println("Monitor Removed: " + m.Name)
		} else {
			log.Println("Monitor couldn't be removed: " + m.Name)
		}
	} else {
		log.Println("RemoveMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	}
}
