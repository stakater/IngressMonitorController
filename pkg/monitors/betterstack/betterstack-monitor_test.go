package betterstack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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

// The account-wide escalation policy from the provider config must apply to a
// monitor whose CR does not name one — the same "global default, per-CR
// override" shape the UptimeRobot and Pingdom providers use for alertContacts.
func TestGlobalPolicyAppliesWhenCRHasNone(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"data":{"id":"1","attributes":{}}}`))
	}))
	defer server.Close()

	service := &BetterStackMonitorService{}
	service.Setup(config.Provider{
		Name: "BetterStack", ApiToken: "t", ApiURL: server.URL, AlertContacts: "12345",
	})
	service.Add(models.NewMonitor("web", "", "https://web.example/", nil))

	if gotBody["policy_id"] != "12345" {
		t.Errorf("expected the global policy, got %v", gotBody["policy_id"])
	}
}

// A policy named on the CR wins over the global default.
func TestCRPolicyOverridesGlobal(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"data":{"id":"1","attributes":{}}}`))
	}))
	defer server.Close()

	service := &BetterStackMonitorService{}
	service.Setup(config.Provider{
		Name: "BetterStack", ApiToken: "t", ApiURL: server.URL, AlertContacts: "12345",
	})
	service.Add(models.NewMonitor("web", "", "https://web.example/",
		&endpointmonitorv1alpha1.BetterStackConfig{PolicyID: "99999"}))

	if gotBody["policy_id"] != "99999" {
		t.Errorf("expected the CR policy to win, got %v", gotBody["policy_id"])
	}
}

// A monitor read back from Better Stack must compare equal to the CR that
// produced it, or the controller updates it once per reconcile forever.
//
// This is the regression behind `Monitor web-redirect-web has been updated`
// repeating every five minutes in production: GetAll built a Monitor with no
// Config at all, so Equal hit its "exactly one side is nil" branch every time
// any EndpointMonitor set betterStackConfig. Monitors without one were quiet
// only because nil == nil compares equal, which is what hid it.
//
// The round trip is the assertion: build attributes from a config, hand them
// back as an API response, and require the result to settle.
func TestEqualAfterRoundTripThroughAPI(t *testing.T) {
	service := newService("http://example.invalid")

	for name, cfg := range map[string]*endpointmonitorv1alpha1.BetterStackConfig{
		"redirect monitor": {
			MonitorType:         "expected_status_code",
			ExpectedStatusCodes: "200,301,302",
		},
		"paid settings": {
			CheckFrequency: 30,
			PolicyID:       "123562",
			Regions:        "eu,us",
		},
		"explicit false stays false": {
			VerifySSL: "false",
			Paused:    "false",
		},
		"empty config": {},
	} {
		t.Run(name, func(t *testing.T) {
			desired := models.NewMonitor("web", "1", "https://web.example/", cfg)

			// What Add/Update would send, echoed back as Better Stack reports it.
			live := toBaseMonitor(monitorData{
				ID:         "1",
				Attributes: service.buildAttributes(desired, true),
			})
			live.Name = desired.Name
			live.URL = desired.URL

			if !service.Equal(live, desired) {
				t.Errorf("monitor does not settle after a round trip: "+
					"live=%+v desired=%+v", getConfig(live), getConfig(desired))
			}
		})
	}
}

// A value set in Better Stack's UI, for a field the CR says nothing about,
// must not read as drift -- otherwise the controller fights whoever set it on
// every reconcile.
func TestEqualIgnoresUnmanagedUIChanges(t *testing.T) {
	service := newService("http://example.invalid")

	cfg := &endpointmonitorv1alpha1.BetterStackConfig{CheckFrequency: 30}
	desired := models.NewMonitor("web", "1", "https://web.example/", cfg)

	// As created, then someone flips a field the CR never mentions.
	attributes := service.buildAttributes(desired, true)
	attributes.TeamWait = intPtr(300)
	attributes.RequestTimeout = intPtr(45)

	live := toBaseMonitor(monitorData{ID: "1", Attributes: attributes})
	live.Name, live.URL = desired.Name, desired.URL

	if !service.Equal(live, desired) {
		t.Errorf("a UI-set field the CR does not manage was treated as drift: "+
			"live=%+v desired=%+v", getConfig(live), getConfig(desired))
	}
}

// The exact production regression: web-redirect-web, the only monitor with a
// betterStackConfig, was logging "has been updated" every five minutes.
// Its CR is reproduced verbatim here.
func TestWebRedirectDoesNotLoop(t *testing.T) {
	service := newService("http://example.invalid")

	cfg := &endpointmonitorv1alpha1.BetterStackConfig{
		MonitorType:         "expected_status_code",
		ExpectedStatusCodes: "200,301,302",
	}
	desired := models.NewMonitor("web-redirect-web", "1", "https://ester.io/", cfg)

	// Created once, then read back on the next reconcile.
	live := toBaseMonitor(monitorData{
		ID:         "1",
		Attributes: service.buildAttributes(desired, true),
	})
	live.Name, live.URL = desired.Name, desired.URL

	if !service.Equal(live, desired) {
		t.Fatalf("web-redirect still updates on every reconcile: live=%+v desired=%+v",
			getConfig(live), getConfig(desired))
	}

	// A real spec change must still be seen, or nothing would ever update.
	changed := models.NewMonitor("web-redirect-web", "1", "https://ester.io/",
		&endpointmonitorv1alpha1.BetterStackConfig{
			MonitorType:         "expected_status_code",
			ExpectedStatusCodes: "200,301,302",
			CheckFrequency:      30,
		})
	if service.Equal(live, changed) {
		t.Error("a changed checkFrequency was not detected as drift")
	}
}

// The invariant that keeps this provider out of an update loop: if Equal
// reports drift, the payload Update actually sends must resolve it. Comparing
// anything Update omits -- a default, a provider-wide policy -- reports drift
// that no PATCH can fix, and the monitor is rewritten every reconcile forever.
//
// Each case below is a live monitor that differs from what the CR asked for in
// a field the CR does not manage. All must settle.
func TestReconcileConverges(t *testing.T) {
	service := newService("http://example.invalid")
	withPolicy := &BetterStackMonitorService{}
	withPolicy.Setup(config.Provider{
		Name: "BetterStack", ApiToken: "t",
		ApiURL: "http://example.invalid", AlertContacts: "999",
	})

	cases := []struct {
		name    string
		service *BetterStackMonitorService
		cr      *endpointmonitorv1alpha1.BetterStackConfig
		mutate  func(*monitorAttributes)
	}{{
		// An adopted monitor, or one whose interval was set in the UI, while
		// the CR says nothing about checkFrequency.
		name:   "UI check_frequency, unset in CR",
		cr:     &endpointmonitorv1alpha1.BetterStackConfig{MonitorType: "status"},
		mutate: func(a *monitorAttributes) { a.CheckFrequency = intPtr(300) },
	}, {
		name:    "provider-wide policy differs from the live monitor",
		service: withPolicy,
		cr:      &endpointmonitorv1alpha1.BetterStackConfig{CheckFrequency: 30},
		mutate:  func(a *monitorAttributes) { a.PolicyID = strPtr("123562") },
	}, {
		name:   "UI monitor_type, unset in CR",
		cr:     &endpointmonitorv1alpha1.BetterStackConfig{CheckFrequency: 30},
		mutate: func(a *monitorAttributes) { a.MonitorType = strPtr("keyword") },
	}, {
		// The API is free to return these in its own order.
		name: "status codes reported in a different order",
		cr: &endpointmonitorv1alpha1.BetterStackConfig{
			MonitorType: "expected_status_code", ExpectedStatusCodes: "302,301,200",
		},
		mutate: func(a *monitorAttributes) { a.ExpectedStatusCodes = &[]int{200, 301, 302} },
	}, {
		name:   "regions reported in a different order",
		cr:     &endpointmonitorv1alpha1.BetterStackConfig{Regions: "us,eu"},
		mutate: func(a *monitorAttributes) { a.Regions = &[]string{"eu", "us"} },
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := tc.service
			if svc == nil {
				svc = service
			}
			desired := models.NewMonitor("web", "1", "https://web.example/", tc.cr)
			attributes := svc.buildAttributes(desired, true)
			tc.mutate(&attributes)

			// Equal -> Update -> re-read -> Equal, as the controller runs it.
			for pass := 0; ; pass++ {
				live := toBaseMonitor(monitorData{ID: "1", Attributes: attributes})
				live.Name, live.URL = desired.Name, desired.URL
				if svc.Equal(live, desired) {
					return
				}
				if pass == 3 {
					t.Fatalf("never settles: the update does not resolve the drift "+
						"it reports (live=%+v desired=%+v)", getConfig(live), getConfig(desired))
				}
				patch := svc.buildAttributes(desired, false)
				attributes = applyNonNil(attributes, patch)
			}
		})
	}
}

// applyNonNil mimics a PATCH: only the fields actually sent are changed.
func applyNonNil(live, patch monitorAttributes) monitorAttributes {
	l, p := reflect.ValueOf(&live).Elem(), reflect.ValueOf(patch)
	for i := 0; i < p.NumField(); i++ {
		if !p.Field(i).IsNil() {
			l.Field(i).Set(p.Field(i))
		}
	}
	return live
}

// A change the CR DOES pin must still be corrected -- the convergence fix must
// not have bought quiet by ignoring real drift.
func TestRealDriftIsStillDetected(t *testing.T) {
	service := newService("http://example.invalid")

	cases := []struct {
		name   string
		cr     *endpointmonitorv1alpha1.BetterStackConfig
		mutate func(*monitorAttributes)
	}{
		{"checkFrequency", &endpointmonitorv1alpha1.BetterStackConfig{CheckFrequency: 30},
			func(a *monitorAttributes) { a.CheckFrequency = intPtr(300) }},
		{"policyID cleared", &endpointmonitorv1alpha1.BetterStackConfig{PolicyID: "123562"},
			func(a *monitorAttributes) { a.PolicyID = strPtr("") }},
		{"expectedStatusCodes", &endpointmonitorv1alpha1.BetterStackConfig{
			MonitorType: "expected_status_code", ExpectedStatusCodes: "200,301,302"},
			func(a *monitorAttributes) { a.ExpectedStatusCodes = &[]int{200} }},
		{"regions", &endpointmonitorv1alpha1.BetterStackConfig{Regions: "eu,us"},
			func(a *monitorAttributes) { a.Regions = &[]string{"eu"} }},
		{"paused", &endpointmonitorv1alpha1.BetterStackConfig{Paused: "false"},
			func(a *monitorAttributes) { a.Paused = boolPtr(true) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			desired := models.NewMonitor("web", "1", "https://web.example/", tc.cr)
			attributes := service.buildAttributes(desired, true)
			tc.mutate(&attributes)

			live := toBaseMonitor(monitorData{ID: "1", Attributes: attributes})
			live.Name, live.URL = desired.Name, desired.URL

			if service.Equal(live, desired) {
				t.Error("drift on a field the CR pins was not detected, so the " +
					"monitor will never be corrected")
			}
		})
	}
}

// equalWhereSet reads every attribute as a pointer. A value-typed field added
// later would panic inside the reconcile loop, so fail here instead, at review.
func TestMonitorAttributesAreAllPointers(t *testing.T) {
	typ := reflect.TypeOf(monitorAttributes{})
	for i := 0; i < typ.NumField(); i++ {
		if f := typ.Field(i); f.Type.Kind() != reflect.Pointer {
			t.Errorf("monitorAttributes.%s is %s, not a pointer: make it a pointer, "+
				"or teach equalWhereSet to handle value fields", f.Name, f.Type.Kind())
		}
	}
}

// The live side in production is decoded JSON, not a struct handed straight
// back, so exercise the tags too.
func TestEqualAfterRealJSONDecode(t *testing.T) {
	service := newService("http://example.invalid")
	desired := models.NewMonitor("web", "1", "https://web.example/",
		&endpointmonitorv1alpha1.BetterStackConfig{MonitorType: "status", CheckFrequency: 30})

	raw, err := json.Marshal(map[string]any{"data": map[string]any{
		"id": "1", "attributes": service.buildAttributes(desired, true),
	}})
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Data monitorData `json:"data"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}

	live := toBaseMonitor(decoded.Data)
	live.Name, live.URL = desired.Name, desired.URL
	if !service.Equal(live, desired) {
		t.Errorf("does not settle after a real JSON round trip: live=%+v desired=%+v",
			getConfig(live), getConfig(desired))
	}
}

// formatBool must be parseBool's exact inverse, or the tri-state leaks and an
// unset boolean starts reading as an explicit false.
func TestFormatBoolInvertsParseBool(t *testing.T) {
	for _, raw := range []string{"", "true", "false"} {
		if got := formatBool(parseBool(raw)); got != raw {
			t.Errorf("formatBool(parseBool(%q)) = %q, want %q", raw, got, raw)
		}
	}
}
