package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// StatusCakeMonitorService is the service structure for StatusCake
type StatusCakeMonitorService struct {
	apiKey   string
	url      string
	username string
	client   *http.Client
}

// AnnotationInfo is the annotation information structure for AnnotationMap
type AnnotationInfo struct {
	name     string
	dataType string
}

// AnnotationMap holds all the enabled annotations for StatusCake
var AnnotationMap = map[string]AnnotationInfo{
	"monitor.stakater.com/statuscake-check-rate":      AnnotationInfo{"CheckRate", "int"},       // Int (0-24000)
	"monitor.stakater.com/statuscake-test-type":       AnnotationInfo{"TestType", "string"},     // String (HTTP, TCP, PING)
	"monitor.stakater.com/statuscake-paused":          AnnotationInfo{"Paused", "bool"},         // Int (0,1)
	"monitor.stakater.com/statuscake-ping-url":        AnnotationInfo{"PingURL", "string"},      // String (url)
	"monitor.stakater.com/statuscake-follow-redirect": AnnotationInfo{"FollowRedirect", "bool"}, // Int (0,1)
	"monitor.stakater.com/statuscake-port":            AnnotationInfo{"Port", "int"},            // Int (TCP Port)
	"monitor.stakater.com/statuscake-trigger-rate":    AnnotationInfo{"TriggerRate", "int"},     // Int (0-60)
	"monitor.stakater.com/statuscake-contact-group":   AnnotationInfo{"ContactGroup", "string"}} // String (0-60)

// buildUpsertForm function is used to create the form needed to Add or update a monitor
func buildUpsertForm(m Monitor) url.Values {
	f := url.Values{}
	f.Add("WebsiteName", m.name)
	f.Add("WebsiteURL", m.url)
	if val, ok := m.annotations["monitor.stakater.com/statuscake/check-rate"]; ok {
		f.Add("CheckRate", val)
		delete(m.annotations, "monitor.stakater.com/statuscake/check-rate")
	} else {
		f.Add("CheckRate", "300")
	}
	if val, ok := m.annotations["monitor.stakater.com/statuscake/test-type"]; ok {
		f.Add("TestType", val)
		delete(m.annotations, "monitor.stakater.com/statuscake/test-type")
	} else {
		f.Add("TestType", "HTTP")
	}
	for key, value := range m.annotations {
		if (AnnotationInfo{}) != AnnotationMap[key] {
			meta := AnnotationMap[key]
			switch strings.ToLower(meta.dataType) {
			case "int":
				f.Add(meta.name, value)
			case "string":
				f.Add(meta.name, value)
			case "bool":
				value = strings.ToLower(value)
				if value == "true" {
					f.Add(meta.name, "1")
				} else {
					f.Add(meta.name, "0")
				}
			}
		}
	}
	return f
}

// Setup function is used to initialise the StatusCake service
func (service *StatusCakeMonitorService) Setup(p Provider) {
	service.apiKey = p.ApiKey
	service.url = p.ApiURL
	service.username = p.Username
	service.client = &http.Client{}
}

// GetByName function will Get a monitor by it's name
func (service *StatusCakeMonitorService) GetByName(name string) (*Monitor, error) {
	monitors := service.GetAll()
	for _, monitor := range monitors {
		if monitor.name == name {
			return &monitor, nil
		}
	}
	errorString := "GetByName Request failed for name: " + name
	return nil, errors.New(errorString)
	// u, err := url.Parse(service.url)
	// if err != nil {
	// 	return nil, err
	// }
	// u.Path = "/API/Tests/Details/"
	// u.Scheme = "https"
	// query := u.Query()
	// query.Set("TestID", name)
	// u.RawQuery = query.Encode()
	// req, err := http.NewRequest("GET", u.String(), nil)
	// if err != nil {
	// 	return nil, err
	// }
	// req.Header.Add("API", service.apiKey)
	// req.Header.Add("Username", service.username)
	// resp, err := service.client.Do(req)
	// if err != nil {
	// 	return nil, err
	// }
	// if resp.StatusCode == 200 {
	// 	var f StatusCakeMonitorMonitor
	// 	err := json.NewDecoder(resp.Body).Decode(&f)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return StatusCakeMonitorMonitorToBaseMonitorMapper(f), nil
	// }
	// errorString := "GetByName Request failed for name: " + name
	// log.Println(errorString)
	// return nil, errors.New(errorString)
}

// GetAll function will fetch all monitors
func (service *StatusCakeMonitorService) GetAll() []Monitor {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Println(err)
		return nil
	}
	u.Path = "/API/Tests/"
	u.Scheme = "https"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	if resp.StatusCode == 200 {
		f := make([]StatusCakeMonitorMonitor, 0)
		err := json.NewDecoder(resp.Body).Decode(&f)
		if err != nil {
			log.Println(err)
			return nil
		}
		return StatusCakeMonitorMonitorsToBaseMonitorsMapper(f)
	} else {
		errorString := "GetAll Request failed"
		log.Println(errorString)
		return nil
	}
}

// Add will create a new Monitor
func (service *StatusCakeMonitorService) Add(m Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = "/API/Tests/Update"
	u.Scheme = "https"
	data := buildUpsertForm(m)
	req, err := http.NewRequest("PUT", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode == 200 {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Println(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Added:", fa.InsertID)
		} else {
			log.Println("Monitor couldn't be added: " + m.name)
			log.Println(fa.Message)
		}
	} else {
		errorString := "Insert Request failed for name: " + m.name
		log.Println(errorString)
	}
}

// Update will update an existing Monitor
func (service *StatusCakeMonitorService) Update(m Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = "/API/Tests/Update"
	u.Scheme = "https"
	data := buildUpsertForm(m)
	data.Add("TestID", m.id)
	req, err := http.NewRequest("PUT", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode == 200 {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Println(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Updated:", m.name)
		} else {
			log.Println("Monitor couldn't be updated: " + m.name)
			log.Println(fa.Message)
		}
	} else {
		errorString := "Update Request failed for name: " + m.name
		log.Println(errorString)
	}
}

// Remove will delete an existing Monitor
func (service *StatusCakeMonitorService) Remove(m Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = "/API/Tests/Details"
	u.Scheme = "https"
	query := u.Query()
	query.Set("TestID", m.id)
	u.RawQuery = query.Encode()
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode == 200 {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Println(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Deleted:", m.id)
		} else {
			log.Println("Monitor couldn't be deleted: " + m.name)
			log.Println(fa.Message)
		}
	} else {
		errorString := "Delete Request failed for name: " + m.name
		log.Println(errorString)
	}
}
