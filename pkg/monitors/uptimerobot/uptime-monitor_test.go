package uptimerobot

import (
	"strconv"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stakater/IngressMonitorController/pkg/util"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

// Not a test case. Cleanup to remove added dummy Monitors
func TestRemoveDanglingMonitors(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	mons, err := service.GetAllByName("google-test")

	log.Println(mons)
	if err == nil && mons == nil {
		log.Println("No Dangling Monitors")
	}
	if err == nil && mons != nil {
		for _, mon := range mons {
			service.Remove(mon)
		}
	}
}

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	// service.Setup(config.Providers[0])
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The initial name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The initial URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}

	mRes.URL = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The URL should have been updated, expected: https://facebook.com, but was: " + mRes.URL)
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithIntervalAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	configInterval := endpointmonitorv1alpha1.UptimeRobotConfig{
		Interval: 600,
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: configInterval}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}
	providerConfig, _ := mRes.Config.(*endpointmonitorv1alpha1.UptimeRobotConfig)

	if 600 != providerConfig.Interval {
		t.Error("The interval is incorrect, expected: 600, but was: " + strconv.Itoa(providerConfig.Interval))
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorIntervalAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	configInterval := endpointmonitorv1alpha1.UptimeRobotConfig{
		Interval: 600,
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: configInterval}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The initial name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The initial URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}
	providerConfig, _ := mRes.Config.(*endpointmonitorv1alpha1.UptimeRobotConfig)

	if 600 != providerConfig.Interval {
		t.Error("The interval is incorrect, expected: 600, but was: " + strconv.Itoa(providerConfig.Interval))
	}

	mRes.URL = "https://facebook.com"
	providerConfig.Interval = 900
	mRes.Config = providerConfig

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The updated URL is incorrect, expected: https://facebook.com, but was: " + mRes.URL)
	}

	providerConfig, _ = mRes.Config.(*endpointmonitorv1alpha1.UptimeRobotConfig)

	if 900 != providerConfig.Interval {
		t.Error("The interval is incorrect, expected: 600, but was: " + strconv.Itoa(providerConfig.Interval))
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithStatusPageAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	statusPageService := UpTimeStatusPageService{}
	statusPageService.Setup(config.Providers[0])

	statusPage := UpTimeStatusPage{Name: "status-page-test"}
	ID, err := statusPageService.Add(statusPage)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage.ID = ID

	configStatusPage := endpointmonitorv1alpha1.UptimeRobotConfig{
		StatusPages: statusPage.ID,
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: configStatusPage}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}
	statusPageRes, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if !util.ContainsString(statusPageRes.Monitors, mRes.ID) {
		t.Error("The status page does not contain the monitor, expected: " + mRes.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}
	service.Remove(*mRes)
	statusPageService.Remove(statusPage)
}

func TestUpdateMonitorIntervalStatusPageAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	statusPageService := UpTimeStatusPageService{}
	statusPageService.Setup(*provider)

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The initial name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The initial URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}

	statusPage := UpTimeStatusPage{Name: "status-page-test"}
	ID, err := statusPageService.Add(statusPage)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage.ID = ID

	configStatusPage := endpointmonitorv1alpha1.UptimeRobotConfig{
		StatusPages: statusPage.ID,
	}

	mRes.URL = "https://facebook.com"
	mRes.Config = configStatusPage

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The updated URL is incorrect, expected: https://facebook.com, but was: " + mRes.URL)
	}
	statusPageRes, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if !util.ContainsString(statusPageRes.Monitors, mRes.ID) {
		t.Error("The status page does not contain the monitor, expected: " + mRes.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}

	service.Remove(*mRes)
	statusPageService.Remove(statusPage)
}

func TestAddMonitorWithMonitorTypeAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	configKeyword := endpointmonitorv1alpha1.UptimeRobotConfig{
		MonitorType:   "keyword",
		KeywordExists: "yes",
		KeywordValue:  "google",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: configKeyword}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}

	service.Remove(*mRes)

	configHttpMonitor := endpointmonitorv1alpha1.UptimeRobotConfig{
		MonitorType: "http",
	}

	m = models.Monitor{Name: "google-test", URL: "https://google.com", Config: configHttpMonitor}
	service.Add(m)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}

	service.Remove(*mRes)
}

func TestAddMonitorWithIncorrectValues(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	provider.ApiKey = "dummy-api-key"
	service.Setup(*provider)

	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}

	if mRes != nil {
		t.Error("Monitor should not be added")
	}
}

func TestAddMonitorWithAlertContactsAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	configAlertContacts := endpointmonitorv1alpha1.UptimeRobotConfig{
		AlertContacts: "2628365_0_0",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Config: configAlertContacts}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}
	if mRes.URL != m.URL {
		t.Error("The URL is incorrect, expected: " + m.URL + ", but was: " + mRes.URL)
	}

	providerConfig, _ := mRes.Config.(*endpointmonitorv1alpha1.UptimeRobotConfig)

	if "2628365_0_0" != providerConfig.AlertContacts {
		t.Error("The alert-contacts is incorrect, expected: 2628365_0_0, but was: " + providerConfig.AlertContacts)
	}
	service.Remove(*mRes)
}
