package uptimerobot

import (
	"encoding/json"
	"errors"
	Http "net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

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

	if response.StatusCode == Http.StatusOK {
		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(response.Bytes, &f)

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

func (monitor *UpTimeMonitorService) GetAllByName(name string) ([]models.Monitor, error) {
	action := "getMonitors"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&logs=1" + "&search=" + name

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(response.Bytes, &f)

		if len(f.Monitors) > 0 {
			return UptimeMonitorMonitorsToBaseMonitorsMapper(f.Monitors), nil
		}
		return nil, nil
	}

	errorString := "GetAllByName Request failed for name: " + name + ". Status Code: " + strconv.Itoa(response.StatusCode)

	log.Println(errorString)
	return nil, errors.New(errorString)
}

func (monitor *UpTimeMonitorService) GetAll() []models.Monitor {

	action := "getMonitors"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&logs=1"

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == Http.StatusOK {

		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(response.Bytes, &f)

		return UptimeMonitorMonitorsToBaseMonitorsMapper(f.Monitors)

	}

	log.Println("GetAllMonitors Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	return nil

}

func (monitor *UpTimeMonitorService) Add(m models.Monitor) {
	action := "newMonitor"

	client := http.CreateHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&url=" + url.QueryEscape(m.URL) + "&friendly_name=" + url.QueryEscape(m.Name) + "&alert_contacts=" + monitor.alertContacts

	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/interval"]; ok {
		body += "&interval=" + val
	}
	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/maintenance-windows"]; ok {
		body += "&mwindows=" + val
	}
	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/monitor-type"]; ok {
		if strings.Contains(strings.ToLower(val), "http") {
			body += "&type=1"
		} else if strings.Contains(strings.ToLower(val), "keyword") {
			body += "&type=2"

			if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/keyword-exists"]; ok {

				if strings.Contains(strings.ToLower(val), "yes") {
					body += "&keyword_type=1"
				} else if strings.Contains(strings.ToLower(val), "no") {
					body += "&keyword_type=2"
				}

			} else {
				body += "&keyword_type=1" // By default 1 (check if keyword exists)
			}

			if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/keyword-value"]; ok {
				body += "&keyword_value=" + val
			} else {
				log.Println("Monitor is of type Keyword but the `keyword-value` annotation is missing")
				log.Println("Monitor couldn't be added: " + m.Name)
				return
			}
		}
	} else {
		body += "&type=1" // By default monitor is of type HTTP
	}

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == Http.StatusOK {
		var f UptimeMonitorNewMonitorResponse
		json.Unmarshal(response.Bytes, &f)

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
	if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/monitor-type"]; ok {
		if strings.Contains(strings.ToLower(val), "http") {
			body += "&type=1"
		} else if strings.Contains(strings.ToLower(val), "keyword") {
			body += "&type=2"

			if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/keyword-exists"]; ok {

				if strings.Contains(strings.ToLower(val), "yes") {
					body += "&keyword_type=1"
				} else if strings.Contains(strings.ToLower(val), "no") {
					body += "&keyword_type=2"
				}

			} else {
				body += "&keyword_type=1" // By default 1 (check if keyword exists)
			}

			if val, ok := m.Annotations["uptimerobot.monitor.stakater.com/keyword-value"]; ok {
				body += "&keyword_value=" + val
			} else {
				log.Println("Monitor is of type Keyword but the `keyword-value` annotation is missing")
				log.Println("Monitor couldn't be updated: " + m.Name)
				return
			}
		}
	} else {
		body += "&type=1" // By default monitor is of type HTTP
	}

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == Http.StatusOK {
		var f UptimeMonitorStatusMonitorResponse
		json.Unmarshal(response.Bytes, &f)

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

	log.Println(m.ID)
	body := "api_key=" + monitor.apiKey + "&format=json&id=" + m.ID

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == Http.StatusOK {
		var f UptimeMonitorStatusMonitorResponse
		json.Unmarshal(response.Bytes, &f)

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
