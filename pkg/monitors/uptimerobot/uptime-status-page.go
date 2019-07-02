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

	body := "api_key=" + statusPageService.apiKey + "&format=json&friendly_name=" + url.QueryEscape(statusPage.Name)

	if statusPage.Monitors != nil {
		monitors := strings.Join(statusPage.Monitors, "-")
		body += "&monitors=" + monitors
	} else {
		body += "&monitors=0"
	}

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == Http.StatusOK {
		var f UptimeStatusPageResponse
		json.Unmarshal(response.Bytes, &f)

		if f.Stat == "ok" {
			log.Println("Status Page Added: " + statusPage.Name)
			return strconv.Itoa(f.UptimePublicStatusPage.ID), nil
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

	if response.StatusCode == Http.StatusOK {
		var f UptimeStatusPageResponse
		json.Unmarshal(response.Bytes, &f)

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

		if response.StatusCode == Http.StatusOK {
			var f UptimeStatusPageResponse
			json.Unmarshal(response.Bytes, &f)

			if f.Stat == "ok" {
				log.Println("Status Page Updated: " + statusPage.Name)
				return strconv.Itoa(f.UptimePublicStatusPage.ID), nil
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

	if response.StatusCode == Http.StatusOK {
		var f UptimeStatusPageResponse
		json.Unmarshal(response.Bytes, &f)

		if f.Stat == "ok" {
			log.Println("Status Page Updated: " + statusPage.Name)
			return strconv.Itoa(f.UptimePublicStatusPage.ID), nil
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

	if response.StatusCode == Http.StatusOK {
		var f UptimeStatusPagesResponse
		json.Unmarshal(response.Bytes, &f)

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

func (statusPageService *UpTimeStatusPageService) GetAllStatusPages(name string) ([]UpTimeStatusPage, error) {
	statusPages := []UpTimeStatusPage{}
	action := "getPsps"

	client := http.CreateHttpClient(statusPageService.url + action)

	body := "api_key=" + statusPageService.apiKey + "&format=json&logs=1"

	response := client.PostUrlEncodedFormBody(body)

	if response.StatusCode == Http.StatusOK {
		var f UptimeStatusPagesResponse
		err := json.Unmarshal(response.Bytes, &f)

		if err == nil && len(f.StatusPages) > 0 {
			for _, statusPage := range f.StatusPages {
				if statusPage.FriendlyName == name {
					sp := UptimeStatusPageToBaseStatusPageMapper(statusPage)
					statusPages = append(statusPages, *sp)
				}
			}
			return statusPages, nil
		}

		return nil, nil
	}

	errorString := "GetAllStatusPages Request failed for: " + name + ". Status Code: " + strconv.Itoa(response.StatusCode)

	log.Println(errorString)
	return nil, errors.New(errorString)
}

func (statusPageService *UpTimeStatusPageService) GetStatusPagesForMonitor(ID string) ([]string, error) {
	IDint, _ := strconv.Atoi(ID)

	var matchingStatusPageIds []string
	var f UptimeStatusPagesResponse

	// Initial dummy values
	f.Pagination.Limit = -1
	f.Pagination.Total = 0
	f.Pagination.Offset = 0
	f.StatusPages = []UptimePublicStatusPage{}

	action := "getPsps"

	client := http.CreateHttpClient(statusPageService.url + action)

	if f.StatusPages != nil {
		for f.Pagination.Limit < f.Pagination.Total {

			body := "api_key=" + statusPageService.apiKey + "&format=json&logs=1&offset=" + strconv.Itoa(f.Pagination.Offset)

			response := client.PostUrlEncodedFormBody(body)

			if response.StatusCode == Http.StatusOK {

				json.Unmarshal(response.Bytes, &f)

				for _, statusPage := range f.StatusPages {
					if util.ContainsInt(statusPage.Monitors, IDint) {
						matchingStatusPageIds = append(matchingStatusPageIds, strconv.Itoa(statusPage.ID))
					}
				}
			}
		}
		return matchingStatusPageIds, nil
	}

	errorString := "GetStatusPagesForMonitor Request failed for ID: " + ID

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
