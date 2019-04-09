// Package UpdownMonitor adds updown website monitoring tool's support in IngressMonitorController
package updown

import (
	"log"
	"fmt"
	"net/http"
	// "net/url"

	"github.com/antoineaugusti/updown"
	"github.com/stakater/IngressMonitorController/pkg/config"
    "github.com/stakater/IngressMonitorController/pkg/constants"
	"github.com/stakater/IngressMonitorController/pkg/models"
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


// func (updownService *UpdownMonitorService) createHttpCheck(updownMonitor models.Monitor) updown.Check {
	
// 	updownCheckObj := updown.Check{}
// 	parsedMonitorUrl, err := url.Parse(updownMonitor.URL)
	
// 	if err != nil {
// 		log.Println("Unable to parse the URL : ", updownMonitor.URL)
// 		return updownCheckObj
// 	}
	
// 	if url.Scheme == "https" {
//         updownCheckObj.SSL.Valid = true
// 	} else {
//         updownCheckObj.SSL.Valid = false
// 	}
	
	                     
// 	updownCheckObj.URL = updownMonitor.URL
// 	updownCheckObj.Alias = parsedMonitorUrl.Name

// 	userIdsStringArray := strings.Split(service.alertContacts, "-")

// 	if userIds, err := util.SliceAtoi(userIdsStringArray); err != nil {
// 		log.Println(err.Error())
// 	} else {
// 		httpCheck.UserIds = userIds
// 	}

// 	return updownCheckObj

// }

// Remove function will remove a check
func (updownService *UpdownMonitorService) Remove(updownMonitor models.Monitor) {
	
	// calling remove method it return 3 arguments
	// arg-1  : [boolean] operation was successful or not
	// arg-2  : [HTTP response object] response object
	// arg-3  : [string] response description. nil in success scenario while error message for other
	//                     scenarios
	response, httpResponse, err := updownService.client.Check.Remove(updownMonitor.ID)
    
    if (response == true) && (httpResponse.StatusCode == constants.StatusCodes["OK"]) && (err == nil) {
		
		log.Println("Monitor %v has been deleted.", updownMonitor.Name)
	
	} else if (response == false) && (httpResponse.StatusCode == constants.StatusCodes["NOT_FOUND"]) && (err != nil)  {
	  
		log.Println("Monitor %v is not found.", updownMonitor.Name)
	
	} else {

		log.Println("Unable to delete %v monitor: ", updownMonitor.Name)
	}

}