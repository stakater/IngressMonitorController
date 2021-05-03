// Package UpdownMonitor adds updown website monitoring tool's support in IngressMonitorController
package updown

import (
	"fmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
		"net/http"
	"net/url"

	"github.com/antoineaugusti/updown"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

const (
	// Default value for updown monitor
	UpdownPeriodDefaultValue    = 15
	UpdownPublishedDefaultValue = true
	UpdownEnableDefaultValue    = true
)

// UpdownMonitorService struct contains parameters required by updown go client
type UpdownMonitorService struct {
	apiKey string
	client *updown.Client
}

func (monitor *UpdownMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO: Retrieve oldMonitor config and compare it here
	return false
}

// Setup method will initialize a updown's go client object by using the configuration parameters
func (updownService *UpdownMonitorService) Setup(confProvider config.Provider) {

	// initializeCustomLog(os.Stdout)
	log.Info("Updown monitor's Setup has been called. Updown monitor initializing")

	// updown go client apiKey
	updownService.apiKey = confProvider.ApiKey

	// creating updown go client
	updownService.client = updown.NewClient(updownService.apiKey, http.DefaultClient)
	log.Info("Updown monitor has been initialized")
}

// GetAll function will return all monitors (updown checks) object in an array
func (updownService *UpdownMonitorService) GetAll() []models.Monitor {

	log.Info("Updown monitor's GetAll method has been called")

	var monitors []models.Monitor

	// getting all monitors(checks) list
	updownChecks, httpResponse, err := updownService.client.Check.List()
	log.Info("Monitors (updown checks) object list has been pulled")

	if (httpResponse.StatusCode == http.StatusOK) && (err == nil) {
		log.Info("Populating monitors list using the updownChecks object given in updownChecks list")

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
		log.Info("Unable to get updown provider checks(monitor) list")
		return nil

	}

}

// GetByName function will return a monitor(updown check) object based on the name provided
func (updownService *UpdownMonitorService) GetByName(monitorName string) (*models.Monitor, error) {

	log.Info("Updown monitor's GetByName method has been called")

	updownMonitors := updownService.GetAll()
	log.Info("Monitors (updown checks) object list has been pulled")

	log.Info("Searching the monitor from monitors object list using its name")
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

	log.Info("Updown monitor's Add method has been called")

	updownCheckItemObj := service.createHttpCheck(updownMonitor)

	_, httpResponse, err := service.client.Check.Add(updownCheckItemObj)
	log.Info("Monitor addition request has been completed")

	if (httpResponse.StatusCode == http.StatusCreated) && (err == nil) {
		log.Printf("Monitor %s has been added.", updownMonitor.Name)

	} else if (httpResponse.StatusCode == http.StatusBadRequest) && (err != nil) {
		log.Printf("Monitor %s is not created because of invalid parameters or it exists.", updownMonitor.Name)

	} else {
		log.Printf("Unable to create monitor %s ", updownMonitor.Name)

	}

}

// createHttpCheck method it will populate updown CheckItem object using updownMonitor's attributes
// and config
func (updownService *UpdownMonitorService) createHttpCheck(updownMonitor models.Monitor) updown.CheckItem {

	log.Info("Updown monitor's createHttpCheck method has been called")

	// populating updownCheckItemObj object attributes using updownMonitor object
	log.Info("Populating updownCheckItemObj object attributes using updownMonitor object")
	updownCheckItemObj := updown.CheckItem{}

	log.Info("Parsing URL")
	_, err := url.Parse(updownMonitor.URL)

	if err != nil {
		log.Info("Unable to parse the URL : ", updownMonitor.URL)
		return updownCheckItemObj
	}

	unEscapedURL, _ := url.QueryUnescape(updownMonitor.URL)
	updownCheckItemObj.URL = unEscapedURL
	updownCheckItemObj.Alias = updownMonitor.Name

	// populating updownCheckItemObj object attributes using Provider Config
	updownService.addConfigToHttpCheck(&updownCheckItemObj, updownMonitor.Config)

	return updownCheckItemObj
}

// addConfigToHttpCheck method will populate Updown's CheckItem object attributes using provider config
func (service *UpdownMonitorService) addConfigToHttpCheck(updownCheckItemObj *updown.CheckItem, config interface{}) {
	// Read provider config, try to map them to updown check configs
	// set some default values if we can't find them

	// Retrieve provider configuration
	providerConfig, _ := config.(*endpointmonitorv1alpha1.UpdownConfig)

	if providerConfig != nil {
		updownCheckItemObj.Enabled = providerConfig.Enable
	} else {
		log.Info("Using default value `true` for enable")
		updownCheckItemObj.Enabled = UpdownEnableDefaultValue
	}

	if providerConfig != nil {
		updownCheckItemObj.Published = providerConfig.PublishPage
	} else {
		log.Info("Using default value `true` for publish-page")
		updownCheckItemObj.Published = UpdownPublishedDefaultValue
	}

	if providerConfig != nil && providerConfig.Period > 0 {
		updownCheckItemObj.Period = providerConfig.Period
	} else {
		log.Info("Using default value `15` for period")
		updownCheckItemObj.Period = UpdownPeriodDefaultValue
	}
}

// Update method will update a monitor (updown check)
func (service *UpdownMonitorService) Update(updownMonitor models.Monitor) {

	log.Info("Updown's Update method has been called")

	httpCheckItemObj := service.createHttpCheck(updownMonitor)
	_, httpResponse, err := service.client.Check.Update(updownMonitor.ID, httpCheckItemObj)
	log.Info("Updown's check Update request has been completed")

	if (httpResponse.StatusCode == http.StatusOK) && (err == nil) {
		log.Printf("Monitor %s has been updated with following parameters", updownMonitor.Name)

	} else {
		log.Printf("Monitor %s is not updated because of %s", updownMonitor.Name, err.Error())

	}

}

// Remove method will remove a monitor (updown check)
func (updownService *UpdownMonitorService) Remove(updownMonitor models.Monitor) {

	log.Info("Updown's Remove method has been called")

	_, httpResponse, err := updownService.client.Check.Remove(updownMonitor.ID)
	log.Info("Updown's check Remove request has been completed")

	if (httpResponse.StatusCode == http.StatusOK) && (err == nil) {
		log.Printf("Monitor %v has been deleted.", updownMonitor.Name)

	} else if (httpResponse.StatusCode == http.StatusNotFound) && (err != nil) {
		log.Printf("Monitor %v is not found.", updownMonitor.Name)

	} else {
		log.Printf("Unable to delete %v monitor: ", updownMonitor.Name)
	}

}
