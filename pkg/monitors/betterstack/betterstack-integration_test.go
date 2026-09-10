//go:build integration

// Live API test for the Better Stack provider.
//
// Skipped unless BETTERSTACK_API_TOKEN is set, and only built under the
// `integration` tag, so a normal `go test ./...` never touches the network or
// a real account.
//
//	BETTERSTACK_API_TOKEN=... go test -tags=integration -v ./pkg/monitors/betterstack/...
//
// It creates a monitor against a throwaway URL, exercises the full lifecycle,
// and deletes it — including on failure. The point is to catch what a fake
// server cannot: whether the auth header is accepted, whether the field names
// and defaults are real, and whether create actually works on this account's
// plan. That last one is exactly what silently failed with UptimeRobot, where
// every read returned 200 and every create returned 403.
package betterstack

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	imchttp "github.com/stakater/IngressMonitorController/v2/pkg/http"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

// A URL that is real, stable and not ours, so a monitor briefly pointed at it
// is harmless and obviously not production.
const probeURL = "https://example.com/"

func liveService(t *testing.T) *BetterStackMonitorService {
	t.Helper()
	token := os.Getenv("BETTERSTACK_API_TOKEN")
	if token == "" {
		t.Skip("BETTERSTACK_API_TOKEN not set; skipping live API test")
	}
	service := &BetterStackMonitorService{}
	service.Setup(config.Provider{Name: "BetterStack", ApiToken: token})
	return service
}

func TestLiveMonitorLifecycle(t *testing.T) {
	service := liveService(t)
	name := fmt.Sprintf("imc-selftest-%d", time.Now().UnixNano())

	// Paused so the monitor never actually probes anything or alerts anyone
	// during the test.
	monitor := models.NewMonitor(name, "", probeURL, &endpointmonitorv1alpha1.BetterStackConfig{
		CheckFrequency: 300,
		Paused:         "true",
		Email:          "false",
		SMS:            "false",
		Call:           "false",
		Push:           "false",
	})

	t.Logf("creating monitor %s", name)
	service.Add(monitor)

	// Add has no return value, so existence is the assertion.
	created, err := service.GetByName(name)
	if err != nil {
		t.Fatalf("GetByName after Add: %v", err)
	}
	if created == nil {
		t.Fatal("monitor was not created — Add reported no error but nothing exists. " +
			"Check the provider log line above for the API's response.")
	}
	t.Logf("created id=%s url=%s", created.ID, created.URL)

	// Always clean up, including when a later assertion fails.
	defer func() {
		t.Logf("deleting monitor id=%s", created.ID)
		service.Remove(*created)
		if remaining, err := service.GetByName(name); err == nil && remaining != nil {
			t.Errorf("monitor %s still exists after Remove", name)
		}
	}()

	if created.URL != probeURL {
		t.Errorf("URL round-trip failed: sent %q, got %q", probeURL, created.URL)
	}
	if created.Name != name {
		t.Errorf("name round-trip failed: sent %q, got %q", name, created.Name)
	}

	// Update must not 404 or reset the monitor.
	updated := models.NewMonitor(name, created.ID, probeURL, &endpointmonitorv1alpha1.BetterStackConfig{
		CheckFrequency: 600,
		Paused:         "true",
	})
	t.Log("updating monitor")
	service.Update(updated)

	afterUpdate, err := service.GetByName(name)
	if err != nil {
		t.Fatalf("GetByName after Update: %v", err)
	}
	if afterUpdate == nil {
		t.Fatal("monitor disappeared after Update")
	}

	// GetAll must include it, which is what GetByName and duplicate-detection
	// both depend on.
	all, err := service.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	var found bool
	for _, m := range all {
		if m.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetAll returned %d monitors but not %s", len(all), name)
	}
	t.Logf("GetAll returned %d monitors", len(all))
}

// Reads alone are not evidence that writes work — the UptimeRobot failure had
// every get returning 200 while create returned 403. This asserts the token is
// accepted at all, so a failing lifecycle test can be told apart from a bad
// token.
func TestLiveTokenIsAccepted(t *testing.T) {
	service := liveService(t)
	if _, err := service.GetAll(); err != nil {
		t.Fatalf("token rejected or API unreachable: %v", err)
	}
}

func attrs(t *testing.T, s *BetterStackMonitorService, id string) map[string]interface{} {
	t.Helper()
	c := imchttp.CreateHttpClient(s.url + monitorsPath + "/" + id)
	resp := c.GetUrl(s.headers(), []byte{})
	var out struct {
		Data struct {
			Attributes map[string]interface{} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Bytes, &out); err != nil {
		t.Fatalf("decode attributes: %v", err)
	}
	return out.Data.Attributes
}

// Every alert-shaping field must survive the round trip. Sending them and never
// reading them back proved nothing: the first live run created a monitor whose
// URL and name were right while the API rejected the whole alert set.
func TestLiveAlertParametersRoundTrip(t *testing.T) {
	s := liveService(t)
	name := "imc-alerts-" + time.Now().Format("150405")

	s.Add(models.NewMonitor(name, "", "https://example.com/", &endpointmonitorv1alpha1.BetterStackConfig{
		CheckFrequency:     120,
		MonitorType:        "status",
		Paused:             "true",
		Email:              "false",
		SMS:                "true",
		Call:               "true",
		Push:               "false",
		VerifySSL:          "true",
		RequestTimeout:     45,
		ConfirmationPeriod: 180,
		RecoveryPeriod:     180,
		TeamWait:           300,
		Regions:            "eu,us",
	}))

	m, err := s.GetByName(name)
	if err != nil || m == nil {
		t.Fatalf("monitor not created: %v", err)
	}
	defer s.Remove(*m)

	got := attrs(t, s, m.ID)
	for field, want := range map[string]interface{}{
		"check_frequency":     float64(120),
		"monitor_type":        "status",
		"paused":              true,
		"email":               false,
		"sms":                 true,
		"call":                true,
		"push":                false,
		"verify_ssl":          true,
		"request_timeout":     float64(45),
		"confirmation_period": float64(180),
		"recovery_period":     float64(180),
		"team_wait":           float64(300),
	} {
		if got[field] != want {
			t.Errorf("%s: sent %v, API returned %v", field, want, got[field])
		}
	}
}

// The apex redirect monitor: expecting a 3xx while following redirects or
// keeping cookies is rejected by Better Stack with a 422. The provider must
// turn both off on its own, or every redirect monitor silently fails to create.
func TestLiveRedirectMonitorCreates(t *testing.T) {
	s := liveService(t)
	name := "imc-redirect-" + time.Now().Format("150405")

	s.Add(models.NewMonitor(name, "", "https://example.com/", &endpointmonitorv1alpha1.BetterStackConfig{
		MonitorType:         "expected_status_code",
		ExpectedStatusCodes: "200,302",
		Paused:              "true",
	}))

	m, err := s.GetByName(name)
	if err != nil || m == nil {
		t.Fatal("redirect monitor was not created; the 3xx/follow-redirects conflict is not handled")
	}
	defer s.Remove(*m)

	got := attrs(t, s, m.ID)
	if got["follow_redirects"] != false {
		t.Errorf("follow_redirects should be forced false for a 3xx monitor, got %v", got["follow_redirects"])
	}
	if got["remember_cookies"] != false {
		t.Errorf("remember_cookies should be forced false for a 3xx monitor, got %v", got["remember_cookies"])
	}
}

// An out-of-enum recoveryPeriod must be dropped rather than sent, or the whole
// create 422s and no monitor exists at all.
func TestLiveInvalidRecoveryPeriodIsDropped(t *testing.T) {
	s := liveService(t)
	name := "imc-recovery-" + time.Now().Format("150405")

	s.Add(models.NewMonitor(name, "", "https://example.com/", &endpointmonitorv1alpha1.BetterStackConfig{
		Paused:         "true",
		RecoveryPeriod: 120, // not in [0 60 180 300 900 1800 3600 7200]
	}))

	m, err := s.GetByName(name)
	if err != nil || m == nil {
		t.Fatal("monitor was not created; an invalid recoveryPeriod must be dropped, not sent")
	}
	s.Remove(*m)
}
