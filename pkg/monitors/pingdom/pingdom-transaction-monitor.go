package pingdom

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/client/ptr"
	pingdomNew "github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/openapi"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
)

var logT = logf.Log.WithName("pingdom-transcation")

// PingdomTransactionMonitorService interfaces with MonitorService
type PingdomTransactionMonitorService struct {
	apiToken          string
	url               string
	alertContacts     string
	alertIntegrations string
	teamAlertContacts string
	client            *pingdomNew.APIClient
	context           context.Context
}

func (monitor *PingdomTransactionMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO: Retrieve oldMonitor config and compare it here
	return false
}

func (service *PingdomTransactionMonitorService) Setup(p config.Provider) {
	service.apiToken = p.ApiToken
	service.url = p.ApiURL
	service.alertContacts = p.AlertContacts
	service.alertIntegrations = p.AlertIntegrations
	service.teamAlertContacts = p.TeamAlertContacts

	cfg := pingdomNew.NewConfiguration()
	if service.apiToken == "" {
		service.apiToken = os.Getenv("PINGDOM_API_TOKEN")
	}
	cfg.SetApiToken(service.apiToken)
	service.client = pingdomNew.NewAPIClient(cfg)
	service.context = context.Background()
}

func (service *PingdomTransactionMonitorService) GetByName(name string) (*models.Monitor, error) {
	var match *models.Monitor

	monitors := service.GetAll()
	for _, mon := range monitors {
		if mon.Name == name {
			return &mon, nil
		}
	}

	return match, fmt.Errorf("Unable to locate monitor with name %v", name)
}

func (service *PingdomTransactionMonitorService) GetAll() []models.Monitor {
	var monitors []models.Monitor
	checks, _, err := service.client.TMSChecksAPI.GetAllChecks(service.context).Type_("script").Execute()
	if err != nil {
		logT.Error(err, "Error getting all transaction checks")
		return monitors
	}
	if checks == nil {
		return monitors
	}
	for _, mon := range checks.GetChecks() {
		newMon := models.Monitor{
			ID:   fmt.Sprintf("%d", mon.Id),
			Name: *mon.Name,
		}
		monitors = append(monitors, newMon)
	}

	return monitors
}

func (service *PingdomTransactionMonitorService) Add(m models.Monitor) {
	transactionCheck := service.createTransactionCheck(m)
	_, _, err := service.client.TMSChecksAPI.AddCheck(service.context).CheckWithoutID(transactionCheck).Execute()
	if err != nil {
		logT.Info("Error Adding Pingdom Transaction Monitor: " + err.Error())
	} else {
		logT.Info("Added monitor for: " + m.Name)
	}
}

func (service *PingdomTransactionMonitorService) Update(m models.Monitor) {
	transactionCheck := service.createTransactionCheck(m)
	monitorID := strToInt64(m.ID)
	_, _, err := service.client.TMSChecksAPI.ModifyCheck(context.Background(), monitorID).CheckWithoutIDPUT(*transactionCheck.AsPut()).Execute()
	if err != nil {
		logT.Info("Error updating Monitor: " + err.Error())
		return
	}
	logT.Info(fmt.Sprintf("Updated Pingdom Transaction Monitor: %s", m.Name))
}

func (service *PingdomTransactionMonitorService) Remove(m models.Monitor) {
	_, resp, err := service.client.TMSChecksAPI.DeleteCheck(context.Background(), strToInt64(m.ID)).Execute()
	if err != nil {
		logT.Info("Error deleting Monitor: " + err.Error())
	} else {
		logT.Info(fmt.Sprintf("Delete Monitor: %v", resp))
	}
}

func (service *PingdomTransactionMonitorService) createTransactionCheck(monitor models.Monitor) pingdomNew.CheckWithoutID {
	transactionCheck := pingdomNew.CheckWithoutID{}
	transactionCheck.Name = monitor.Name

	userIds := parseIDs(service.alertContacts)
	if userIds != nil {
		transactionCheck.ContactIds = userIds
	}
	integrationIds := parseIDs(service.alertIntegrations)
	if integrationIds != nil {
		transactionCheck.IntegrationIds = integrationIds
	}
	teamAlertContacts := parseIDs(service.teamAlertContacts)
	if teamAlertContacts != nil {
		transactionCheck.TeamIds = teamAlertContacts
	}
	service.addConfigToHttpCheck(&transactionCheck, monitor.Config)

	return transactionCheck
}

func (service *PingdomTransactionMonitorService) addConfigToHttpCheck(transactionCheck *pingdomNew.CheckWithoutID, config interface{}) {

	// Retrieve provider configuration
	providerConfig, _ := config.(*endpointmonitorv1alpha1.PingdomTransactionConfig)

	if providerConfig == nil {
		logT.Info("Error retrieving provider configuration for Pingdom Transaction Monitor")
		return
	}

	// Set contact ids
	userIds := parseIDs(providerConfig.AlertContacts)
	if userIds != nil {
		transactionCheck.ContactIds = userIds
	}
	teamAlertContacts := parseIDs(providerConfig.TeamAlertContacts)
	if teamAlertContacts != nil {
		transactionCheck.TeamIds = teamAlertContacts
	}
	integrationIds := parseIDs(providerConfig.AlertIntegrations)
	if integrationIds != nil {
		transactionCheck.IntegrationIds = integrationIds
	}

	// Set transaction check configuration
	if providerConfig.CustomMessage != "" {
		transactionCheck.CustomMessage = ptr.String(providerConfig.CustomMessage)
	}
	if providerConfig.Region != "" {
		transactionCheck.Region = ptr.String(providerConfig.Region)
	}
	if providerConfig.SendNotificationWhenDown > 0 {
		transactionCheck.SendNotificationWhenDown = ptr.Int64(providerConfig.SendNotificationWhenDown)
	}
	if providerConfig.Paused {
		transactionCheck.Active = ptr.Bool(!providerConfig.Paused)
	}
	if len(providerConfig.Tags) > 0 {
		transactionCheck.Tags = providerConfig.Tags
	}
}

// parseIDs splits a string of IDs into an array of integers
func parseIDs(field string) []int64 {
	if len(field) > 0 {
		stringArray := strings.Split(field, "-")
		ids, err := util.SliceAtoi64(stringArray)
		if err != nil {
			logT.Info("Error decoding ids into integers" + err.Error())
			return nil
		}
		return ids
	}
	return nil
}

func strToInt64(str string) int64 {
	// Parse the string as a base-10 integer with a bit size of 64.
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		logT.Error(err, "Error converting string to int64")
		return 0
	}
	return value
}
