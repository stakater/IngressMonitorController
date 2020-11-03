package pingdom

import (
	log "github.com/sirupsen/logrus"
	"net/url"
	"testing"
	"time"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
)

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := PingdomMonitorService{}
	provider := util.GetProviderWithName(config, "Pingdom")
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error("Failed to find provider")
		return
	}
	service.Setup(*provider)
	urlToTest := "https://google1.com"
	parsedUrl, err := url.Parse(urlToTest)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	m := models.Monitor{Name: "google-test", URL: urlToTest}
	service.Add(m)
	//Creation delay
	time.Sleep(5 * time.Second)
	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	// mRes.URL is a domain name, without scheme, so we parsed URL previously
	if mRes.Name != m.Name || mRes.URL != parsedUrl.Host {
		t.Errorf("URL and name should be the same")
	}
	service.Remove(*mRes)
	monitor, err := service.GetByName(mRes.Name)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor)
	}
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := PingdomMonitorService{}

	provider := util.GetProviderWithName(config, "Pingdom")
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error("Failed to find provider")
		return
	}
	service.Setup(*provider)
	urlToTest := "https://google.com"
	parsedUrl, err := url.Parse(urlToTest)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	m := models.Monitor{Name: "google-test", URL: urlToTest}
	service.Add(m)
	//Creation delay
	time.Sleep(5 * time.Second)
	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name || mRes.URL != parsedUrl.Host {
		t.Error("URL and name should be the same")
	}

	mRes.URL = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.URL != "facebook.com" {
		t.Error("URL and name should be the same")
	}

	service.Remove(*mRes)

	monitor, err := service.GetByName(mRes.Name)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor)
	}
}
