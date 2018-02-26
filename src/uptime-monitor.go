package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
)

type UpTimeMonitorService struct {
	apiKey        string
	url           string
	alertContacts string
}

func (monitor *UpTimeMonitorService) Setup(apiKey string, url string, alertContacts string) {
	monitor.apiKey = apiKey
	monitor.url = url
	monitor.alertContacts = alertContacts
}

func (monitor *UpTimeMonitorService) GetByName(name string) (*Monitor, error) {
	action := "getMonitors"

	client := createHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&logs=1" + "&search=" + name

	response := client.postUrlEncodedFormBody(body)

	if response.statusCode == 200 {

		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(response.bytes, &f)

		if f.Monitors != nil && len(f.Monitors) > 0 {
			return UptimeMonitorMonitorToBaseMonitorMapper(f.Monitors[0]), nil
		}
		return nil, nil
	}

	errorString := "GetByName Request failed"

	log.Println(errorString)
	return nil, errors.New(errorString)
}

func (monitor *UpTimeMonitorService) GetAll() []Monitor {

	action := "getMonitors"

	client := createHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&logs=1"

	response := client.postUrlEncodedFormBody(body)

	if response.statusCode == 200 {

		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(response.bytes, &f)

		return UptimeMonitorMonitorsToBaseMonitorsMapper(f.Monitors)

	}

	log.Println("GetAllMonitors Request failed")
	return nil

}

func (monitor *UpTimeMonitorService) Add(m Monitor) {
	action := "newMonitor"

	client := createHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&type=1&url=" + url.QueryEscape(m.url) + "&friendly_name=" + url.QueryEscape(m.name) + "&alert_contacts=" + monitor.alertContacts

	response := client.postUrlEncodedFormBody(body)

	if response.statusCode == 200 {
		var f UptimeMonitorNewMonitorResponse
		json.Unmarshal(response.bytes, &f)

		if f.Stat == "ok" {
			log.Println("Monitor Added")
		} else {
			log.Println("Monitor couldn't be added")
			log.Println(string(body))
		}
	} else {
		log.Printf("AddMonitor Request failed")
	}
}

func (monitor *UpTimeMonitorService) Update(m Monitor) {
	action := "editMonitor"

	client := createHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&id=" + m.id + "&friendly_name=" + m.name + "&url=" + m.url

	response := client.postUrlEncodedFormBody(body)

	if response.statusCode == 200 {
		var f UptimeMonitorStatusMonitorResponse
		json.Unmarshal(response.bytes, &f)

		if f.Stat == "ok" {
			log.Println("Monitor Updated")
		} else {
			log.Println("Monitor couldn't be updated")
			log.Println(string(body))
		}
	} else {
		log.Println("UpdateMonitor Request failed")
	}
}

func (monitor *UpTimeMonitorService) Remove(m Monitor) {
	action := "deleteMonitor"

	client := createHttpClient(monitor.url + action)

	body := "api_key=" + monitor.apiKey + "&format=json&id=" + m.id

	response := client.postUrlEncodedFormBody(body)

	if response.statusCode == 200 {
		var f UptimeMonitorStatusMonitorResponse
		json.Unmarshal(response.bytes, &f)

		if f.Stat == "ok" {
			log.Println("Monitor Removed")
		} else {
			log.Println("Monitor couldn't be removed")
			log.Println(string(body))
		}
	} else {
		log.Println("RemoveMonitor Request failed")
	}
}
