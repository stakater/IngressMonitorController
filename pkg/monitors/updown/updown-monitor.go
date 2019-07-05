// Package UpdownMonitor adds updown website monitoring tool's support in IngressMonitorController
package updown

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/antoineaugusti/updown"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

const (
	UpdownEnableCheckAnnotation = "updown.monitor.stakater.com/enable"
	UpdownPeriodAnnotation      = "updown.monitor.stakater.com/period"
	UpdownPublishPageAnnotation = "updown.monitor.stakater.com/publish-page"
	// this annotation is not enabled
	UpdownRequestHeadersAnnotation = "updown.monitor.stakater.com/request-headers"

	// Default value for annotations
	UpdownPeriodAnnotationDefaultValue    = 15
	UpdownPublishedAnnotationDefaultValue = true
	UpdownEnableAnnotationDefaultValue    = true
)

// UpdownMonitorService struct contains parameters required by updown go client
type UpdownMonitorService struct {
	apiKey string
	client *updown.Client
}

// Setup method will initialize a updown's go client object by using the configuration parameters
func (updownService *UpdownMonitorService) Setup(confProvider config.Provider) {

	// initializeCustomLog(os.Stdout)
	log.Println("Updown monitor's Setup has been called. Updown monitor initializing")

	// updown go client apiKey
	updownService.apiKey = confProvider.ApiKey

	// creating updown go client
	updownService.client = updown.NewClient(updownService.apiKey, http.DefaultClient)
	log.Println("Updown monitor has been initialized")
}

// GetAll function will return all monitors (updown checks) object in an array
func (updownService *UpdownMonitorService) GetAll() []models.Monitor {

	log.Println("Updown monitor's GetAll method has been called")

	var monitors []models.Monitor

	// getting all monitors(checks) list
	updownChecks, httpResponse, err := updownService.client.Check.List()
	log.Println("Monitors (updown checks) object list has been pulled")

	if (httpResponse.StatusCode == http.StatusOK) && (err == nil) {
		log.Println("Populating monitors list using the updownChecks object given in updownChecks list")

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

	log.Println("Updown monitor's GetByName method has been called")

	updownMonitors := updownService.GetAll()
	log.Println("Monitors (updown checks) object list has been pulled")

	log.Println("Searching the monitor from monitors object list using its name")
	for _, updownMonitor := range updownMonitors {
		if updownMonitor.Name == monitorName {
			// Test the code below
			return &updownMonitor, nil
		}
	}

	return nil, fmt.Errorf("Unable to locate %v monitor", monitorName)
}

// Add function method will add a monitor (updown check)
func (service *UpdownMonitorService) Add(updownMonitor models.Monitor) {

	log.Println("Updown monitor's Add method has been called")

	updownCheckItemObj := service.createHttpCheck(updownMonitor)

	_, httpResponse, err := service.client.Check.Add(updownCheckItemObj)
	log.Println("Monitor addition request has been completed")

	if (httpResponse.StatusCode == http.StatusCreated) && (err == nil) {
		log.Printf("Monitor %s has been added.", updownMonitor.Name)

	} else if (httpResponse.StatusCode == http.StatusBadRequest) && (err != nil) {
		log.Printf("Monitor %s is not created because of invalid parameters or it exists.", updownMonitor.Name)

	} else {
		log.Printf("Unable to create monitor %s ", updownMonitor.Name)

	}

}

// createHttpCheck method it will populate updown CheckItem object using updownMonitor's attributes
// and annotations
func (updownService *UpdownMonitorService) createHttpCheck(updownMonitor models.Monitor) updown.CheckItem {

	log.Println("Updown monitor's createHttpCheck method has been called")

	// populating updownCheckItemObj object attributes using updownMonitor object
	log.Println("Populating updownCheckItemObj object attributes using updownMonitor object")
	updownCheckItemObj := updown.CheckItem{}

	log.Println("Parsing URL")
	_, err := url.Parse(updownMonitor.URL)

	if err != nil {
		log.Println("Unable to parse the URL : ", updownMonitor.URL)
		return updownCheckItemObj
	}

	unEscapedURL, _ := url.QueryUnescape(updownMonitor.URL)
	updownCheckItemObj.URL = unEscapedURL
	updownCheckItemObj.Alias = updownMonitor.Name

	// populating updownCheckItemObj object attributes using
	log.Println("Populating updownCheckItemObj object attributes using annotations")
	updownService.addAnnotationConfigToHttpCheck(&updownCheckItemObj, updownMonitor.Annotations)

	return updownCheckItemObj
}

// addAnnotationConfigToHttpCheck method will populate Updown's CheckItem object attributes using the annotations map
func (service *UpdownMonitorService) addAnnotationConfigToHttpCheck(updownCheckItemObj *updown.CheckItem, annotations map[string]string) {
	// Read known annotations, try to map them to updown check configs
	// set some default values if we can't find them

	log.Println("Updown monitor's addAnnotationConfigToHttpCheck method has been called")

	// Enable Annotation
	if value, ok := annotations[UpdownEnableCheckAnnotation]; ok {
		boolValue, err := strconv.ParseBool(value)

		if err == nil {
			updownCheckItemObj.Enabled = boolValue

		} else {
			log.Println("Error decoding input into an boolean")
			updownCheckItemObj.Enabled = UpdownEnableAnnotationDefaultValue

		}
	}

	// Published Annotation
	if value, ok := annotations[UpdownPublishPageAnnotation]; ok {
		boolValue, err := strconv.ParseBool(value)

		if err == nil {
			updownCheckItemObj.Published = boolValue

		} else {
			log.Println("Error decoding input into an boolean")
			updownCheckItemObj.Published = UpdownPublishedAnnotationDefaultValue

		}
	}

	// Period Annotation
	if value, ok := annotations[UpdownPeriodAnnotation]; ok {
		intValue, err := strconv.Atoi(value)

		if err == nil {
			updownCheckItemObj.Period = intValue

		} else {
			log.Println("Error decoding input into an integer")
			updownCheckItemObj.Period = UpdownPeriodAnnotationDefaultValue

		}

	} else {
		updownCheckItemObj.Period = UpdownPeriodAnnotationDefaultValue

	}

}

// Update method will update a monitor (updown check)
func (service *UpdownMonitorService) Update(updownMonitor models.Monitor) {

	log.Println("Updown's Update method has been called")

	httpCheckItemObj := service.createHttpCheck(updownMonitor)
	_, httpResponse, err := service.client.Check.Update(updownMonitor.ID, httpCheckItemObj)
	log.Println("Updown's check Update request has been completed")

	if (httpResponse.StatusCode == http.StatusOK) && (err == nil) {
		log.Printf("Monitor %s has been updated with following parameters", updownMonitor.Name)

	} else {
		log.Printf("Monitor %s is not updated because of %s", updownMonitor.Name, err.Error())

	}

}

// Remove method will remove a monitor (updown check)
func (updownService *UpdownMonitorService) Remove(updownMonitor models.Monitor) {

	log.Println("Updown's Remove method has been called")

	_, httpResponse, err := updownService.client.Check.Remove(updownMonitor.ID)
	log.Println("Updown's check Remove request has been completed")

	if (httpResponse.StatusCode == http.StatusOK) && (err == nil) {
		log.Printf("Monitor %v has been deleted.", updownMonitor.Name)

	} else if (httpResponse.StatusCode == http.StatusNotFound) && (err != nil) {
		log.Printf("Monitor %v is not found.", updownMonitor.Name)

	} else {
		log.Printf("Unable to delete %v monitor: ", updownMonitor.Name)
	}

}
