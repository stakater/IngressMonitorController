package uptimerobot

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/http"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

type UpTimeStatusPageService struct {
	apiKey string
	url    string
}

type UpTimeStatusPage struct {
	ID       string
	Name     string
	Monitors []string
}

func (statusPage *UpTimeStatusPageService) Setup(p config.Provider) {
	statusPage.apiKey = p.ApiKey
	statusPage.url = p.ApiURL
}

func (statusPageService *UpTimeStatusPageService) Add(statusPage UpTimeStatusPage) (string, error) {
	action := "newPSP"

	client := http.CreateHttpClient(statusPageService.url + action)

	body := "api_key=" + statusPageService.apiKey + "&type=1&friendly_name=" + url.QueryEscape(statusPage.Name)

	if statusPage.Monitors != nil {
		monitors := strings.Join(statusPage.Monitors, "-")
		body += "&monitors=" + monitors
	} else {
		body += "&monitors=0"
	}

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeNewStatusPageResponse
		log.Println(string(response.Bytes))
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}
		//log.Println("ssssssss", f)
		if f.Stat == "ok" {
			log.Println("Status Page Added: " + statusPage.Name)
			return statusPage.ID, nil
			//return f.statusPage, nil
		} else {
			errorString := "Status Page couldn't be added: " + statusPage.Name
			log.Println(errorString)
			return "", errors.New(errorString)
		}
	} else {
		errorString := "Add Status Page Request failed. Status Code: " + strconv.Itoa(response.StatusCode)
		log.Println(errorString)
		return "", errors.New(errorString)
	}
}

func (statusPageService *UpTimeStatusPageService) Remove(statusPage UpTimeStatusPage) {
	action := "deletePSP"

	client := http.CreateHttpClient(statusPageService.url + action)

	body := "api_key=" + statusPageService.apiKey + "&format=json&id=" + statusPage.ID

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeStatusPageResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response")
		}

		if f.Stat == "ok" {
			log.Println("Status Page Removed: " + statusPage.Name)
		} else {
			log.Println("Status Page couldn't be removed: " + statusPage.Name)
			log.Println(string(body))
		}
	} else {
		log.Println("Remove Status Page Request failed. Status Code: " + strconv.Itoa(response.StatusCode))
	}
}

func (statusPageService *UpTimeStatusPageService) AddMonitorToStatusPage(statusPage UpTimeStatusPage, monitor models.Monitor) (string, error) {
	existingStatusPage, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		errorString := "Updated Page Request failed. Error: " + err.Error()
		log.Println(errorString)
		return "", errors.New(errorString)
	}
	if util.ContainsString(existingStatusPage.Monitors, monitor.ID) {
		log.Println("Status Page Already Up To Date: " + statusPage.ID)
		return statusPage.ID, nil
	} else {
		existingStatusPage.Monitors = append(existingStatusPage.Monitors, monitor.ID)

		action := "editPSP"

		client := http.CreateHttpClient(statusPageService.url + action)

		body := "api_key=" + statusPageService.apiKey + "&format=json&id=" + statusPage.ID

		if existingStatusPage.Monitors != nil {
			monitors := strings.Join(existingStatusPage.Monitors, "-")
			body += "&monitors=" + monitors
		} else {
			body += "&monitors=0"
		}

		response := client.PostUrlEncodedFormBody(body)

		if response.StatusCode == 200 {
			var f UptimeStatusPageResponse
			err := json.Unmarshal(response.Bytes, &f)
			// log.Println(f)
			if err != nil {
				log.Println("Could not Unmarshal Json Response" + err.Error())
			}

			if f.Stat == "ok" {
				log.Println("Status Page Updated: " + statusPage.Name)
				return strconv.Itoa(f.statusPage.ID), nil
			} else {
				errorString := "Status Page couldn't be updated: " + statusPage.Name
				log.Println(errorString)
				return "", errors.New(errorString)
			}
		} else {
			errorString := "Updated Page Request failed. Status Code: " + strconv.Itoa(response.StatusCode)
			log.Println(errorString)
			return "", errors.New(errorString)
		}
	}
}

func (statusPageService *UpTimeStatusPageService) RemoveMonitorFromStatusPage(statusPage UpTimeStatusPage, monitor models.Monitor) (string, error) {
	existingStatusPage, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		errorString := "Updated Page Request failed. Error: " + err.Error()
		log.Println(errorString)
		return "", errors.New(errorString)
	}
	existingStatusPage.Monitors = remove(existingStatusPage.Monitors, monitor.ID)

	action := "editPSP"

	client := http.CreateHttpClient(statusPageService.url + action)

	body := "api_key=" + statusPageService.apiKey + "&format=json&id=" + statusPage.ID

	if existingStatusPage.Monitors != nil && len(existingStatusPage.Monitors) > 0 {
		monitors := strings.Join(existingStatusPage.Monitors, "-")
		body += "&monitors=" + monitors
	} else {
		body += "&monitors=0"
	}

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeStatusPageResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response")
		}

		if f.Stat == "ok" {
			log.Println("Status Page Updated: " + statusPage.Name)
			return strconv.Itoa(f.statusPage.ID), nil
		} else {
			errorString := "Status Page couldn't be updated: " + statusPage.Name
			log.Println(errorString)
			return "", errors.New(errorString)
		}
	} else {
		errorString := "Updated Page Request failed. Status Code: " + strconv.Itoa(response.StatusCode)
		log.Println(errorString)
		return "", errors.New(errorString)
	}
}

func (statusPageService *UpTimeStatusPageService) Get(ID string) (*UpTimeStatusPage, error) {
	action := "getPsps"

	client := http.CreateHttpClient(statusPageService.url + action)

	body := "api_key=" + statusPageService.apiKey + "&format=json&logs=1" + "&psps=" + ID

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeStatusPagesResponse
		err := json.Unmarshal(response.Bytes, &f)
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}

		if f.StatusPages != nil {
			for _, statusPage := range f.StatusPages {
				return UptimeStatusPageToBaseStatusPageMapper(statusPage), nil
			}
		}

		return nil, nil
	}

	errorString := "GetByName Request failed for ID: " + ID + ". Status Code: " + strconv.Itoa(response.StatusCode)

	log.Println(errorString)
	return nil, errors.New(errorString)
}

func (statusPageService *UpTimeStatusPageService) GetStatusPagesForMonitor(ID string) ([]string, error) {
	IDint, _ := strconv.Atoi(ID)

	var matchingStatusPageIds []string

	action := "getPsps"

	client := http.CreateHttpClient(statusPageService.url + action)

	body := "api_key=" + statusPageService.apiKey + "&format=json&logs=1"

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == 200 {
		var f UptimeStatusPagesResponse
		err := json.Unmarshal(response.Bytes, &f)
		//log.Println("=======")
		//log.Println(f)
		//log.Println("=======")
		if err != nil {
			log.Println("Could not Unmarshal Json Response" + err.Error())
		}

		if f.StatusPages != nil {
			for _, statusPage := range f.StatusPages {
				if util.ContainsInt(statusPage.Monitors, IDint) {
					matchingStatusPageIds = append(matchingStatusPageIds, strconv.Itoa(statusPage.ID))
				}
			}
		}

		return matchingStatusPageIds, nil
	}

	errorString := "GetStatusPagesForMonitor Request failed for ID: " + ID + ". Status Code: " + strconv.Itoa(response.StatusCode)

	log.Println(errorString)
	return nil, errors.New(errorString)
}

func remove(s []string, i string) []string {
	j := 0
	for _, n := range s {
		if n != i {
			s[j] = n
			j++
		}
	}
	s = s[:j]
	return s
}
