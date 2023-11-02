package pingdom

import (
	"net/url"
	"testing"

	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// To allow normal logging to be printed if tests fails
	// Dev mode is an extra feature to make output more readable
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := PingdomMonitorService{}
	provider := util.GetProviderWithName(config, "Pingdom")
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error(nil, "Failed to find provider")
		return
	}
	service.Setup(*provider)
	m := models.Monitor{Name: "google-test", URL: "https://google1.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}

	// Pingdom returns the hostname only without prefix
	mURL, _ := url.Parse(m.URL)
	if mRes.Name != m.Name || mRes.URL != mURL.Host {
		t.Errorf("URL and name should be the same. request: %+v response: %+v", m, mRes)
	}

	// Cleanup
	service.Remove(*mRes)
	monitor, err := service.GetByName(mRes.Name)
	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := PingdomMonitorService{}

	provider := util.GetProviderWithName(config, "Pingdom")
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error(nil, "Failed to find provider")
		return
	}
	service.Setup(*provider)

	// Create initial record
	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	}

	// Pingdom returns the hostname only without prefix
	mURL, _ := url.Parse(m.URL)
	if mRes.Name != m.Name || mRes.URL != mURL.Host {
		t.Errorf("URL and name should be the same. request: %+v response: %+v", m, mRes)
	}

	// Update the record
	mRes.URL = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	if mRes.Name != m.Name || mRes.URL != "facebook.com" {
		t.Errorf("URL and name should be the same. request: %+v response: %+v", m, mRes)
	}

	// Cleanup
	service.Remove(*mRes)
	monitor, err := service.GetByName(mRes.Name)
	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}
