package statuscake

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
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

func TestAddHeartbeatMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := StatusCakeMonitorService{}
	provider := util.GetProviderWithName(config, "StatusCake")
	if provider == nil {
		return
	}
	service.Setup(*provider)

	m := models.Monitor{
		Name: "heartbeat-imc-test",
		Config: &endpointmonitorv1alpha1.StatusCakeConfig{
			TestType:  "Heartbeat",
			CheckRate: 300,
			TestTags:  "imc-test",
		},
	}
	service.Add(m)

	mRes, err := service.GetByName(m.Name)
	if err != nil {
		t.Error("Error: " + err.Error())
	} else if mRes == nil {
		t.Errorf("Found empty response for Monitor. Name: %s", m.Name)
	}
	if mRes.Name != m.Name {
		t.Error("Name should be the same")
	}

	service.Remove(*mRes)

	time.Sleep(5 * time.Second)

	monitor, err := service.GetByName(mRes.Name)
	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}

func TestUpdateHeartbeatMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := StatusCakeMonitorService{}
	provider := util.GetProviderWithName(config, "StatusCake")
	if provider == nil {
		return
	}
	service.Setup(*provider)

	m := models.Monitor{
		Name: "heartbeat-imc-test",
		Config: &endpointmonitorv1alpha1.StatusCakeConfig{
			TestType:  "Heartbeat",
			CheckRate: 300,
			TestTags:  "imc-test",
		},
	}
	service.Add(m)

	mRes, err := service.GetByName(m.Name)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != m.Name {
		t.Error("Name should be the same")
	}

	updatedName := m.Name + "-updated"
	t.Cleanup(func() {
		for _, name := range []string{m.Name, updatedName} {
			if existing, _ := service.GetByName(name); existing != nil {
				service.Remove(*existing)
			}
		}
	})

	mRes.Name = updatedName
	service.Update(*mRes)

	mRes, err = service.GetHeartbeatByID(mRes.ID)
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes.Name != updatedName {
		t.Error("Name should be updated")
	}

	time.Sleep(5 * time.Second)
	service.Remove(*mRes)

	monitor, err := service.GetByName(mRes.Name)
	if monitor != nil {
		t.Error("Monitor should've been deleted ", monitor, err)
	}
}

func TestBuildHeartbeatForm(t *testing.T) {
	tests := []struct {
		name           string
		monitor        models.Monitor
		cgroup         string
		expectedPeriod string
		expectedCgroup string
		expectedName   string
		expectedTags   string
		expectedPaused string
		absentFields   []string
	}{
		{
			name: "full config",
			monitor: models.Monitor{
				Name: "heartbeat-test",
				Config: &endpointmonitorv1alpha1.StatusCakeConfig{
					TestType:     "Heartbeat",
					CheckRate:    60,
					ContactGroup: "123456,654321",
					TestTags:     "prod,heartbeat",
					Paused:       true,
				},
			},
			expectedPeriod: "60",
			expectedCgroup: "123456,654321",
			expectedName:   "heartbeat-test",
			expectedTags:   "prod,heartbeat",
			expectedPaused: "1",
			absentFields:   []string{"website_url", "test_type", "status_codes_csv"},
		},
		{
			name:           "defaults when CheckRate unset",
			monitor:        models.Monitor{Name: "heartbeat-defaults", Config: &endpointmonitorv1alpha1.StatusCakeConfig{TestType: "Heartbeat"}},
			cgroup:         "fallback-group",
			expectedPeriod: "300",
			expectedCgroup: "fallback-group",
		},
		{
			name:           "out of range: below minimum (29)",
			monitor:        models.Monitor{Name: "heartbeat-bad-rate", Config: &endpointmonitorv1alpha1.StatusCakeConfig{TestType: "Heartbeat", CheckRate: 29}},
			expectedPeriod: "300",
		},
		{
			name:           "out of range: above maximum (172801)",
			monitor:        models.Monitor{Name: "heartbeat-bad-rate", Config: &endpointmonitorv1alpha1.StatusCakeConfig{TestType: "Heartbeat", CheckRate: 172801}},
			expectedPeriod: "300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := buildHeartbeatForm(tt.monitor, tt.cgroup)
			assert.Equal(t, tt.expectedPeriod, vals.Get("period"))
			if tt.expectedCgroup != "" {
				assert.Equal(t, tt.expectedCgroup, convertUrlValuesToString(vals, "contact_groups[]"))
			}
			if tt.expectedName != "" {
				assert.Equal(t, tt.expectedName, vals.Get("name"))
			}
			if tt.expectedTags != "" {
				assert.Equal(t, tt.expectedTags, convertUrlValuesToString(vals, "tags[]"))
			}
			if tt.expectedPaused != "" {
				assert.Equal(t, tt.expectedPaused, vals.Get("paused"))
			}
			for _, field := range tt.absentFields {
				assert.Equal(t, "", vals.Get(field))
			}
		})
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

const (
	testMonitorName = "google-test"
	testMonitorID   = "12345"
	testMonitorURL  = "https://google.com"
)

func uptimeListJSON() string {
	return `{"data":[{"id":"` + testMonitorID + `","name":"` + testMonitorName + `","website_url":"` + testMonitorURL + `","tags":[]}],"metadata":{"page":1,"per_page":100,"page_count":1,"total_count":1}}`
}

func heartbeatEmptyListJSON() string {
	return `{"data":[],"metadata":{"page":1,"per_page":100,"page_count":1,"total_count":0}}`
}

func uptimeByIDJSON() string {
	return `{"data":{"id":"` + testMonitorID + `","name":"` + testMonitorName + `","website_url":"` + testMonitorURL + `","tags":[]}}`
}

func newTestService(t *testing.T, handler http.HandlerFunc) *StatusCakeMonitorService {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	service := &StatusCakeMonitorService{}
	service.Setup(config.Provider{ApiKey: "test-key", ApiURL: server.URL})
	service.client = server.Client()
	return service
}

func writeBody(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("write response: %v", err)
	}
}

func TestListRequestsIncludeNouptime(t *testing.T) {
	var sawUptime, sawHeartbeat bool
	service := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("nouptime"); got != "true" {
			t.Errorf("expected nouptime=true on %s, got %q", r.URL.String(), got)
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/v1/uptime"):
			sawUptime = true
			writeBody(t, w, uptimeListJSON())
		case strings.HasPrefix(r.URL.Path, "/v1/heartbeat"):
			sawHeartbeat = true
			writeBody(t, w, heartbeatEmptyListJSON())
		default:
			http.NotFound(w, r)
		}
	})

	monitors, err := service.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if !sawUptime || !sawHeartbeat {
		t.Fatalf("expected both uptime and heartbeat list calls, uptime=%v heartbeat=%v", sawUptime, sawHeartbeat)
	}
	if len(monitors) != 1 || monitors[0].Name != testMonitorName {
		t.Fatalf("unexpected monitors: %+v", monitors)
	}
}

func TestGetByNameRetriesHTTP500(t *testing.T) {
	var uptimeListCalls atomic.Int32
	service := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/heartbeat") {
			writeBody(t, w, heartbeatEmptyListJSON())
			return
		}
		if r.URL.Path == "/v1/uptime/" || r.URL.Path == "/v1/uptime" {
			if uptimeListCalls.Add(1) == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			writeBody(t, w, uptimeListJSON())
			return
		}
		http.NotFound(w, r)
	})

	monitor, err := service.GetByName(testMonitorName)
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if monitor == nil || monitor.ID != testMonitorID {
		t.Fatalf("unexpected monitor: %+v", monitor)
	}
	if uptimeListCalls.Load() < 2 {
		t.Fatalf("expected retry after 500, got %d list calls", uptimeListCalls.Load())
	}
}

func TestGetByNameUsesCachedIDOnSecondLookup(t *testing.T) {
	var listCalls, getByIDCalls atomic.Int32
	service := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/uptime/"+testMonitorID:
			getByIDCalls.Add(1)
			writeBody(t, w, uptimeByIDJSON())
		case r.URL.Path == "/v1/uptime/" || r.URL.Path == "/v1/uptime":
			listCalls.Add(1)
			writeBody(t, w, uptimeListJSON())
		case strings.HasPrefix(r.URL.Path, "/v1/heartbeat"):
			writeBody(t, w, heartbeatEmptyListJSON())
		default:
			http.NotFound(w, r)
		}
	})

	first, err := service.GetByName(testMonitorName)
	if err != nil {
		t.Fatalf("first GetByName: %v", err)
	}
	if first == nil || first.ID != testMonitorID {
		t.Fatalf("unexpected first monitor: %+v", first)
	}

	second, err := service.GetByName(testMonitorName)
	if err != nil {
		t.Fatalf("second GetByName: %v", err)
	}
	if second == nil || second.ID != testMonitorID {
		t.Fatalf("unexpected second monitor: %+v", second)
	}
	if listCalls.Load() != 1 {
		t.Fatalf("expected one list call, got %d", listCalls.Load())
	}
	if getByIDCalls.Load() != 1 {
		t.Fatalf("expected one get-by-id call, got %d", getByIDCalls.Load())
	}
}

func TestGetByNameFallsBackToListAfterCachedID404(t *testing.T) {
	var listCalls, getByIDCalls atomic.Int32
	service := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/uptime/"+testMonitorID:
			getByIDCalls.Add(1)
			w.WriteHeader(http.StatusNotFound)
		case r.URL.Path == "/v1/uptime/" || r.URL.Path == "/v1/uptime":
			listCalls.Add(1)
			writeBody(t, w, uptimeListJSON())
		case strings.HasPrefix(r.URL.Path, "/v1/heartbeat"):
			writeBody(t, w, heartbeatEmptyListJSON())
		default:
			http.NotFound(w, r)
		}
	})

	first, err := service.GetByName(testMonitorName)
	if err != nil {
		t.Fatalf("first GetByName: %v", err)
	}
	if first == nil || first.ID != testMonitorID {
		t.Fatalf("unexpected first monitor: %+v", first)
	}

	second, err := service.GetByName(testMonitorName)
	if err != nil {
		t.Fatalf("second GetByName: %v", err)
	}
	if second == nil || second.ID != testMonitorID {
		t.Fatalf("unexpected second monitor: %+v", second)
	}
	if getByIDCalls.Load() != 1 {
		t.Fatalf("expected one get-by-id call, got %d", getByIDCalls.Load())
	}
	if listCalls.Load() != 2 {
		t.Fatalf("expected list to run again after 404, got %d list calls", listCalls.Load())
	}
}

func TestAddRetries429PreservesFormBody(t *testing.T) {
	var posts atomic.Int32
	var retryBody string
	service := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && (r.URL.Path == "/v1/uptime" || r.URL.Path == "/v1/uptime/") {
			n := posts.Add(1)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read body: %v", err)
			}
			if n == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			retryBody = string(body)
			if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
				t.Errorf("expected form content type, got %q", ct)
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.NotFound(w, r)
	})

	service.Add(models.Monitor{
		Name: "example-http-check",
		URL:  "https://example.com",
		Config: &endpointmonitorv1alpha1.StatusCakeConfig{
			TestType:  "HTTP",
			CheckRate: 300,
		},
	})

	if posts.Load() != 2 {
		t.Fatalf("expected POST then retry, got %d", posts.Load())
	}
	if !strings.Contains(retryBody, "name=example-http-check") {
		t.Fatalf("retry body missing name: %q", retryBody)
	}
	if !strings.Contains(retryBody, "website_url=https") {
		t.Fatalf("retry body missing website_url: %q", retryBody)
	}
	if !strings.Contains(retryBody, "check_rate=300") {
		t.Fatalf("retry body missing check_rate: %q", retryBody)
	}
}
