package pingdomtransaction

import (
	"testing"

	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	"gotest.tools/assert"
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

	service := PingdomTransactionMonitorService{}
	provider := util.GetProviderWithName(config, "PingdomTransaction")
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
	assert.NilError(t, err)

	defer func() {
		// Cleanup
		service.Remove(*mRes)
	}()

	if err != nil {
		t.Error("Error: " + err.Error())
	}

	assert.Equal(t, mRes.Name, m.Name)
	assert.Equal(t, mRes.URL, "https://google1.com")
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := PingdomTransactionMonitorService{}

	provider := util.GetProviderWithName(config, "Pingdom")
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error(nil, "Failed to find provider")
		return
	}
	service.Setup(*provider)

	// Create initial record
	m := models.Monitor{Name: "google-update-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-update-test")
	assert.NilError(t, err)

	defer func() {
		// Cleanup
		service.Remove(*mRes)
	}()

	// Update the record
	mRes.URL = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-update-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	assert.Equal(t, mRes.Name, m.Name)
	assert.Equal(t, mRes.URL, "https://facebook.com")

}
