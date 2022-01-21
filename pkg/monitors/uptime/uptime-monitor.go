package uptime

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	Http "net/http"
	"net/url"

	gocache "github.com/patrickmn/go-cache"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/http"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

var cache = gocache.New(5*time.Minute, 5*time.Minute)
var log = logf.Log.WithName("uptime-monitor")

type UpTimeMonitorService struct {
	apiKey        string
	url           string
	alertContacts string
}

func (monitor *UpTimeMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {

	// using processed config to avoid unnecessary update call because of default values
	// like contacts and sorted locations
	oldConfig := processProviderConfig(oldMonitor)
	newConfig := processProviderConfig(newMonitor)
	if !(reflect.DeepEqual(oldConfig, newConfig)) {
		log.Info(fmt.Sprintf("There are some new changes in %s monitor: old: %s, new: %s", newMonitor.Name, oldConfig, newConfig))
		return false
	}
	return true
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
	log.Info(errorString)
	return nil, errors.New(errorString)
}

func (monitor *UpTimeMonitorService) GetAll() []models.Monitor {

	var monitors []UptimeMonitorMonitor
	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey
	headers["Content-Type"] = "application/json"
	pageNo := 1
	val := "notNull"
	next := &val

	cached, found := cache.Get("uptime-checks")
	if found {
		return UptimeMonitorMonitorsToBaseMonitorsMapper(cached.([]UptimeMonitorMonitor))
	}

	// Loop over paginated response until Next is null
	for next != nil {
		var f UptimeMonitorGetMonitorsResponse
		checksUrl := fmt.Sprintf("%schecks/?page=%d", monitor.url, pageNo)
		client := http.CreateHttpClient(checksUrl)
		response := client.GetUrl(headers, []byte(""))
		if response.StatusCode == Http.StatusTooManyRequests {
			log.Info("failed getting monitors due to rate limit")
			ObserveRateLimit(response)
			return nil
		} else if response.StatusCode != Http.StatusOK {
			log.Info("GetAllMonitors Request for Uptime failed. Status Code: " + strconv.Itoa(response.StatusCode))
			return nil
		}

		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Info(fmt.Sprintf("Could not Unmarshal Json Response with error: %v", err))
		}
		monitors = append(monitors, f.Monitors...)
		pageNo++
		next = f.Next
	}
	cache.Set("uptime-checks", monitors, gocache.DefaultExpiration)
	return UptimeMonitorMonitorsToBaseMonitorsMapper(monitors)
}

func (monitor *UpTimeMonitorService) Add(m models.Monitor) {

	defer cache.Flush()
	action := "checks/add-http/"
	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey
	headers["Content-Type"] = "application/json"
	headers["Accepts"] = "application/json"

	body := processProviderConfig(m)

	jsonBody, err := json.Marshal(body)
	if err == nil {
		log.Info(string(jsonBody))
		response := client.PostUrl(headers, jsonBody)

		if response.StatusCode == Http.StatusOK {
			var f UptimeMonitorMonitorResponse

			err := json.Unmarshal(response.Bytes, &f)
			if err != nil {
				log.Info("Failed to Unmarshal Response Json Object")
			}

			if !f.Errors {
				log.Info("Monitor Added: " + m.Name)
			} else {
				log.Info("Monitor couldn't be added: " + m.Name +
					"Response: ")
				log.Info(string(response.Bytes))
			}
		} else if response.StatusCode == Http.StatusTooManyRequests {
			log.Info("failed adding monitor due to rate limit")
			err := ObserveRateLimit(response)
			if err == nil {
				monitor.Add(m)
			}
		} else {
			log.Info("AddMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode) + "\n" + string(response.Bytes))
		}
	} else {
		log.Info(err.Error())
	}

}

func (monitor *UpTimeMonitorService) Update(m models.Monitor) {

	log.Info("Updating Monitor: " + m.Name)
	defer cache.Flush()

	action := "checks/" + m.ID + "/"
	client := http.CreateHttpClient(monitor.url + action)

	headers := make(map[string]string)
	headers["Authorization"] = "Token " + monitor.apiKey
	headers["Content-Type"] = "application/json"

	body := processProviderConfig(m)

	jsonBody, err := json.Marshal(body)
	log.Info(string(jsonBody))
	if err == nil {
		response := client.PutUrl(headers, jsonBody)

		if response.StatusCode == Http.StatusOK {
			var f UptimeMonitorMonitorResponse
			err := json.Unmarshal(response.Bytes, &f)
			if err != nil {
				log.Info("Failed to Unmarshal Response Json Object")
			}
			if !f.Errors {
				log.Info("Monitor Updated: " + m.Name)
			} else {
				log.Info("Monitor couldn't be updated: " + m.Name)
			}
		} else if response.StatusCode == Http.StatusTooManyRequests {
			log.Info("failed updating monitor due to rate limit")
			err := ObserveRateLimit(response)
			if err == nil {
				monitor.Update(m)
			}
		} else {
			log.Info("UpdateMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
		}
	} else {
		log.Info("Failed to Marshal JSON Object")
	}
}

func (monitor *UpTimeMonitorService) Remove(m models.Monitor) {

	defer cache.Flush()
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
			log.Error(err, "Unable to unmarshal JSON")
		}
		if !f.Errors {
			log.Info("Monitor Removed: " + m.Name)
		} else {
			log.Info("Monitor couldn't be removed: " + m.Name)
		}
	} else if response.StatusCode == Http.StatusTooManyRequests {
		log.Info("failed removing monitor due to rate limit")
		err := ObserveRateLimit(response)
		if err == nil {
			monitor.Remove(m)
		}
	} else {
		log.Info("RemoveMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
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

	// sorting locations which is useful during Equal method used in Update.
	if providerConfig != nil && len(providerConfig.Locations) != 0 {
		body["locations"] = util.SplitAndSort(providerConfig.Locations, ",")
	} else {
		locations := strings.Split("US-East,US-West,GBR", ",") // by default 3 lcoations for a check
		sort.Strings(locations)
		body["locations"] = util.SplitAndSort(providerConfig.Locations, ",")
	}

	if providerConfig != nil && len(providerConfig.Contacts) != 0 {
		body["contact_groups"] = util.SplitAndSort(providerConfig.Contacts, ",")
	} else {
		body["contact_groups"] = strings.Split("Default", ",") // use default use email as a contact
	}

	if providerConfig != nil && len(providerConfig.Tags) != 0 {
		body["tags"] = util.SplitAndSort(providerConfig.Tags, ",")
	}

	return body
}

func ObserveRateLimit(resp http.HttpResponse) error {
	strDuration := resp.Headers["retry-after"]
	if strDuration == "" {
		log.Info("No retry-after header was present in the rate limited response")
		return errors.New("No retry-after header was present in the rate limited response")
	}
	intDuration, err := strconv.Atoi(strDuration)
	if err != nil {
		log.Info(fmt.Sprintf("error parsing duration from value: %s", strDuration))
		return errors.New("failed parsing duration value from the retry-after response header")
	}
	log.Info(fmt.Sprintf("rate limit hit, backing off for %s seconds", strDuration))
	time.Sleep(time.Duration(intDuration) * time.Second)

	return nil
}
