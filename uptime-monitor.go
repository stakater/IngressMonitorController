package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type UpTimeMonitorService struct {
	apiKey string
	url    string
}

func (monitor *UpTimeMonitorService) Authorize(apiKey string) {
	monitor.apiKey = apiKey
	monitor.url = "https://api.uptimerobot.com/v2/getMonitors"
}

func (monitor *UpTimeMonitorService) GetAll() {
	payload := strings.NewReader("api_key=" + monitor.apiKey + "&format=json&logs=1")

	req, _ := http.NewRequest("POST", monitor.url, payload)

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))
	var f interface{}

	json.Unmarshal(body, &f)
}
