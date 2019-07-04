package uptimerobot

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stakater/IngressMonitorController/pkg/util"

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

	var annotations = map[string]string{
		"uptimerobot.monitor.stakater.com/interval": "600",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
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
	if "600" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("The interval is incorrect, expected: 600, but was: " + mRes.Annotations["uptimerobot.monitor.stakater.com/interval"])
	}
	service.Remove(*mRes)
}

func TestUpdateMonitorIntervalAnnotations(t *testing.T) {
	config := config.GetControllerConfig()

	service := UpTimeMonitorService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	var annotations = map[string]string{
		"uptimerobot.monitor.stakater.com/interval": "600",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
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
	if "600" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("The initial interval is incorrect: 600, but was: " + mRes.Annotations["uptimerobot.monitor.stakater.com/interval"])
	}

	mRes.URL = "https://facebook.com"
	annotations["uptimerobot.monitor.stakater.com/interval"] = "900"
	mRes.Annotations = annotations

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "https://facebook.com" {
		t.Error("The updated URL is incorrect, expected: https://facebook.com, but was: " + mRes.URL)
	}
	if "900" != mRes.Annotations["uptimerobot.monitor.stakater.com/interval"] {
		t.Error("The updated interval is incorrect, expected: 900, but was: " + mRes.Annotations["uptimerobot.monitor.stakater.com/interval"])
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

	var annotations = map[string]string{
		"uptimerobot.monitor.stakater.com/status-pages": statusPage.ID,
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
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

	var annotations = map[string]string{
		"uptimerobot.monitor.stakater.com/status-pages": statusPage.ID,
	}

	mRes.URL = "https://facebook.com"
	mRes.Annotations = annotations

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

	// Check for monitor type 'keyword'
	var annotations = map[string]string{
		"uptimerobot.monitor.stakater.com/monitor-type":   "keyword",
		"uptimerobot.monitor.stakater.com/keyword-exists": "yes",
		"uptimerobot.monitor.stakater.com/keyword-value":  "google",
	}

	m := models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("The name is incorrect, expected: " + m.Name + ", but was: " + mRes.Name)
	}

	service.Remove(*mRes)

	// Check for monitor type 'http'
	annotations = map[string]string{
		"uptimerobot.monitor.stakater.com/monitor-type": "http",
	}

	m = models.Monitor{Name: "google-test", URL: "https://google.com", Annotations: annotations}
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
