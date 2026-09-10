package betterstack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

func newService(url string) *BetterStackMonitorService {
	service := &BetterStackMonitorService{}
	service.Setup(config.Provider{Name: "BetterStack", ApiToken: "test-token", ApiURL: url})
	return service
}

// Add must send the URL and name Better Stack expects, under a bearer token.
// A regression here creates monitors that watch nothing.
func TestAddSendsExpectedPayload(t *testing.T) {
	var gotBody map[string]interface{}
	var gotAuth, gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"data":{"id":"42","attributes":{}}}`))
	}))
	defer server.Close()

	newService(server.URL).Add(models.NewMonitor("web", "", "https://web.example/health", nil))

	if gotMethod != "POST" || gotPath != monitorsPath {
		t.Errorf("expected POST %s, got %s %s", monitorsPath, gotMethod, gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("expected bearer token, got %q", gotAuth)
	}
	if gotBody["url"] != "https://web.example/health" {
		t.Errorf("unexpected url: %v", gotBody["url"])
	}
	if gotBody["pronounceable_name"] != "web" {
		t.Errorf("unexpected name: %v", gotBody["pronounceable_name"])
	}
	// Defaults must be sent on create, or the monitor silently takes whatever
	// Better Stack decides today.
	if gotBody["check_frequency"] != float64(DefaultCheckFrequency) {
		t.Errorf("expected default check_frequency, got %v", gotBody["check_frequency"])
	}
	if gotBody["monitor_type"] != DefaultMonitorType {
		t.Errorf("expected default monitor_type, got %v", gotBody["monitor_type"])
	}
}

// An unset boolean must be omitted, not sent as false: sending false would
// overwrite a setting made in Better Stack's UI.
func TestUnsetBooleansAreOmitted(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"data":{"id":"1","attributes":{}}}`))
	}))
	defer server.Close()

	monitor := models.NewMonitor("web", "", "https://web.example/", &endpointmonitorv1alpha1.BetterStackConfig{
		VerifySSL: "false",
	})
	newService(server.URL).Add(monitor)

	if _, present := gotBody["paused"]; present {
		t.Error("paused was sent despite being unset")
	}
	if verify, present := gotBody["verify_ssl"]; !present || verify != false {
		t.Errorf("verify_ssl should be explicit false, got %v (present=%v)", verify, present)
	}
}

// Update must PATCH, not PUT. A PUT replaces the resource and clears every
// field the controller does not manage.
func TestUpdateUsesPatchAndId(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data":{"id":"42","attributes":{}}}`))
	}))
	defer server.Close()

	newService(server.URL).Update(models.NewMonitor("web", "42", "https://web.example/", nil))

	if gotMethod != "PATCH" {
		t.Errorf("expected PATCH, got %s", gotMethod)
	}
	if gotPath != monitorsPath+"/42" {
		t.Errorf("expected id in path, got %s", gotPath)
	}
}

// GetAll must follow pagination. Stopping at page one makes GetByName report
// an existing monitor as missing, and the controller then creates a duplicate.
func TestGetAllFollowsPagination(t *testing.T) {
	// `next` is deliberately a RELATIVE url here: the provider must page by
	// number and treat next only as a "there is more" flag. A provider that
	// fetched this value directly would request a bad url and lose page 2.
	var pagesSeen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pagesSeen = append(pagesSeen, page)
		w.WriteHeader(200)
		if page == "2" {
			_, _ = w.Write([]byte(`{"data":[{"id":"2","attributes":{"url":"https://b.example/","pronounceable_name":"b"}}],"pagination":{"next":""}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"url":"https://a.example/","pronounceable_name":"a"}}],"pagination":{"next":"/api/v2/monitors?page=2"}}`))
	}))
	defer server.Close()

	monitors, err := newService(server.URL).GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(monitors) != 2 {
		t.Fatalf("expected 2 monitors across pages, got %d", len(monitors))
	}
	if monitors[1].Name != "b" || monitors[1].ID != "2" {
		t.Errorf("second page not mapped: %+v", monitors[1])
	}
	if len(pagesSeen) != 2 || pagesSeen[0] != "1" || pagesSeen[1] != "2" {
		t.Errorf("expected pages 1 then 2 by number, got %v", pagesSeen)
	}
}

// GetByName returning (nil, nil) is the create path; an error would stop the
// controller reconciling a monitor that simply does not exist yet.
func TestGetByNameMissingIsNotAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data":[],"pagination":{"next":""}}`))
	}))
	defer server.Close()

	monitor, err := newService(server.URL).GetByName("absent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if monitor != nil {
		t.Errorf("expected nil monitor, got %+v", monitor)
	}
}

// A delete of an already-absent monitor is the requested end state. Treating
// 404 as failure leaves the controller retrying forever.
func TestRemoveTreatsMissingAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()
	// No assertion beyond "does not panic and does not block": Remove has no
	// return value, so this guards the 404 branch staying reachable.
	newService(server.URL).Remove(models.NewMonitor("gone", "7", "https://gone.example/", nil))
}

func TestEqual(t *testing.T) {
	service := &BetterStackMonitorService{}
	base := models.NewMonitor("web", "1", "https://web.example/", nil)

	if !service.Equal(base, models.NewMonitor("web", "1", "https://web.example/", nil)) {
		t.Error("identical monitors should compare equal")
	}
	if service.Equal(base, models.NewMonitor("web", "1", "https://other.example/", nil)) {
		t.Error("differing URL should not compare equal")
	}

	withConfig := models.NewMonitor("web", "1", "https://web.example/",
		&endpointmonitorv1alpha1.BetterStackConfig{CheckFrequency: 30})
	if service.Equal(base, withConfig) {
		t.Error("added config should not compare equal")
	}
	if !service.Equal(withConfig, models.NewMonitor("web", "1", "https://web.example/",
		&endpointmonitorv1alpha1.BetterStackConfig{CheckFrequency: 30})) {
		t.Error("identical config should compare equal")
	}
}

func TestParseStatusCodesIgnoresJunk(t *testing.T) {
	codes := parseStatusCodes("200, 302 ,notanumber,201")
	if len(codes) != 3 || codes[0] != 200 || codes[1] != 302 || codes[2] != 201 {
		t.Errorf("unexpected codes: %v", codes)
	}
}
