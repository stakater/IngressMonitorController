package pingdom

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/russellcardullo/go-pingdom/pingdom"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

var log = logf.Log.WithName("pingdom")

// PingdomMonitorService interfaces with MonitorService
type PingdomMonitorService struct {
	apiToken          string
	url               string
	alertContacts     string
	alertIntegrations string
	teamAlertContacts string
	client            *pingdom.Client
}

func (monitor *PingdomMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO: Retrieve oldMonitor config and compare it here
	return false
}

func (service *PingdomMonitorService) Setup(p config.Provider) {
	service.apiToken = p.ApiToken
	service.url = p.ApiURL
	service.alertContacts = p.AlertContacts
	service.alertIntegrations = p.AlertIntegrations
	service.teamAlertContacts = p.TeamAlertContacts

	var err error
	service.client, err = pingdom.NewClientWithConfig(pingdom.ClientConfig{
		APIToken: service.apiToken,
		BaseURL:  service.url,
	})
	if err != nil {
		log.Info("Error Seting Up Monitor Service: ", err.Error())
	}
}

func (service *PingdomMonitorService) GetByName(name string) (*models.Monitor, error) {
	var match *models.Monitor

	monitors := service.GetAll()
	for _, mon := range monitors {
		if mon.Name == name {
			return &mon, nil
		}
	}

	return match, fmt.Errorf("Unable to locate monitor with name %v", name)
}

func (service *PingdomMonitorService) GetAll() []models.Monitor {
	var monitors []models.Monitor

	checks, err := service.client.Checks.List()
	if err != nil {
		log.Info("Error received while listing checks: ", err.Error())
		return nil
	}
	for _, mon := range checks {
		newMon := models.Monitor{
			URL:  mon.Hostname,
			ID:   fmt.Sprintf("%v", mon.ID),
			Name: mon.Name,
		}
		monitors = append(monitors, newMon)
	}

	return monitors
}

func (service *PingdomMonitorService) Add(m models.Monitor) {
	httpCheck := service.createHttpCheck(m)

	_, err := service.client.Checks.Create(&httpCheck)
	if err != nil {
		log.Info("Error Adding Monitor: ", err.Error())
	} else {
		log.Info("Added monitor for: ", m.Name)
	}
}

func (service *PingdomMonitorService) Update(m models.Monitor) {
	httpCheck := service.createHttpCheck(m)
	monitorID, _ := strconv.Atoi(m.ID)

	resp, err := service.client.Checks.Update(monitorID, &httpCheck)
	if err != nil {
		log.Info("Error updating Monitor: ", err.Error())
	} else {
		log.Info("Updated Monitor: ", resp)
	}
}

func (service *PingdomMonitorService) Remove(m models.Monitor) {
	monitorID, _ := strconv.Atoi(m.ID)

	resp, err := service.client.Checks.Delete(monitorID)
	if err != nil {
		log.Info("Error deleting Monitor: ", err.Error())
	} else {
		log.Info("Delete Monitor: ", resp)
	}
}

func (service *PingdomMonitorService) createHttpCheck(monitor models.Monitor) pingdom.HttpCheck {
	httpCheck := pingdom.HttpCheck{}
	url, err := url.Parse(monitor.URL)
	if err != nil {
		log.Info("Unable to parse the URL: ", service.url)
	}

	if url.Scheme == "https" {
		httpCheck.Encryption = true
	} else {
		httpCheck.Encryption = false
	}

	httpCheck.Hostname = url.Host
	httpCheck.Url = url.Path
	httpCheck.Name = monitor.Name
	// Set the default values if they are present in provider config
	// all of them can be overridden via EndpointMonitor specific options
	// Default alert contacts
	if len(service.alertContacts) > 0 {
		userIdsStringArray := strings.Split(service.alertContacts, "-")

		if userIds, err := util.SliceAtoi(userIdsStringArray); err != nil {
			log.Info(err.Error())
		} else {
			httpCheck.UserIds = userIds
		}
	}
	// Default alert integrations
	if len(service.alertIntegrations) > 0 {
		integrationIdsStringArray := strings.Split(service.alertIntegrations, "-")

		if integrationIds, err := util.SliceAtoi(integrationIdsStringArray); err != nil {
			log.Info(err.Error())
		} else {
			httpCheck.IntegrationIds = integrationIds
		}
	}
	// Default team alert contacts
	if len(service.teamAlertContacts) > 0 {
		teamAlertContactsStringArray := strings.Split(service.teamAlertContacts, "-")

		if teamAlertsIds, err := util.SliceAtoi(teamAlertContactsStringArray); err != nil {
			log.Info(err.Error())
		} else {
			httpCheck.TeamIds = teamAlertsIds
		}
	}
	// Generate check itself
	service.addConfigToHttpCheck(&httpCheck, monitor.Config)

	return httpCheck
}

func (service *PingdomMonitorService) addConfigToHttpCheck(httpCheck *pingdom.HttpCheck, config interface{}) {
	// Read config, try to map them to pingdom configs
	// set some default values if we can't find them

	// Retrieve provider configuration
	providerConfig, _ := config.(*endpointmonitorv1alpha1.PingdomConfig)
	if providerConfig != nil && len(providerConfig.AlertContacts) != 0 {
		userIdsStringArray := strings.Split(providerConfig.AlertContacts, "-")

		if userIds, err := util.SliceAtoi(userIdsStringArray); err != nil {
			log.Info("Error decoding user alert contact IDs from config", err.Error())
		} else {
			httpCheck.UserIds = userIds
		}
	}

	if providerConfig != nil && len(providerConfig.AlertIntegrations) != 0 {
		integrationIdsStringArray := strings.Split(providerConfig.AlertIntegrations, "-")

		if integrationIds, err := util.SliceAtoi(integrationIdsStringArray); err != nil {
			log.Info("Error decoding integration ids into integers", err.Error())
		} else {
			httpCheck.IntegrationIds = integrationIds
		}
	}

	if providerConfig != nil && len(providerConfig.TeamAlertContacts) != 0 {
		integrationTeamIdsStringArray := strings.Split(providerConfig.TeamAlertContacts, "-")

		if integrationTeamIdsStringArray, err := util.SliceAtoi(integrationTeamIdsStringArray); err != nil {
			log.Info("Error decoding integration ids into integers", err.Error())
		} else {
			httpCheck.TeamIds = integrationTeamIdsStringArray
		}
	}

	if providerConfig != nil && providerConfig.Resolution > 0 {
		httpCheck.Resolution = providerConfig.Resolution
	} else {
		httpCheck.Resolution = 1
	}

	if providerConfig != nil && providerConfig.SendNotificationWhenDown > 0 {
		httpCheck.SendNotificationWhenDown = providerConfig.SendNotificationWhenDown
	} else {
		httpCheck.SendNotificationWhenDown = 3
	}

	if providerConfig != nil && len(providerConfig.RequestHeaders) > 0 {
		httpCheck.RequestHeaders = make(map[string]string)
		err := json.Unmarshal([]byte(providerConfig.RequestHeaders), &httpCheck.RequestHeaders)
		if err != nil {
			log.Info("Error Converting from string to JSON object")
		}
	}

	if providerConfig != nil && len(providerConfig.BasicAuthUser) > 0 {
		// This should be set to the username to set on the httpCheck
		// Environment variable should define the password
		// Mounted via a secret; key is the username, value the password
		passwordValue := os.Getenv(providerConfig.BasicAuthUser)
		if passwordValue != "" {
			// Env variable set, pass user/pass to httpCheck
			httpCheck.Username = providerConfig.BasicAuthUser
			httpCheck.Password = passwordValue
			log.Info("Basic auth requirement detected. Setting username and password for httpCheck")
		} else {
			log.Info("Error reading basic auth password from environment variable")
		}
	}

	if providerConfig != nil && len(providerConfig.ShouldContain) > 0 {
		httpCheck.ShouldContain = providerConfig.ShouldContain
		log.Info("Should contain detected. Setting Should Contain string: ", providerConfig.ShouldContain)
	}

	// Tags should be a single word or multiple comma-seperated words
	if providerConfig != nil && len(providerConfig.Tags) > 0 {
		if !strings.Contains(providerConfig.Tags, " ") {
			httpCheck.Tags = providerConfig.Tags
			log.Info("Tags detected. Setting Tags as: ", providerConfig.Tags)
		} else {
			log.Info("Tag string should not contain spaces. Not applying tags.")
		}
	}

	if providerConfig != nil {
		httpCheck.Paused = providerConfig.Paused
		httpCheck.NotifyWhenBackup = providerConfig.NotifyWhenBackUp
	}
}
