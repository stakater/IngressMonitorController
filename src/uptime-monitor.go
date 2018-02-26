package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/glog"
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
	payload := strings.NewReader("api_key=" + monitor.apiKey + "&format=json&logs=1" + "&search=" + name)

	req, _ := http.NewRequest("POST", monitor.url+action, payload)

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	if res.StatusCode == 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)

		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(body, &f)

		if f.Monitors != nil && len(f.Monitors) > 0 {
			return UptimeMonitorMonitorToBaseMonitorMapper(f.Monitors[0]), nil
		}
		return nil, nil
	}

	errorString := "GetByName Request failed"
	glog.Errorln(errorString)
	return nil, errors.New(errorString)
}

func (monitor *UpTimeMonitorService) GetAll() []Monitor {

	action := "getMonitors"
	payload := strings.NewReader("api_key=" + monitor.apiKey + "&format=json&logs=1")

	req, _ := http.NewRequest("POST", monitor.url+action, payload)

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	if res.StatusCode == 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)

		var f UptimeMonitorGetMonitorsResponse
		json.Unmarshal(body, &f)

		return UptimeMonitorMonitorsToBaseMonitorsMapper(f.Monitors)

	}

	glog.Errorln("GetAllMonitors Request failed")
	return nil

}

func (monitor *UpTimeMonitorService) Add(m Monitor) {

	action := "newMonitor"
	payload := strings.NewReader("api_key=" + monitor.apiKey + "&format=json&type=1&url=" + url.QueryEscape(m.url) + "&friendly_name=" + url.QueryEscape(m.name) + "&alert_contacts=" + monitor.alertContacts)

	req, _ := http.NewRequest("POST", monitor.url+action, payload)

	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, _ := http.DefaultClient.Do(req)

	if res.StatusCode == 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)

		var f UptimeMonitorNewMonitorResponse
		json.Unmarshal(body, &f)

		if f.Stat == "ok" {
			fmt.Println("Monitor Added")
		} else {
			fmt.Println("Monitor couldn't be added")
			fmt.Println(string(body))
		}
	} else {
		glog.Errorln("AddMonitor Request failed")
	}
}

func (monitor *UpTimeMonitorService) Update(m Monitor) {
	action := "editMonitor"

	payload := strings.NewReader("api_key=" + monitor.apiKey + "&format=json&id=" + m.id + "&friendly_name=" + m.name + "&url=" + m.url)

	req, _ := http.NewRequest("POST", monitor.url+action, payload)

	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, _ := http.DefaultClient.Do(req)

	if res.StatusCode == 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)

		var f UptimeMonitorStatusMonitorResponse
		json.Unmarshal(body, &f)

		if f.Stat == "ok" {
			fmt.Println("Monitor Updated")
		} else {
			fmt.Println("Monitor couldn't be updated")
			fmt.Println(string(body))
		}
	} else {
		glog.Errorln("UpdateMonitor Request failed")
	}
}

func (monitor *UpTimeMonitorService) Remove(m Monitor) {
	action := "deleteMonitor"

	payload := strings.NewReader("api_key=" + monitor.apiKey + "&format=json&id=" + m.id)

	req, _ := http.NewRequest("POST", monitor.url+action, payload)

	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, _ := http.DefaultClient.Do(req)

	if res.StatusCode == 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)

		var f UptimeMonitorStatusMonitorResponse
		json.Unmarshal(body, &f)

		if f.Stat == "ok" {
			fmt.Println("Monitor Removed")
		} else {
			fmt.Println("Monitor couldn't be removed")
			fmt.Println(string(body))
		}
	} else {
		glog.Errorln("RemoveMonitor Request failed")
	}
}
