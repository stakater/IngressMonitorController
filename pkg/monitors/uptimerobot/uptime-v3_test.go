package uptimerobot

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

// recordedRequest captures what the fake v3 API received
type recordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   map[string]interface{}
}

type recorder struct {
	requests []recordedRequest
}

// handler records every request and delegates responding to the test
func (r *recorder) handler(t *testing.T, respond func(w http.ResponseWriter, req *recordedRequest)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		parsed := map[string]interface{}{}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &parsed); err != nil {
				t.Errorf("request body is not valid JSON: %v: %s", err, string(body))
			}
		}
		recorded := recordedRequest{Method: req.Method, Path: req.URL.Path, Query: req.URL.Query(), Header: req.Header, Body: parsed}
		r.requests = append(r.requests, recorded)
		respond(w, &recorded)
	}
}

func newTestService(serverURL string, alertContacts string) UpTimeMonitorService {
	service := UpTimeMonitorService{}
	service.Setup(config.Provider{
		Name:          "UptimeRobot",
		ApiKey:        "test-api-key",
		ApiURL:        serverURL,
		AlertContacts: alertContacts,
	})
	return service
}

func writeJSON(w http.ResponseWriter, statusCode int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	io.WriteString(w, body)
}

func TestV3AddMonitorHttpRequest(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusCreated, `{"id":42,"friendlyName":"google-test","url":"https://google.com","type":"HTTP","interval":300}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "111_0_0-222_5_30")
	service.Add(models.Monitor{Name: "google-test", URL: "https://google.com"})

	if len(rec.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(rec.requests))
	}
	req := rec.requests[0]
	if req.Method != "POST" || req.Path != "/monitors" {
		t.Errorf("expected POST /monitors, got %s %s", req.Method, req.Path)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer test-api-key" {
		t.Errorf("expected Bearer auth header, got %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("expected JSON content type, got %q", got)
	}
	if req.Body["type"] != "HTTP" {
		t.Errorf("expected type HTTP, got %v", req.Body["type"])
	}
	if req.Body["friendlyName"] != "google-test" {
		t.Errorf("expected friendlyName google-test, got %v", req.Body["friendlyName"])
	}
	if req.Body["url"] != "https://google.com" {
		t.Errorf("expected url, got %v", req.Body["url"])
	}
	if req.Body["interval"] != float64(300) {
		t.Errorf("expected default interval 300, got %v", req.Body["interval"])
	}
	if req.Body["timeout"] != float64(30) {
		t.Errorf("expected timeout 30, got %v", req.Body["timeout"])
	}
	contacts, ok := req.Body["assignedAlertContacts"].([]interface{})
	if !ok || len(contacts) != 2 {
		t.Fatalf("expected 2 assignedAlertContacts, got %v", req.Body["assignedAlertContacts"])
	}
	first := contacts[0].(map[string]interface{})
	if first["alertContactId"] != float64(111) || first["threshold"] != float64(0) || first["recurrence"] != float64(0) {
		t.Errorf("first alert contact parsed incorrectly: %v", first)
	}
	second := contacts[1].(map[string]interface{})
	if second["alertContactId"] != float64(222) || second["threshold"] != float64(5) || second["recurrence"] != float64(30) {
		t.Errorf("second alert contact parsed incorrectly: %v", second)
	}
}

func TestV3AddMonitorKeywordRequest(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusCreated, `{"id":43,"type":"KEYWORD"}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	service.Add(models.Monitor{
		Name: "google-test",
		URL:  "https://google.com",
		Config: &endpointmonitorv1alpha1.UptimeRobotConfig{
			MonitorType:        "keyword",
			KeywordExists:      "no",
			KeywordValue:       "forty-four",
			MaintenanceWindows: "77-78",
			CustomHTTPStatuses: "200:0_401:1_503:1",
			Interval:           600,
			AlertContacts:      "999_1_2",
		},
	})

	if len(rec.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(rec.requests))
	}
	body := rec.requests[0].Body
	if body["type"] != "KEYWORD" {
		t.Errorf("expected type KEYWORD, got %v", body["type"])
	}
	if body["keywordType"] != "ALERT_NOT_EXISTS" {
		t.Errorf("expected keywordType ALERT_NOT_EXISTS, got %v", body["keywordType"])
	}
	if body["keywordValue"] != "forty-four" {
		t.Errorf("expected keywordValue, got %v", body["keywordValue"])
	}
	if body["interval"] != float64(600) {
		t.Errorf("expected interval 600, got %v", body["interval"])
	}
	if ids, ok := body["maintenanceWindowsIds"].([]interface{}); !ok || len(ids) != 2 || ids[0] != float64(77) || ids[1] != float64(78) {
		t.Errorf("expected maintenanceWindowsIds [77 78], got %v", body["maintenanceWindowsIds"])
	}
	if codes, ok := body["successHttpResponseCodes"].([]interface{}); !ok || len(codes) != 2 || codes[0] != "401" || codes[1] != "503" {
		t.Errorf("expected successHttpResponseCodes [401 503], got %v", body["successHttpResponseCodes"])
	}
	if contacts, ok := body["assignedAlertContacts"].([]interface{}); !ok || len(contacts) != 1 {
		t.Errorf("expected 1 assignedAlertContact from provider config, got %v", body["assignedAlertContacts"])
	}
}

func TestV3ProcessProviderConfigDefaults(t *testing.T) {
	service := UpTimeMonitorService{}

	// No config: HTTP monitor with defaults
	request := service.processProviderConfig(models.Monitor{Name: "m", URL: "https://example.com"})
	if request.Type != "HTTP" || request.Interval != 300 || request.Timeout != 30 {
		t.Errorf("expected HTTP/300/30 defaults, got %v/%v/%v", request.Type, request.Interval, request.Timeout)
	}

	// Keyword monitor without KeywordExists defaults to ALERT_EXISTS
	request = service.processProviderConfig(models.Monitor{
		Name:   "m",
		URL:    "https://example.com",
		Config: &endpointmonitorv1alpha1.UptimeRobotConfig{MonitorType: "keyword"},
	})
	if request.Type != "KEYWORD" || request.KeywordType != "ALERT_EXISTS" {
		t.Errorf("expected KEYWORD/ALERT_EXISTS defaults, got %v/%v", request.Type, request.KeywordType)
	}

	// KeywordExists "yes" maps to ALERT_EXISTS, "no" to ALERT_NOT_EXISTS
	request = service.processProviderConfig(models.Monitor{
		Config: &endpointmonitorv1alpha1.UptimeRobotConfig{MonitorType: "keyword", KeywordExists: "yes"},
	})
	if request.KeywordType != "ALERT_EXISTS" {
		t.Errorf("expected ALERT_EXISTS for yes, got %v", request.KeywordType)
	}
}

func TestV3UpdateMonitorPatchRequest(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusOK, `{"id":42,"friendlyName":"google-test","url":"https://facebook.com","type":"HTTP"}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	service.Update(models.Monitor{ID: "42", Name: "google-test", URL: "https://facebook.com"})

	if len(rec.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(rec.requests))
	}
	req := rec.requests[0]
	if req.Method != "PATCH" || req.Path != "/monitors/42" {
		t.Errorf("expected PATCH /monitors/42, got %s %s", req.Method, req.Path)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer test-api-key" {
		t.Errorf("expected Bearer auth header, got %q", got)
	}
	if req.Body["friendlyName"] != "google-test" || req.Body["url"] != "https://facebook.com" {
		t.Errorf("update body mapped incorrectly: %v", req.Body)
	}
}

func TestV3RemoveMonitorDeletesAndDetachesFromStatusPages(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		switch {
		case req.Method == "GET" && req.Path == "/monitors/42":
			writeJSON(w, http.StatusOK, `{"id":42,"friendlyName":"google-test","psps":[{"id":7,"friendlyName":"sp"}]}`)
		case req.Method == "GET" && req.Path == "/psps/7":
			writeJSON(w, http.StatusOK, `{"id":7,"friendlyName":"sp","monitorIds":[42,43]}`)
		case req.Method == "PATCH" && req.Path == "/psps/7":
			writeJSON(w, http.StatusOK, `{"id":7,"friendlyName":"sp"}`)
		case req.Method == "DELETE" && req.Path == "/monitors/42":
			writeJSON(w, http.StatusOK, `{}`)
		default:
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
			writeJSON(w, http.StatusBadRequest, `{}`)
		}
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	service.Remove(models.Monitor{ID: "42", Name: "google-test", URL: "https://google.com"})

	if len(rec.requests) != 4 {
		t.Fatalf("expected 4 requests, got %d: %+v", len(rec.requests), rec.requests)
	}
	// Detach happens before the delete
	expectedSequence := []struct{ method, path string }{
		{"GET", "/monitors/42"},
		{"GET", "/psps/7"},
		{"PATCH", "/psps/7"},
		{"DELETE", "/monitors/42"},
	}
	for i, expected := range expectedSequence {
		if rec.requests[i].Method != expected.method || rec.requests[i].Path != expected.path {
			t.Errorf("request %d: expected %s %s, got %s %s", i, expected.method, expected.path, rec.requests[i].Method, rec.requests[i].Path)
		}
	}
	patchBody := rec.requests[2].Body
	if ids, ok := patchBody["monitorIds"].([]interface{}); !ok || len(ids) != 1 || ids[0] != float64(43) {
		t.Errorf("expected PATCH /psps/7 with monitorIds [43], got %v", patchBody["monitorIds"])
	}
	if patchBody["friendlyName"] != "sp" {
		t.Errorf("expected friendlyName in PATCH body, got %v", patchBody["friendlyName"])
	}
}

func TestV3GetByNamePaginatesNextLink(t *testing.T) {
	rec := &recorder{}
	var server *httptest.Server
	server = httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		if req.Query.Get("cursor") == "" {
			writeJSON(w, http.StatusOK, `{"nextLink":"`+server.URL+`/monitors?cursor=2&name=google-test","data":[{"id":1,"friendlyName":"google-test-similar","url":"https://similar.com","interval":300}]}`)
			return
		}
		writeJSON(w, http.StatusOK, `{"nextLink":null,"data":[{"id":5,"friendlyName":"google-test","url":"https://google.com","type":"HTTP","interval":300,"assignedAlertContacts":[{"alertContactId":111,"threshold":0,"recurrence":0}]}]}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	monitor, err := service.GetByName("google-test")
	if err != nil {
		t.Fatal("GetByName failed: " + err.Error())
	}
	if monitor == nil {
		t.Fatal("expected monitor to be found")
	}
	if monitor.ID != "5" || monitor.Name != "google-test" || monitor.URL != "https://google.com" {
		t.Errorf("monitor mapped incorrectly: %+v", monitor)
	}
	providerConfig, _ := monitor.Config.(*endpointmonitorv1alpha1.UptimeRobotConfig)
	if providerConfig.Interval != 300 {
		t.Errorf("expected interval 300, got %d", providerConfig.Interval)
	}
	if providerConfig.AlertContacts != "111_0_0" {
		t.Errorf("expected alertContacts 111_0_0, got %q", providerConfig.AlertContacts)
	}

	if len(rec.requests) != 2 {
		t.Fatalf("expected 2 paginated requests, got %d", len(rec.requests))
	}
	for i, req := range rec.requests {
		if req.Query.Get("name") != "google-test" {
			t.Errorf("request %d: expected name filter, got query %v", i, req.Query)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-api-key" {
			t.Errorf("request %d: expected Bearer auth header, got %q", i, got)
		}
	}
	if rec.requests[1].Query.Get("cursor") != "2" {
		t.Errorf("expected second page to follow nextLink cursor, got query %v", rec.requests[1].Query)
	}
}

func TestV3GetByNameNotFound(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusOK, `{"nextLink":null,"data":[]}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	monitor, err := service.GetByName("nope")
	if err != nil {
		t.Fatal("expected no error for empty result, got: " + err.Error())
	}
	if monitor != nil {
		t.Errorf("expected nil monitor, got %+v", monitor)
	}
}

func TestV3GetAllAggregatesPages(t *testing.T) {
	rec := &recorder{}
	var server *httptest.Server
	server = httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		if req.Query.Get("cursor") == "" {
			writeJSON(w, http.StatusOK, `{"nextLink":"`+server.URL+`/monitors?cursor=9","data":[{"id":1,"friendlyName":"one","url":"https://one.com","interval":300}]}`)
			return
		}
		writeJSON(w, http.StatusOK, `{"nextLink":null,"data":[{"id":2,"friendlyName":"two","url":"https://two.com","interval":300}]}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	monitors, err := service.GetAll()
	if err != nil {
		t.Fatal("GetAll failed: " + err.Error())
	}
	if len(monitors) != 2 || monitors[0].ID != "1" || monitors[1].ID != "2" {
		t.Errorf("expected monitors from both pages, got %+v", monitors)
	}
}

func TestV3ErrorStatusReturnsError(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusUnauthorized, `{"message":"invalid api key"}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	_, err := service.GetByName("google-test")
	if err == nil {
		t.Fatal("expected error on 401 response")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("error should include status code and body excerpt, got: %s", err.Error())
	}
}

func TestV3RateLimitRetriesHonoredAndBounded(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		w.Header().Set("Retry-After", "0")
		writeJSON(w, http.StatusTooManyRequests, `{}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	_, err := service.GetByName("google-test")
	if err == nil {
		t.Fatal("expected error after rate limit retries are exhausted")
	}
	// 1 initial request + maxRateLimitRetries retries
	if len(rec.requests) != maxRateLimitRetries+1 {
		t.Errorf("expected %d requests before giving up, got %d", maxRateLimitRetries+1, len(rec.requests))
	}
}

func TestV3RateLimitRecoversAfterRetry(t *testing.T) {
	rec := &recorder{}
	requests := 0
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		requests++
		if requests == 1 {
			w.Header().Set("Retry-After", "0")
			writeJSON(w, http.StatusTooManyRequests, `{}`)
			return
		}
		writeJSON(w, http.StatusOK, `{"nextLink":null,"data":[{"id":5,"friendlyName":"google-test","url":"https://google.com","interval":300}]}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	monitor, err := service.GetByName("google-test")
	if err != nil || monitor == nil {
		t.Fatalf("expected monitor after successful retry, got monitor=%v err=%v", monitor, err)
	}
	if monitor.ID != "5" {
		t.Errorf("expected monitor id 5, got %s", monitor.ID)
	}
}

func TestV3AddMonitorToStatusPage(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		switch {
		case req.Method == "GET" && req.Path == "/psps/7":
			writeJSON(w, http.StatusOK, `{"id":7,"friendlyName":"sp","monitorIds":[43]}`)
		case req.Method == "PATCH" && req.Path == "/psps/7":
			writeJSON(w, http.StatusOK, `{"id":7,"friendlyName":"sp"}`)
		default:
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
			writeJSON(w, http.StatusBadRequest, `{}`)
		}
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	returnedID, err := service.statusPageService.AddMonitorToStatusPage(UpTimeStatusPage{ID: "7", Name: "sp"}, models.Monitor{ID: "42"})
	if err != nil {
		t.Fatal("AddMonitorToStatusPage failed: " + err.Error())
	}
	if returnedID != "7" {
		t.Errorf("expected status page id 7, got %s", returnedID)
	}

	if len(rec.requests) != 2 {
		t.Fatalf("expected GET + PATCH, got %d requests: %+v", len(rec.requests), rec.requests)
	}
	patch := rec.requests[1]
	if patch.Method != "PATCH" || patch.Path != "/psps/7" {
		t.Errorf("expected PATCH /psps/7, got %s %s", patch.Method, patch.Path)
	}
	if ids, ok := patch.Body["monitorIds"].([]interface{}); !ok || len(ids) != 2 || ids[0] != float64(43) || ids[1] != float64(42) {
		t.Errorf("expected full monitorIds [43 42] in PATCH, got %v", patch.Body["monitorIds"])
	}
}

func TestV3AddMonitorToStatusPageIdempotent(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusOK, `{"id":7,"friendlyName":"sp","monitorIds":[42,43]}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	_, err := service.statusPageService.AddMonitorToStatusPage(UpTimeStatusPage{ID: "7", Name: "sp"}, models.Monitor{ID: "42"})
	if err != nil {
		t.Fatal("AddMonitorToStatusPage failed: " + err.Error())
	}
	if len(rec.requests) != 1 {
		t.Errorf("expected only the GET (no PATCH when monitor already attached), got %d requests", len(rec.requests))
	}
}

func TestV3RemoveMonitorFromStatusPage(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		switch {
		case req.Method == "GET" && req.Path == "/psps/7":
			writeJSON(w, http.StatusOK, `{"id":7,"friendlyName":"sp","monitorIds":[42,43]}`)
		case req.Method == "PATCH" && req.Path == "/psps/7":
			writeJSON(w, http.StatusOK, `{"id":7,"friendlyName":"sp"}`)
		default:
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
			writeJSON(w, http.StatusBadRequest, `{}`)
		}
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	_, err := service.statusPageService.RemoveMonitorFromStatusPage(UpTimeStatusPage{ID: "7", Name: "sp"}, models.Monitor{ID: "42"})
	if err != nil {
		t.Fatal("RemoveMonitorFromStatusPage failed: " + err.Error())
	}
	if len(rec.requests) != 2 {
		t.Fatalf("expected GET + PATCH, got %d requests", len(rec.requests))
	}
	if ids, ok := rec.requests[1].Body["monitorIds"].([]interface{}); !ok || len(ids) != 1 || ids[0] != float64(43) {
		t.Errorf("expected monitorIds [43] after removal, got %v", rec.requests[1].Body["monitorIds"])
	}
}

func TestV3StatusPageAddUsesAllMonitorsSentinel(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusCreated, `{"id":9,"friendlyName":"my-page"}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	id, err := service.statusPageService.Add(UpTimeStatusPage{Name: "my-page"})
	if err != nil {
		t.Fatal("Status page Add failed: " + err.Error())
	}
	if id != "9" {
		t.Errorf("expected status page id 9, got %s", id)
	}
	body := rec.requests[0].Body
	if ids, ok := body["monitorIds"].([]interface{}); !ok || len(ids) != 1 || ids[0] != float64(0) {
		t.Errorf("expected [0] all-monitors sentinel, got %v", body["monitorIds"])
	}
	if body["status"] != "ENABLED" {
		t.Errorf("expected status ENABLED, got %v", body["status"])
	}
}

func TestV3GetStatusPagesForMonitorUsesMonitorPsps(t *testing.T) {
	rec := &recorder{}
	server := httptest.NewServer(rec.handler(t, func(w http.ResponseWriter, req *recordedRequest) {
		writeJSON(w, http.StatusOK, `{"id":42,"friendlyName":"m","psps":[{"id":7,"friendlyName":"sp1"},{"id":8,"friendlyName":"sp2"}]}`)
	}))
	defer server.Close()

	service := newTestService(server.URL, "")
	pspIDs, err := service.statusPageService.GetStatusPagesForMonitor("42")
	if err != nil {
		t.Fatal("GetStatusPagesForMonitor failed: " + err.Error())
	}
	if len(pspIDs) != 2 || pspIDs[0] != "7" || pspIDs[1] != "8" {
		t.Errorf("expected psp ids [7 8], got %v", pspIDs)
	}
	if rec.requests[0].Path != "/monitors/42" {
		t.Errorf("expected request to /monitors/42, got %s", rec.requests[0].Path)
	}
}
