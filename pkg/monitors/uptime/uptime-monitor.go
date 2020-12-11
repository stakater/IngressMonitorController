package uptime

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	Http "net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/http"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

type UpTimeMonitorService struct {
	apiKey        string
	url           string
	alertContacts string
}

func (monitor *UpTimeMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO: Retrieve oldMonitor config and compare it here
	return false
}

func (monitor *UpTimeMonitorService) Setup(p config.Provider) {
	monitor.apiKey = p.ApiKey
	monitor.url = p.ApiURL
	monitor.alertContacts = p.AlertContacts
}

func (monitor *UpTimeMonitorService) GetByName(name string) (*models.Monitor, error) {

	monitors := monitor.GetAll()

	for _, monitor := range monitors {
		if monitor.Name == name {
			return &monitor, nil
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

	if response.StatusCode == Http.StatusOK {

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

	log.Println("GetAllMonitors Request for Uptime failed. Status Code: " + strconv.Itoa(response.StatusCode))
	return nil

}

func (monitor *UpTimeMonitorService) Add(m models.Monitor) {

	action := "checks/add-http/"
	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey
	headers["Content-Type"] = "application/json"
	headers["Accepts"] = "application/json"

	body := processProviderConfig(m)

	jsonBody, err := json.Marshal(body)
	if err == nil {
		log.Println(string(jsonBody))
		response := client.PostUrl(headers, jsonBody)

		if response.StatusCode == Http.StatusOK {
			var f UptimeMonitorMonitorResponse

			err := json.Unmarshal(response.Bytes, &f)
			if err != nil {
				log.Println("Failed to Unmarshal Response Json Object")
			}

			if !f.Errors {
				log.Println("Monitor Added: " + m.Name)
			} else {
				log.Print("Monitor couldn't be added: " + m.Name +
					"Response: ")
				log.Println(string(response.Bytes))
			}
		} else {
			log.Printf("AddMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode) + "\n" + string(response.Bytes))
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

	body := processProviderConfig(m)

	jsonBody, err := json.Marshal(body)
	log.Println(string(jsonBody))
	if err == nil {
		response := client.PutUrl(headers, jsonBody)

		if response.StatusCode == Http.StatusOK {
			var f UptimeMonitorMonitorResponse
			err := json.Unmarshal(response.Bytes, &f)
			if err != nil {
				log.Println("Failed to Unmarshal Response Json Object")
			}
			if !f.Errors {
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

	if response.StatusCode == Http.StatusOK {
		var f UptimeMonitorMonitorResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Error(err)
		}
		if !f.Errors {
			log.Println("Monitor Removed: " + m.Name)
		} else {
			log.Println("Monitor couldn't be removed: " + m.Name)
		}
	} else {
		log.Println("RemoveMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	}
}

func processProviderConfig(m models.Monitor) map[string]interface{} {

	// Retrieve provider configuration
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.UptimeConfig)

	body := make(map[string]interface{})
	body["name"] = m.Name
	unEscapedURL, _ := url.QueryUnescape(m.URL)
	body["msp_address"] = unEscapedURL

	if providerConfig != nil && providerConfig.Interval > 0 {
		body["msp_interval"] = strconv.Itoa(providerConfig.Interval)
	} else {
		body["msp_interval"] = 5 // by default interval check is 5 minutes
	}

	if providerConfig != nil && len(providerConfig.Locations) != 0 {
		body["locations"] = strings.Split(providerConfig.Locations, ",")
	} else {
		body["locations"] = strings.Split("US-East,US-West,GBR", ",") // by default 3 lcoations for a check
	}

	if providerConfig != nil && len(providerConfig.Contacts) != 0 {
		body["contact_groups"] = strings.Split(providerConfig.Contacts, ",")
	} else {
		body["contact_groups"] = strings.Split("Default", ",") // use default use email as a contact
	}

	if providerConfig != nil && len(providerConfig.Tags) != 0 {
		body["tags"] = strings.Split(providerConfig.Tags, ",")
	}

	return body
}
