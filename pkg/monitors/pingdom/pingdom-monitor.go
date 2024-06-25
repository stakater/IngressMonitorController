package pingdom

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/russellcardullo/go-pingdom/pingdom"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
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

func (service *PingdomMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
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
		log.Info("Error setting up Monitor Service", "error", err)
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

	return match, fmt.Errorf("Unable to locate monitor with name '%v'", name)
}

func (service *PingdomMonitorService) GetAll() []models.Monitor {
	var monitors []models.Monitor

	checks, err := service.client.Checks.List()
	if err != nil {
		log.Info("Error received while listing checks", "error", err)
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
		log.Info(fmt.Sprintf("Error adding Monitor '%s': %v", m.Name, err.Error()))
	} else {
		log.Info("Successfully added Monitor " + m.Name)
	}
}

func (service *PingdomMonitorService) Update(m models.Monitor) {
	httpCheck := service.createHttpCheck(m)
	monitorID, _ := strconv.Atoi(m.ID)

	resp, err := service.client.Checks.Update(monitorID, &httpCheck)
	if err != nil {
		log.Info(fmt.Sprintf("Error updating Monitor '%s': %v", m.Name, err.Error()))
	} else {
		log.Info("Successfully updated Monitor "+m.Name, "response", resp.Message)
	}
}

func (service *PingdomMonitorService) Remove(m models.Monitor) {
	monitorID, _ := strconv.Atoi(m.ID)

	resp, err := service.client.Checks.Delete(monitorID)
	if err != nil {
		log.Info(fmt.Sprintf("Error deleting Monitor '%s': %v", m.Name, err.Error()))
	} else {
		log.Info("Successfully deleted Monitor "+m.Name, "response", resp.Message)
	}
}

func (service *PingdomMonitorService) createHttpCheck(monitor models.Monitor) pingdom.HttpCheck {
	httpCheck := pingdom.HttpCheck{}
	url, err := url.Parse(monitor.URL)
	if err != nil {
		log.Info(fmt.Sprintf("Error parsing url '%s' of monitor %s", service.url, monitor.Name))
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
			log.Info("Error decoding user alert contact IDs from config", "error", err)
		} else {
			httpCheck.UserIds = userIds
		}
	}

	if providerConfig != nil && len(providerConfig.AlertIntegrations) != 0 {
		integrationIdsStringArray := strings.Split(providerConfig.AlertIntegrations, "-")

		if integrationIds, err := util.SliceAtoi(integrationIdsStringArray); err != nil {
			log.Info("Error decoding integration ids into integers", "error", err)
		} else {
			httpCheck.IntegrationIds = integrationIds
		}
	}

	if providerConfig != nil && len(providerConfig.TeamAlertContacts) != 0 {
		integrationTeamIdsStringArray := strings.Split(providerConfig.TeamAlertContacts, "-")

		if integrationTeamIdsStringArray, err := util.SliceAtoi(integrationTeamIdsStringArray); err != nil {
			log.Info("Error decoding integration ids into integers", "error", err)
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
			log.Info("Error converting request headers from string to JSON", "value", providerConfig.RequestHeaders, "error", err)
		}
	}
	if providerConfig != nil && len(providerConfig.RequestHeadersEnvVar) > 0 {
		requestHeaderEnvValue := os.Getenv(providerConfig.RequestHeadersEnvVar)
		if requestHeaderEnvValue != "" {
			requestHeadersValue := make(map[string]string)
			err := json.Unmarshal([]byte(requestHeaderEnvValue), &requestHeadersValue)
			if err != nil {
				log.Info("Error converting request headers from environment from string to JSON", "envVar", providerConfig.RequestHeadersEnvVar, "error", err)
			}

			if httpCheck.RequestHeaders != nil {
				for key, value := range requestHeadersValue {
					httpCheck.RequestHeaders[key] = value
				}
			} else {
				httpCheck.RequestHeaders = requestHeadersValue
			}

		} else {
			log.Error(errors.New("error reading request headers from environment variable"), "Environment Variable does not exist", "envVar", providerConfig.RequestHeadersEnvVar)
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
		} else {
			log.Error(errors.New("error reading basic auth password from environment variable"), "Environment Variable does not exist", "envVar", providerConfig.BasicAuthUser)
		}
	}

	if providerConfig != nil && len(providerConfig.ShouldContain) > 0 {
		httpCheck.ShouldContain = providerConfig.ShouldContain
	}

	// Tags should be a single word or multiple comma-separated words
	if providerConfig != nil && len(providerConfig.Tags) > 0 {
		if !strings.Contains(providerConfig.Tags, " ") {
			httpCheck.Tags = providerConfig.Tags
		} else {
			log.Info("Tag string should not contain spaces. Not applying tags for monitor: " + httpCheck.Name)
		}
	}

	if providerConfig != nil {
		// Enable SSL validation
		httpCheck.VerifyCertificate = &providerConfig.VerifyCertificate
		// Add post data if exists
		if len(providerConfig.PostDataEnvVar) > 0 {
			postDataValue := os.Getenv(providerConfig.PostDataEnvVar)
			if postDataValue != "" {
				httpCheck.PostData = postDataValue
			} else {
				log.Error(errors.New("error reading post data from environment variable"), "Environment Variable does not exist", "envVar", providerConfig.PostDataEnvVar)
			}
		}
	}

	// Set certificate not valid before, default to 28 days to accommodate Let's Encrypt 30 day renewals + 2 days grace period.
	defaultSSLDownDaysBefore := 28
	// Pingdom doesn't allow SSLDownDaysBefore to be set if VerifyCertificate isn't set to true
	if providerConfig != nil && providerConfig.VerifyCertificate {
		if providerConfig.SSLDownDaysBefore > 0 {
			httpCheck.SSLDownDaysBefore = &providerConfig.SSLDownDaysBefore
		} else {
			httpCheck.SSLDownDaysBefore = &defaultSSLDownDaysBefore
		}
	}

	if providerConfig != nil {
		httpCheck.Paused = providerConfig.Paused
		httpCheck.NotifyWhenBackup = providerConfig.NotifyWhenBackUp
	}
}
