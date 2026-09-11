package uptimekuma

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/breml/go-uptime-kuma-client/monitor"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
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

func TestParseNotificationIDs(t *testing.T) {
	cases := []struct {
		input    string
		expected []int64
	}{
		{"", nil},
		{"1", []int64{1}},
		{"1,2,3", []int64{1, 2, 3}},
		{"3-1-2", []int64{1, 2, 3}},
		{"5,1", []int64{1, 5}},
		{" 4 , 2 ", []int64{2, 4}},
		{"1,abc,3", []int64{1, 3}},
		{"abc", nil},
	}
	for _, c := range cases {
		actual := parseNotificationIDs(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("parseNotificationIDs(%q) = %v, expected %v", c.input, actual, c.expected)
		}
	}
}

func TestJoinNotificationIDs(t *testing.T) {
	if got := joinNotificationIDs([]int64{1, 2}); got != "1,2" {
		t.Errorf("joinNotificationIDs = %q, expected %q", got, "1,2")
	}
	if got := joinNotificationIDs(nil); got != "" {
		t.Errorf("joinNotificationIDs(nil) = %q, expected empty", got)
	}
}

func TestKeywordInvertMapping(t *testing.T) {
	if keywordInvert("yes") {
		t.Error("KeywordExists yes should map to invertKeyword false")
	}
	if !keywordInvert("no") {
		t.Error("KeywordExists no should map to invertKeyword true")
	}
	if keywordInvert("") {
		t.Error("empty KeywordExists should default to invertKeyword false")
	}
	if got := invertToKeywordExists(true); got != "no" {
		t.Errorf("invertToKeywordExists(true) = %q, expected %q", got, "no")
	}
	if got := invertToKeywordExists(false); got != "yes" {
		t.Errorf("invertToKeywordExists(false) = %q, expected %q", got, "yes")
	}
}

func TestNormalizeConfigDefaults(t *testing.T) {
	normalized := normalizeConfig(nil)
	if normalized.Interval != DefaultInterval || normalized.MonitorType != "http" || normalized.KeywordExists != "yes" || normalized.Notifications != "" {
		t.Errorf("normalizeConfig(nil) did not apply defaults: %+v", normalized)
	}

	normalized = normalizeConfig(&endpointmonitorv1alpha1.UptimeKumaConfig{})
	if normalized.Interval != DefaultInterval || normalized.MonitorType != "http" || normalized.KeywordExists != "yes" {
		t.Errorf("normalizeConfig(empty) did not apply defaults: %+v", normalized)
	}
}

func TestNormalizeConfigAppliesValues(t *testing.T) {
	normalized := normalizeConfig(&endpointmonitorv1alpha1.UptimeKumaConfig{
		Interval:      120,
		MonitorType:   "Keyword",
		KeywordExists: "No",
		KeywordValue:  "foo",
		Notifications: "5,1",
	})

	if normalized.Interval != 120 {
		t.Errorf("expected interval 120, got %d", normalized.Interval)
	}
	if normalized.MonitorType != "keyword" {
		t.Errorf("expected lowercased monitor type keyword, got %q", normalized.MonitorType)
	}
	if normalized.KeywordExists != "no" {
		t.Errorf("expected lowercased keyword exists no, got %q", normalized.KeywordExists)
	}
	if normalized.KeywordValue != "foo" {
		t.Errorf("expected keyword value foo, got %q", normalized.KeywordValue)
	}
	// Canonical (sorted) form so "5,1" and "1-5" compare equal
	if normalized.Notifications != "1,5" {
		t.Errorf("expected canonical notifications 1,5, got %q", normalized.Notifications)
	}
}

func TestNormalizeConfigCanonicalNotifications(t *testing.T) {
	commaSeparated := normalizeConfig(&endpointmonitorv1alpha1.UptimeKumaConfig{Notifications: "5,1"})
	dashSeparated := normalizeConfig(&endpointmonitorv1alpha1.UptimeKumaConfig{Notifications: "1-5"})
	if !reflect.DeepEqual(commaSeparated, dashSeparated) {
		t.Errorf("notification formats should normalize to the same config: %+v vs %+v", commaSeparated, dashSeparated)
	}
}

func TestBuildKumaMonitorHttp(t *testing.T) {
	normalized := normalizeConfig(&endpointmonitorv1alpha1.UptimeKumaConfig{Interval: 90, Notifications: "2,1"})

	mon := buildKumaMonitor(models.Monitor{Name: "test", URL: "https://example.com"}, normalized)
	httpMonitor, ok := mon.(*monitor.HTTP)
	if !ok {
		t.Fatalf("expected a monitor.HTTP, got %T", mon)
	}
	if httpMonitor.Name != "test" || httpMonitor.URL != "https://example.com" {
		t.Errorf("monitor mapped incorrectly: %+v", httpMonitor)
	}
	if httpMonitor.Interval != 90 {
		t.Errorf("expected interval 90, got %d", httpMonitor.Interval)
	}
	if !httpMonitor.IsActive {
		t.Error("expected created monitor to be active")
	}
	if !reflect.DeepEqual(httpMonitor.NotificationIDs, []int64{1, 2}) {
		t.Errorf("expected notification ids [1 2], got %v", httpMonitor.NotificationIDs)
	}
	if httpMonitor.Method != "GET" {
		t.Errorf("expected method GET, got %q", httpMonitor.Method)
	}
}

func TestBuildKumaMonitorKeyword(t *testing.T) {
	normalized := normalizeConfig(&endpointmonitorv1alpha1.UptimeKumaConfig{
		MonitorType:   "keyword",
		KeywordValue:  "healthy",
		KeywordExists: "no",
	})

	mon := buildKumaMonitor(models.Monitor{Name: "test", URL: "https://example.com"}, normalized)
	keywordMonitor, ok := mon.(*monitor.HTTPKeyword)
	if !ok {
		t.Fatalf("expected a monitor.HTTPKeyword, got %T", mon)
	}
	if keywordMonitor.Keyword != "healthy" {
		t.Errorf("expected keyword healthy, got %q", keywordMonitor.Keyword)
	}
	if !keywordMonitor.InvertKeyword {
		t.Error("expected KeywordExists no to map to invertKeyword true")
	}
	if keywordMonitor.URL != "https://example.com" || keywordMonitor.Name != "test" {
		t.Errorf("monitor mapped incorrectly: %+v", keywordMonitor)
	}
}

// unmarshalBase builds a monitor.Base the same way the client does when it
// receives a monitor list, so the mapper can be tested offline
func unmarshalBase(t *testing.T, raw string) monitor.Base {
	t.Helper()
	base := monitor.Base{}
	if err := json.Unmarshal([]byte(raw), &base); err != nil {
		t.Fatalf("failed to unmarshal test monitor: %v", err)
	}
	return base
}

func TestKumaMonitorToBaseMonitorHttp(t *testing.T) {
	base := unmarshalBase(t, `{"id":7,"type":"http","name":"test","url":"https://example.com","interval":120,"active":true,"notificationIDList":{"3":true}}`)

	m := kumaMonitorToBaseMonitor(base)

	if m.ID != "7" || m.Name != "test" || m.URL != "https://example.com" {
		t.Errorf("monitor mapped incorrectly: %+v", m)
	}
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.UptimeKumaConfig)
	if providerConfig.Interval != 120 {
		t.Errorf("expected interval 120, got %d", providerConfig.Interval)
	}
	if providerConfig.MonitorType != "http" {
		t.Errorf("expected monitor type http, got %q", providerConfig.MonitorType)
	}
	if providerConfig.Notifications != "3" {
		t.Errorf("expected notifications 3, got %q", providerConfig.Notifications)
	}
}

func TestKumaMonitorToBaseMonitorKeyword(t *testing.T) {
	base := unmarshalBase(t, `{"id":9,"type":"keyword","name":"kw","url":"https://example.com","interval":60,"keyword":"healthy","invertKeyword":true}`)

	m := kumaMonitorToBaseMonitor(base)

	if m.ID != "9" || m.URL != "https://example.com" {
		t.Errorf("monitor mapped incorrectly: %+v", m)
	}
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.UptimeKumaConfig)
	if providerConfig.MonitorType != "keyword" {
		t.Errorf("expected monitor type keyword, got %q", providerConfig.MonitorType)
	}
	if providerConfig.KeywordValue != "healthy" {
		t.Errorf("expected keyword value healthy, got %q", providerConfig.KeywordValue)
	}
	if providerConfig.KeywordExists != "no" {
		t.Errorf("expected keyword exists no, got %q", providerConfig.KeywordExists)
	}
}

func TestEqual(t *testing.T) {
	service := UpTimeKumaMonitorService{}

	oldMonitor := models.Monitor{
		Name: "test",
		URL:  "https://example.com",
		Config: &endpointmonitorv1alpha1.UptimeKumaConfig{
			Interval:      120,
			MonitorType:   "keyword",
			KeywordValue:  "healthy",
			KeywordExists: "yes",
			Notifications: "5,1",
		},
	}

	if !service.Equal(oldMonitor, oldMonitor) {
		t.Error("identical monitors should be equal")
	}

	differentInterval := oldMonitor
	differentInterval.Config = &endpointmonitorv1alpha1.UptimeKumaConfig{
		Interval:      300,
		MonitorType:   "keyword",
		KeywordValue:  "healthy",
		KeywordExists: "yes",
		Notifications: "5,1",
	}
	if service.Equal(oldMonitor, differentInterval) {
		t.Error("different interval should not be equal")
	}

	differentURL := oldMonitor
	differentURL.URL = "https://other.example.com"
	if service.Equal(oldMonitor, differentURL) {
		t.Error("different URL should not be equal")
	}

	// A monitor read back from Kuma ("1,5" canonical) equals a CRD config using "5,1"
	readBack := models.Monitor{
		Name: "test",
		URL:  "https://example.com",
		Config: &endpointmonitorv1alpha1.UptimeKumaConfig{
			Interval:      120,
			MonitorType:   "keyword",
			KeywordValue:  "healthy",
			KeywordExists: "yes",
			Notifications: "1,5",
		},
	}
	if !service.Equal(readBack, oldMonitor) {
		t.Error("canonically equal notification lists should be equal")
	}
}

// Live integration tests below require a configured UptimeKuma provider in the
// test config; they skip silently when it is absent.

func setupTestService(t *testing.T) *UpTimeKumaMonitorService {
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if len(configFilePath) == 0 {
		configFilePath = "../../../.local/test-config.yaml"
	}
	if _, err := os.Stat(configFilePath); err != nil {
		// No test config available, skip the live test
		return nil
	}

	config := config.GetControllerConfigTest()
	provider := util.GetProviderWithName(config, "UptimeKuma")
	if provider == nil {
		return nil
	}
	service := UpTimeKumaMonitorService{}
	service.Setup(*provider)
	return &service
}

func TestUpTimeKumaMonitorLifecycle(t *testing.T) {
	service := setupTestService(t)
	if service == nil {
		return
	}

	m := models.Monitor{Name: "kuma-test", URL: "https://example.com"}
	service.Add(m)

	mRes, err := service.GetByName("kuma-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if mRes == nil || mRes.Name != m.Name || mRes.URL != m.URL {
		t.Errorf("monitor was not added correctly: %+v", mRes)
		return
	}
	defer service.Remove(*mRes)

	mRes.URL = "https://other.example.com"
	service.Update(*mRes)

	updated, err := service.GetByName("kuma-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}
	if updated == nil || updated.URL != "https://other.example.com" {
		t.Errorf("monitor was not updated correctly: %+v", updated)
	}
}
