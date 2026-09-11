package uptimekuma

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("uptime-kuma-monitor")

// Default Interval for status checking (Uptime Kuma minimum is 20 seconds)
const DefaultInterval = 60

const (
	connectTimeout = 30 * time.Second
	requestTimeout = 30 * time.Second
)

// UpTimeKumaMonitorService manages monitors on an Uptime Kuma server (Kuma 2.x)
// via its socket.io API. One client connection is held for the lifetime of the
// service and lazily re-established after failures.
type UpTimeKumaMonitorService struct {
	url      string
	username string
	password string
	client   *kuma.Client
	mutex    sync.Mutex
}

func (service *UpTimeKumaMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	if oldMonitor.URL != newMonitor.URL || !reflect.DeepEqual(normalizeConfig(oldMonitor.Config), normalizeConfig(newMonitor.Config)) {
		log.Info(fmt.Sprintf("There are some new changes in %s monitor", newMonitor.Name))
		return false
	}
	return true
}

func (service *UpTimeKumaMonitorService) Setup(p config.Provider) {
	service.url = p.ApiURL
	service.username = p.Username
	service.password = p.Password
	service.client = nil
}

// ensureClient returns the connected client, (re)connecting on first use or
// after a previous failure invalidated the connection
func (service *UpTimeKumaMonitorService) ensureClient() (*kuma.Client, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.client == nil {
		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		defer cancel()

		client, err := kuma.New(ctx, service.url, service.username, service.password,
			kuma.WithConnectTimeout(connectTimeout),
			kuma.WithLogLevel(kuma.LogLevelNone),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to Uptime Kuma at %s: %w", service.url, err)
		}
		service.client = client
	}

	return service.client, nil
}

// invalidateClient drops the connection so the next call reconnects
func (service *UpTimeKumaMonitorService) invalidateClient() {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.client != nil {
		if err := service.client.Disconnect(); err != nil {
			log.Error(err, "Failed to disconnect Uptime Kuma client")
		}
		service.client = nil
	}
}

func (service *UpTimeKumaMonitorService) GetAll() ([]models.Monitor, error) {
	client, err := service.ensureClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	kumaMonitors, err := client.GetMonitors(ctx)
	if err != nil {
		service.invalidateClient()
		return nil, fmt.Errorf("unable to get monitors from Uptime Kuma: %w", err)
	}

	monitors := make([]models.Monitor, 0, len(kumaMonitors))
	for _, kumaMonitor := range kumaMonitors {
		monitors = append(monitors, *kumaMonitorToBaseMonitor(kumaMonitor))
	}

	return monitors, nil
}

func (service *UpTimeKumaMonitorService) GetByName(name string) (*models.Monitor, error) {
	monitors, err := service.GetAll()
	if err != nil {
		return nil, err
	}

	// Uptime Kuma has no lookup by name, so match client-side
	for _, monitor := range monitors {
		if monitor.Name == name {
			return &monitor, nil
		}
	}

	return nil, nil
}

func (service *UpTimeKumaMonitorService) Add(m models.Monitor) {
	normalized := normalizeConfig(m.Config)

	client, err := service.ensureClient()
	if err != nil {
		log.Error(err, "Monitor couldn't be added: "+m.Name)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	_, err = client.CreateMonitor(ctx, buildKumaMonitor(m, normalized))
	if err != nil {
		service.invalidateClient()
		log.Error(err, "Monitor couldn't be added: "+m.Name)
		return
	}

	log.Info("Monitor Added: " + m.Name)
}

func (service *UpTimeKumaMonitorService) Update(m models.Monitor) {
	id, err := strconv.ParseInt(m.ID, 10, 64)
	if err != nil {
		log.Error(err, "Monitor couldn't be updated, invalid monitor id: "+m.ID)
		return
	}

	normalized := normalizeConfig(m.Config)

	client, err := service.ensureClient()
	if err != nil {
		log.Error(err, "Monitor couldn't be updated: "+m.Name)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// Uptime Kuma's editMonitor needs the full object, so fetch the current
	// monitor and override the fields managed by the controller
	current, err := client.GetMonitor(ctx, id)
	if err != nil {
		service.invalidateClient()
		log.Error(err, "Monitor couldn't be updated: "+m.Name)
		return
	}

	var updated monitor.Monitor
	if normalized.MonitorType == "keyword" {
		keywordMonitor := monitor.HTTPKeyword{}
		if err := current.As(&keywordMonitor); err != nil {
			log.Error(err, "Unable to read current keyword monitor: "+m.Name)
		}
		keywordMonitor.Name = m.Name
		keywordMonitor.URL = m.URL
		keywordMonitor.Interval = int64(normalized.Interval)
		keywordMonitor.NotificationIDs = parseNotificationIDs(normalized.Notifications)
		keywordMonitor.Keyword = normalized.KeywordValue
		keywordMonitor.InvertKeyword = keywordInvert(normalized.KeywordExists)
		updated = &keywordMonitor
	} else {
		httpMonitor := monitor.HTTP{}
		if err := current.As(&httpMonitor); err != nil {
			log.Error(err, "Unable to read current http monitor: "+m.Name)
		}
		httpMonitor.Name = m.Name
		httpMonitor.URL = m.URL
		httpMonitor.Interval = int64(normalized.Interval)
		httpMonitor.NotificationIDs = parseNotificationIDs(normalized.Notifications)
		updated = &httpMonitor
	}

	if err := client.UpdateMonitor(ctx, updated); err != nil {
		service.invalidateClient()
		log.Error(err, "Monitor couldn't be updated: "+m.Name)
		return
	}

	log.Info("Monitor Updated: " + m.Name)
}

func (service *UpTimeKumaMonitorService) Remove(m models.Monitor) {
	id, err := strconv.ParseInt(m.ID, 10, 64)
	if err != nil {
		log.Error(err, "Monitor couldn't be removed, invalid monitor id: "+m.ID)
		return
	}

	client, err := service.ensureClient()
	if err != nil {
		log.Error(err, "Monitor couldn't be removed: "+m.Name)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	if err := client.DeleteMonitor(ctx, id); err != nil {
		service.invalidateClient()
		log.Error(err, "Monitor couldn't be removed: "+m.Name)
		return
	}

	log.Info("Monitor Removed: " + m.Name)
}

// buildKumaMonitor builds a typed monitor for creation from the base monitor
// and the normalized provider config
func buildKumaMonitor(m models.Monitor, normalized *endpointmonitorv1alpha1.UptimeKumaConfig) monitor.Monitor {
	base := monitor.Base{
		Name:            m.Name,
		Interval:        int64(normalized.Interval),
		NotificationIDs: parseNotificationIDs(normalized.Notifications),
		IsActive:        true,
	}
	details := monitor.HTTPDetails{
		URL:                 m.URL,
		Method:              "GET",
		AcceptedStatusCodes: []string{"200-299"},
	}

	if normalized.MonitorType == "keyword" {
		if len(normalized.KeywordValue) == 0 {
			log.Error(nil, "Monitor is of type Keyword but the `keyword-value` is missing")
		}
		return &monitor.HTTPKeyword{
			Base:        base,
			HTTPDetails: details,
			HTTPKeywordDetails: monitor.HTTPKeywordDetails{
				Keyword:       normalized.KeywordValue,
				InvertKeyword: keywordInvert(normalized.KeywordExists),
			},
		}
	}

	return &monitor.HTTP{
		Base:        base,
		HTTPDetails: details,
	}
}

// kumaMonitorToBaseMonitor maps a monitor read from Uptime Kuma to the base
// monitor model, rebuilding the provider config for Equal() comparisons
func kumaMonitorToBaseMonitor(kumaMonitor monitor.Base) *models.Monitor {
	var m models.Monitor

	m.Name = kumaMonitor.Name
	m.ID = strconv.FormatInt(kumaMonitor.ID, 10)

	providerConfig := endpointmonitorv1alpha1.UptimeKumaConfig{
		Interval:      int(kumaMonitor.Interval),
		MonitorType:   kumaMonitor.Type(),
		Notifications: joinNotificationIDs(kumaMonitor.NotificationIDs),
	}

	if kumaMonitor.Type() == "keyword" {
		keywordMonitor := monitor.HTTPKeyword{}
		if err := kumaMonitor.As(&keywordMonitor); err == nil {
			m.URL = keywordMonitor.URL
			providerConfig.KeywordValue = keywordMonitor.Keyword
			providerConfig.KeywordExists = invertToKeywordExists(keywordMonitor.InvertKeyword)
		}
	} else {
		httpMonitor := monitor.HTTP{}
		if err := kumaMonitor.As(&httpMonitor); err == nil {
			m.URL = httpMonitor.URL
		}
	}

	m.Config = &providerConfig

	return &m
}

// normalizeConfig applies the provider defaults so monitors read back from
// Uptime Kuma can be compared against monitors built from the CRD
func normalizeConfig(providerConfig interface{}) *endpointmonitorv1alpha1.UptimeKumaConfig {
	normalized := &endpointmonitorv1alpha1.UptimeKumaConfig{
		Interval:      DefaultInterval,
		MonitorType:   "http",
		KeywordExists: "yes",
	}

	if kumaConfig, ok := providerConfig.(*endpointmonitorv1alpha1.UptimeKumaConfig); ok && kumaConfig != nil {
		if kumaConfig.Interval > 0 {
			normalized.Interval = kumaConfig.Interval
		}
		if len(kumaConfig.MonitorType) != 0 {
			normalized.MonitorType = strings.ToLower(kumaConfig.MonitorType)
		}
		if len(kumaConfig.KeywordExists) != 0 {
			normalized.KeywordExists = strings.ToLower(kumaConfig.KeywordExists)
		}
		normalized.KeywordValue = kumaConfig.KeywordValue
		normalized.Notifications = joinNotificationIDs(parseNotificationIDs(kumaConfig.Notifications))
	}

	return normalized
}

// parseNotificationIDs parses a comma- or dash-separated list of notification
// IDs into a sorted slice
func parseNotificationIDs(notifications string) []int64 {
	if notifications == "" {
		return nil
	}

	var ids []int64
	for _, part := range strings.FieldsFunc(notifications, func(r rune) bool { return r == ',' || r == '-' }) {
		if id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil {
			ids = append(ids, id)
		}
	}

	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func joinNotificationIDs(ids []int64) string {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, strconv.FormatInt(id, 10))
	}
	return strings.Join(strs, ",")
}

// keywordInvert maps KeywordExists to Uptime Kuma's invertKeyword flag:
// "no" means alert when the keyword does NOT exist
func keywordInvert(keywordExists string) bool {
	return strings.EqualFold(keywordExists, "no")
}

// invertToKeywordExists maps Uptime Kuma's invertKeyword flag back to KeywordExists
func invertToKeywordExists(invertKeyword bool) string {
	if invertKeyword {
		return "no"
	}
	return "yes"
}
