package uptimerobot

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

// Not a test case. Cleanup to remove added dummy StatusPages
func TestRemoveDanglingStatusPages(t *testing.T) {
	config := config.GetControllerConfig()
	service := UpTimeStatusPageService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	statusPages, err := service.GetAllStatusPages("status-page-test")

	if err == nil && statusPages == nil {
		log.Println("No dangling StatusPages named: status-page-test")
	}
	if err != nil && statusPages != nil {
		for _, statusPage := range statusPages {
			service.Remove(statusPage)
		}
	}

	statusPages1, err := service.GetAllStatusPages("status-page-test-1")

	if err == nil && statusPages1 == nil {
		log.Println("No dangling StatusPages named: status-page-test-1")
	}
	if err != nil && statusPages1 != nil {
		for _, statusPage := range statusPages1 {
			service.Remove(statusPage)
		}
	}

	statusPages2, err := service.GetAllStatusPages("status-page-test-2")

	if err == nil && statusPages2 == nil {
		log.Println("No dangling StatusPages named: status-page-test-2")
	}
	if err != nil && statusPages2 != nil {
		for _, statusPage := range statusPages2 {
			service.Remove(statusPage)
		}
	}

	statusPages3, err := service.GetAllStatusPages("status-page-test-3")

	if err == nil && statusPages3 == nil {
		log.Println("No dangling StatusPages named: status-page-test-3")
	}
	if err == nil && statusPages3 != nil {
		for _, statusPage := range statusPages3 {
			service.Remove(statusPage)
		}
	}

}

func TestAddMonitorMultipleTimesToStatusPage(t *testing.T) {
	config := config.GetControllerConfig()
	service := UpTimeStatusPageService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	statusPage := UpTimeStatusPage{Name: "status-page-test"}
	ID, err := service.Add(statusPage)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage.ID = ID

	monitorService := UpTimeMonitorService{}
	provider = util.GetProviderWithName(config, "UptimeRobot")
	monitorService.Setup(*provider)

	monitor := models.Monitor{Name: "google-test", URL: "https://google.com"}
	monitorService.Add(monitor)

	monitorRes, err := monitorService.GetByName("google-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	service.AddMonitorToStatusPage(statusPage, *monitorRes)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	statusPageRes, err := service.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if !util.ContainsString(statusPageRes.Monitors, monitorRes.ID) {
		t.Error("The status page does not contain the monitor, expected: " + monitorRes.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}

	service.AddMonitorToStatusPage(statusPage, *monitorRes)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	statusPageRes, err = service.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if !util.ContainsString(statusPageRes.Monitors, monitorRes.ID) {
		t.Error("The status page does not contain the monitor, expected: " + monitorRes.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}

	// Tidy up
	monitorService.Remove(*monitorRes)
	service.Remove(statusPage)
}

func TestAddMultipleMonitorsToStatusPage(t *testing.T) {
	config := config.GetControllerConfig()
	service := UpTimeStatusPageService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	statusPage := UpTimeStatusPage{Name: "status-page-test"}
	ID, err := service.Add(statusPage)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage.ID = ID

	monitorService := UpTimeMonitorService{}
	provider = util.GetProviderWithName(config, "UptimeRobot")
	monitorService.Setup(*provider)
	monitor1 := models.Monitor{Name: "google-test-1", URL: "https://google.com"}
	monitorService.Add(monitor1)

	monitor1Res, err := monitorService.GetByName("google-test-1")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	service.AddMonitorToStatusPage(statusPage, *monitor1Res)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	monitor2 := models.Monitor{Name: "google-test-2", URL: "https://google.co.uk"}
	monitorService.Add(monitor2)

	monitor2Res, err := monitorService.GetByName("google-test-2")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	service.AddMonitorToStatusPage(statusPage, *monitor2Res)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	statusPageRes, err := service.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if !util.ContainsString(statusPageRes.Monitors, monitor1Res.ID) {
		t.Error("The status page does not contain the first monitor, expected: " + monitor1Res.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}
	if !util.ContainsString(statusPageRes.Monitors, monitor2Res.ID) {
		t.Error("The status page does not contain the second monitor, expected: " + monitor2Res.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}

	// Tidy up
	monitorService.Remove(*monitor1Res)
	monitorService.Remove(*monitor2Res)
	service.Remove(statusPage)
}

func TestGetStatusPagesForMonitor(t *testing.T) {
	config := config.GetControllerConfig()
	service := UpTimeStatusPageService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	statusPage1 := UpTimeStatusPage{Name: "status-page-test-1"}
	ID1, err := service.Add(statusPage1)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage1.ID = ID1

	statusPage2 := UpTimeStatusPage{Name: "status-page-test-2"}
	ID2, err := service.Add(statusPage2)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage2.ID = ID2

	statusPage3 := UpTimeStatusPage{Name: "status-page-test-3"}
	ID3, err := service.Add(statusPage3)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage3.ID = ID3

	monitorService := UpTimeMonitorService{}
	provider = util.GetProviderWithName(config, "UptimeRobot")
	monitorService.Setup(*provider)
	monitor := models.Monitor{Name: "google-test", URL: "https://google.com"}
	monitorService.Add(monitor)

	monitorRes, err := monitorService.GetByName("google-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	service.AddMonitorToStatusPage(statusPage1, *monitorRes)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	service.AddMonitorToStatusPage(statusPage2, *monitorRes)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	statusPageIds, err := service.GetStatusPagesForMonitor(monitorRes.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if !util.ContainsString(statusPageIds, statusPage1.ID) {
		t.Error("The first status page does not contain the monitor, expected: " + statusPage1.ID + ", but was: " + strings.Join(statusPageIds, "-"))
	}
	if !util.ContainsString(statusPageIds, statusPage2.ID) {
		t.Error("The second status page does not contain the monitor, expected: " + statusPage2.ID + ", but was: " + strings.Join(statusPageIds, "-"))
	}

	if util.ContainsString(statusPageIds, statusPage3.ID) {
		t.Error("The third status page should not contain the monitor, but was: " + strings.Join(statusPageIds, "-"))
	}

	// Tidy up
	monitorService.Remove(*monitorRes)
	service.Remove(statusPage1)
	service.Remove(statusPage2)
	service.Remove(statusPage3)
}

func TestRemoveMonitorFromStatusPage(t *testing.T) {
	config := config.GetControllerConfig()
	service := UpTimeStatusPageService{}
	provider := util.GetProviderWithName(config, "UptimeRobot")
	service.Setup(*provider)

	statusPage := UpTimeStatusPage{Name: "status-page-test"}
	ID, err := service.Add(statusPage)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	statusPage.ID = ID

	monitorService := UpTimeMonitorService{}
	provider = util.GetProviderWithName(config, "UptimeRobot")
	monitorService.Setup(*provider)
	monitor := models.Monitor{Name: "google-test", URL: "https://google.com"}
	monitorService.Add(monitor)

	monitorRes, err := monitorService.GetByName("google-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	_, err = service.AddMonitorToStatusPage(statusPage, *monitorRes)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	statusPageRes, err := service.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if !util.ContainsString(statusPageRes.Monitors, monitorRes.ID) {
		t.Error("The status page does not contain the monitor, expected: " + monitorRes.ID + ", but was: " + strings.Join(statusPageRes.Monitors, "-"))
	}

	_, err = service.RemoveMonitorFromStatusPage(statusPage, *monitorRes)
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	statusPageRes, err = service.Get(statusPage.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if util.ContainsString(statusPageRes.Monitors, monitorRes.ID) {
		t.Error("The status page still contains the monitor, but shouldn't, monitors: " + strings.Join(statusPageRes.Monitors, "-"))
	}

	// Tidy up
	monitorService.Remove(*monitorRes)
	service.Remove(statusPage)
}
