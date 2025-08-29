package statuscake

import (
	"os"
	"testing"
	"time"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
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

	service := StatusCakeMonitorService{}
	provider := util.GetProviderWithName(config, "StatusCake")
	if provider == nil {
		return
	}
	service.Setup(*provider)
	m := models.Monitor{Name: "google-test", URL: "https://google1.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-test")

	if err != nil {
		t.Error("Error: " + err.Error())
	} else if mRes == nil {
		t.Errorf("Found empty response for Monitor. Name: %s and URL: %s", m.Name, m.URL)
	}
	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Error("URL and name should be the same")
	}
	service.Remove(*mRes)

	time.Sleep(5 * time.Second)

	monitor, err := service.GetByName(mRes.Name)

	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := StatusCakeMonitorService{}

	provider := util.GetProviderWithName(config, "StatusCake")
	if provider == nil {
		return
	}
	service.Setup(*provider)

	m := models.Monitor{Name: "google-test-statuscake", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName(m.Name)

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Error("URL and name should be the same")
	}

	mRes.Name = "google-test-statuscake-updated"

	service.Update(*mRes)

	mRes, err = service.GetByID(mRes.ID)

	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != "google-test-statuscake-updated" {
		t.Error("Name and ID should be the same")
	}

	time.Sleep(5 * time.Second)
	service.Remove(*mRes)

	monitor, err := service.GetByName(mRes.Name)

	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}

func TestBuildUpsertForm(t *testing.T) {
	m := models.Monitor{Name: "google-test", URL: "https://google.com"}

	monitorConfig := &endpointmonitorv1alpha1.StatusCakeConfig{
		CheckRate:      60,
		TestType:       "TCP",
		Paused:         true,
		PingURL:        "",
		FollowRedirect: true,
		Port:           7070,
		TriggerRate:    1,
		BasicAuthUser:  "testuser",
		Confirmation:   2,
		EnableSSLAlert: true,
		FindString:     "",
                Timeout:        30,

		// changed to string array type on statuscake api
		// TODO: release new apiVersion to cater new type in apiVersion struct
		ContactGroup: "123456,654321",
		TestTags:     "test,testrun,uptime",
		StatusCodes:  "500,501,502,503,504,505",
	}
	m.Config = monitorConfig

	oldEnv := os.Getenv("testuser")
	os.Setenv("testuser", "testpass")
	defer os.Setenv("testuser", oldEnv)

	vals := buildUpsertForm(m, "")
	assert.Equal(t, "testuser", vals.Get("basic_username"))
	assert.Equal(t, "testpass", vals.Get("basic_password"))
	assert.Equal(t, "60", vals.Get("check_rate"))
	assert.Equal(t, "2", vals.Get("confirmation"))
	assert.Equal(t, "123456,654321", convertUrlValuesToString(vals, "contact_groups[]"))
	assert.Equal(t, "1", vals.Get("enable_ssl_alert"))
	assert.Equal(t, "", vals.Get("find_string"))
	assert.Equal(t, "1", vals.Get("follow_redirects"))
	assert.Equal(t, "1", vals.Get("paused"))
	assert.Equal(t, "", vals.Get("ping_url"))
	assert.Equal(t, "7070", vals.Get("port"))
	assert.Equal(t, "500,501,502,503,504,505", vals.Get("status_codes_csv"))
	assert.Equal(t, "test,testrun,uptime", convertUrlValuesToString(vals, "tags[]"))
	assert.Equal(t, "TCP", vals.Get("test_type"))
	assert.Equal(t, "1", vals.Get("trigger_rate"))
	assert.Equal(t, "30", vals.Get("timeout"))
}
