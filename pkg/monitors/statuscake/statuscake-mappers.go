package statuscake

import (
	"strings"

	statuscake "github.com/StatusCakeDev/statuscake-go"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

// StatusCakeMonitorMonitorToBaseMonitorMapper maps StatusCakeMonitorData to the internal Monitor model
func StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeData StatusCakeMonitorData) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.WebsiteName
	m.URL = statuscakeData.WebsiteURL
	m.ID = statuscakeData.TestID

	var providerConfig endpointmonitorv1alpha1.StatusCakeConfig
	providerConfig.Paused = statuscakeData.Paused
	providerConfig.TestType = statuscakeData.TestType
	providerConfig.CheckRate = statuscakeData.CheckRate
	providerConfig.ContactGroup = strings.Join(statuscakeData.ContactGroup, ",")
	providerConfig.TestTags = strings.Join(statuscakeData.Tags, ",")
	providerConfig.FollowRedirect = statuscakeData.FollowRedirect
	providerConfig.Port = statuscakeData.Port
	providerConfig.TriggerRate = statuscakeData.TriggerRate
	providerConfig.FindString = statuscakeData.FindString
	providerConfig.EnableSSLAlert = statuscakeData.EnableSSLAlert
	providerConfig.Confirmation = int(statuscakeData.Confirmation)

	m.Config = &providerConfig
	return &m
}

// StatusCakeApiResponseDataToBaseMonitorMapper maps statuscake.UptimeTestResponse to the internal Monitor model
func StatusCakeApiResponseDataToBaseMonitorMapper(statuscakeData statuscake.UptimeTestResponse) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.Data.Name
	m.URL = statuscakeData.Data.WebsiteURL
	m.ID = statuscakeData.Data.ID

	var providerConfig endpointmonitorv1alpha1.StatusCakeConfig
	providerConfig.Paused = statuscakeData.Data.Paused
	providerConfig.TestType = string(statuscakeData.Data.TestType)
	providerConfig.CheckRate = int(statuscakeData.Data.CheckRate)
	providerConfig.ContactGroup = strings.Join(statuscakeData.Data.ContactGroups, ",")
	providerConfig.TestTags = strings.Join(statuscakeData.Data.Tags, ",")
	providerConfig.FollowRedirect = statuscakeData.Data.FollowRedirects
	providerConfig.EnableSSLAlert = statuscakeData.Data.EnableSSLAlert
	providerConfig.Confirmation = int(statuscakeData.Data.Confirmation)
	if statuscakeData.Data.Port != nil {
		providerConfig.Port = int(*statuscakeData.Data.Port)
	}
	providerConfig.TriggerRate = int(statuscakeData.Data.TriggerRate)
	if statuscakeData.Data.FindString != nil {
		providerConfig.FindString = *statuscakeData.Data.FindString
	}

	m.Config = &providerConfig
	return &m
}

// StatusCakeMonitorMonitorsToBaseMonitorsMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorsToBaseMonitorsMapper(statuscakeData []StatusCakeMonitorData) []models.Monitor {
	var monitors []models.Monitor
	for _, payloadData := range statuscakeData {
		monitors = append(monitors, *StatusCakeMonitorMonitorToBaseMonitorMapper(payloadData))
	}
	return monitors
}
