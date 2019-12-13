package appinsights

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/appinsights/mgmt/2015-05-01/insights"
	log "github.com/sirupsen/logrus"
	"github.com/stakater/IngressMonitorController/pkg/models"
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
		log.Errorf("Failed to parse XML configuration for WebTest")
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

// getAnnotation fetch values from annotations and return object of Annotations
func getAnnotation(monitor models.Monitor) Annotation {

	var anno Annotation
	// ExpectedStatusCode is configurable via annotation, Default value 200
	if val, ok := monitor.Annotations[AppInsightsStatusCodeAnnotation]; ok {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			log.Errorf("Error fetching isRetryEnabled from annotation, setting it to default value: %v", err)
			anno.expectedStatusCode = AppInsightsStatusCodeAnnotationDefaultValue
		}
		anno.expectedStatusCode = intVal
	} else {
		anno.expectedStatusCode = AppInsightsStatusCodeAnnotationDefaultValue
	}

	// isRetryEnabled is configurable via annotation, Default value true
	if val, ok := monitor.Annotations[AppInsightsRetryEnabledAnnotation]; ok {
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			log.Errorf("Error fetching isRetryEnabled from annotation, setting it to default value: %v", err)
			anno.isRetryEnabled = AppInsightsRetryEnabledAnnotationDefaultValue
		}
		anno.isRetryEnabled = boolVal
	} else {
		anno.isRetryEnabled = AppInsightsRetryEnabledAnnotationDefaultValue
	}

	// frequency is configurable via annotation, Default value 300
	if val, ok := monitor.Annotations[AppInsightsFrequency]; ok {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			log.Errorf("Error fetching frequency from annotation, setting it to default value: %v", err)
			anno.frequency = AppInsightsFrequencyDefaultValue
		}
		anno.frequency = int32(intVal)
	} else {
		anno.frequency = AppInsightsFrequencyDefaultValue
	}

	return anno
}
