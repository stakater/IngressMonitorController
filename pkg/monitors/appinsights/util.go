package appinsights

import (
	"encoding/xml"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/appinsights/mgmt/2015-05-01/insights"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

// isAlertEnabled returns true if Alertrule is required
func (aiService *AppinsightsMonitorService) isAlertEnabled() bool {
	if aiService.emailToOwners || len(aiService.emailAction) != 0 || aiService.webhookAction != "" {
		return true
	}
	return false
}

// getTags returns map[string]*string which required for resource mapping
func (aiService *AppinsightsMonitorService) getTags(tagType string, name string) map[string]*string {

	tags := make(map[string]*string)

	componentHiddenlink := fmt.Sprintf("hidden-link:/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/components/%s", aiService.subscriptionID, aiService.resourceGroup, aiService.name)
	webtestHiddenlink := fmt.Sprintf("hidden-link:/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/webtests/%s", aiService.subscriptionID, aiService.resourceGroup, name)
	value := "Resource"

	if tagType == "webtest" {
		tags[componentHiddenlink] = &value
	}

	if tagType == "alert" {

		tags[componentHiddenlink] = &value
		tags[webtestHiddenlink] = &value
	}

	return tags
}

func getURL(rawXmlData string) string {
	var w WebTest
	err := xml.Unmarshal([]byte(rawXmlData), &w)
	if err != nil {
		log.Error(err, "Failed to parse XML configuration for WebTest")
	}
	return w.Items.Request.URL
}

// getGeolocation converts slice of locations into slice of location struct
func getGeoLocation(locations []interface{}) *[]insights.WebTestGeolocation {
	var geoLocations []insights.WebTestGeolocation
	for _, v := range locations {
		l := v.(string)
		geoLocations = append(geoLocations, insights.WebTestGeolocation{
			Location: &l,
		})
	}
	return &geoLocations
}

// getConfiguration fetch values from config and return object of Configuration
func getConfiguration(monitor models.Monitor) Configuration {

	providerConfig, _ := monitor.Config.(*endpointmonitorv1alpha1.AppInsightsConfig)

	var config Configuration

	// ExpectedStatusCode is configurable via Config, Default value 200
	if providerConfig != nil && providerConfig.StatusCode > 0 {
		config.expectedStatusCode = providerConfig.StatusCode
	} else {
		config.expectedStatusCode = AppInsightsStatusCodeDefaultValue
	}

	// isRetryEnabled is configurable via config, Default value true
	if providerConfig != nil {
		config.isRetryEnabled = providerConfig.RetryEnable
	} else {
		config.isRetryEnabled = AppInsightsRetryEnabledDefaultValue
	}

	// frequency is configurable via config, Default value 300
	if providerConfig != nil && providerConfig.Frequency > 0 {
		config.frequency = int32(providerConfig.Frequency)
	} else {
		config.frequency = AppInsightsFrequencyDefaultValue
	}

	return config
}
