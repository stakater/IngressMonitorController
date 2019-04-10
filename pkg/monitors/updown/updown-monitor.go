// Package UpdownMonitor adds updown website monitoring tool's support in IngressMonitorController
package updown

import (
	"log"
	"fmt"
	"net/url"
	"strconv"
	"net/http"
	"encoding/json"

	"github.com/antoineaugusti/updown"
	"github.com/stakater/IngressMonitorController/pkg/config"
    "github.com/stakater/IngressMonitorController/pkg/constants"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

const (

	UpdownEnableCheckAnnotation                         = "updown.monitor.stakater.com/enable"
	UpdownPeriodAnnotation                              = "updown.monitor.stakater.com/period"
	UpdownRequestHeadersAnnotation                      = "updown.monitor.stakater.com/request-headers"

	// Default value for annotations
	UpdownPeriodAnnotationDefaultValue                  = 15
	
)

// UpdownMonitorService struct contains parameters required by updown go client
type UpdownMonitorService struct {
	apiKey        string
	alertContacts string
	client        *updown.Client
}

// Setup function will create a updown's go client object by using the configuration parameters
func (updownService *UpdownMonitorService) Setup(confProvider config.Provider) {
	// configuration parameters for creating a updown client
	updownService.apiKey = confProvider.ApiKey
	updownService.alertContacts = confProvider.AlertContacts
	// creating updown go client
	updownService.client = updown.NewClient(updownService.apiKey, http.DefaultClient)
}

// GetAll function will return all checks object in an array
func (updownService *UpdownMonitorService) GetAll() []models.Monitor {

	var monitors []models.Monitor

	// getting all checks list
	updownChecks, httpResponse, err := updownService.client.Check.List()

	if (httpResponse.StatusCode == constants.StatusCodes["OK"]) && (err == nil) {

		// populating a monitors slice using the updownChecks objects given in updownChecks slice
		for _, updownCheck := range updownChecks {
			newMonitor := models.Monitor{
				URL:  updownCheck.URL,
				Name: updownCheck.Alias,
				ID:   updownCheck.Token,
			}
			monitors = append(monitors, newMonitor)
		}
		return monitors

	} else {
		log.Println("Unable to get updown provider checks(monitor) list")
		return nil
	}

}

// GetByName function will return a monitor(updown check) object based on the name provided 
func (updownService *UpdownMonitorService) GetByName(monitorName string) (*models.Monitor, error) {

	updownMonitors := updownService.GetAll()
	for _, updownMonitor := range updownMonitors {
		if updownMonitor.Name == monitorName {
			// Test the code below
			// match = &mon
			return &updownMonitor, nil
		}
	}

	return nil, fmt.Errorf("Unable to locate updown provider monitor with name %v", monitorName)
}

// Remove function will remove a check
func (updownService *UpdownMonitorService) Remove(updownMonitor models.Monitor) {
	
	// calling remove method it return 3 arguments
	// arg-1  : [boolean] operation was successful or not
	// arg-2  : [HTTP response object] response object
	// arg-3  : [string] response description. nil in success scenario while error message for other
	//                     scenarios
	result, httpResponse, err := updownService.client.Check.Remove(updownMonitor.ID)
    
    if (result) && (httpResponse.StatusCode == constants.StatusCodes["OK"]) && (err == nil) {
		log.Printf("Monitor %v has been deleted.", updownMonitor.Name)
	
	} else if (!result) && (httpResponse.StatusCode == constants.StatusCodes["NOT_FOUND"]) && (err != nil)  {
		log.Printf("Monitor %v is not found.", updownMonitor.Name)
	
	} else {
		log.Printf("Unable to delete %v monitor: ", updownMonitor.Name)
	}

}

func (service *UpdownMonitorService) Add(updownMonitor models.Monitor) {
	updownCheckItemObj := service.createHttpCheck(updownMonitor)

	_, httpResponse, err := service.client.Check.Add(updownCheckItemObj)
	if (httpResponse.StatusCode == constants.StatusCodes["CREATED"]) && (err == nil) {
		log.Printf("Monitor %s has been added.", updownMonitor.Name)

	} else if (httpResponse.StatusCode == constants.StatusCodes["BAD_REQUEST"]) && (err != nil ) {
		log.Printf("Monitor %s is not created because of invalid parameters or it exists.", updownMonitor.Name)

	} else {
		log.Printf("Unable to create monitor %s ", updownMonitor.Name)
	}
	
}

func (service *UpdownMonitorService) Update(updownMonitor models.Monitor) {
	httpCheckItemObj := service.createHttpCheck(updownMonitor)
	
	_, httpResponse, err := service.client.Check.Update(updownMonitor.ID, httpCheckItemObj)
	
	if (httpResponse.StatusCode == constants.StatusCodes["OK"]) && (err == nil) {
		marshaledConfig, _ := json.Marshal(httpCheckItemObj)
		log.Printf("Monitor %s has been updated with following parameters: %s ", updownMonitor.Name, marshaledConfig)

	} else {
		log.Printf("Monitor %s is not updated because of %s", updownMonitor.Name, err.Error())

	}

}


// createHttpCheck it will create a httpCheck
func (updownService *UpdownMonitorService) createHttpCheck(updownMonitor models.Monitor) updown.CheckItem {
	
	updownCheckItemObj := updown.CheckItem{}
	_, err := url.Parse(updownMonitor.URL)

	if err != nil {
		log.Println("Unable to parse the URL : ", updownMonitor.URL)
		return updownCheckItemObj
	}

	// if parsedMonitorUrl.Scheme == "https" {
    //     updownCheckObj.SSL.Valid = true
	// } else {
    //     updownCheckObj.SSL.Valid = false
	// }

	updownCheckItemObj.URL = updownMonitor.URL
   
	updownService.addAnnotationConfigToHttpCheck(&updownCheckItemObj, updownMonitor.Annotations)

	return updownCheckItemObj
}


func (service *UpdownMonitorService) addAnnotationConfigToHttpCheck(httpCheckItem *updown.CheckItem, annotations map[string] string) {
	// Read known annotations, try to map them to updown check configs
	// set some default values if we can't find them

    // updown check enabled not
	if value, ok := annotations[UpdownEnableCheckAnnotation]; ok {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			httpCheckItem.Enabled = boolValue
		}
	}

	// updown check interval aka period
	// 15, 30, 60, 120, 300, 600, 1800 or 3600 recommended period values
	if value, ok := annotations[UpdownPeriodAnnotation]; ok {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			httpCheckItem.Period = intValue
		} else {
			log.Println("Error decoding input into an integer")
			httpCheckItem.Period = UpdownPeriodAnnotationDefaultValue
		}
	} else {
		httpCheckItem.Period = UpdownPeriodAnnotationDefaultValue
	}
    
	// updown check item request header
	if value, ok := annotations[UpdownRequestHeadersAnnotation]; ok {
		httpCheckItem.CustomHeaders = make(map[string]string)
		err := json.Unmarshal([]byte(value), &httpCheckItem.CustomHeaders)
		if err != nil {
			log.Println("Error Converting from string to JSON object")
		}
	}

}