package uptimekuma

import (
	"context"
	"errors"
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

	// Defaults matching the Uptime Kuma UI. Uptime Kuma rejects a monitor
	// without a retry interval ("Retry interval cannot be less than 1 seconds")
	// and treats a timeout of 0 as "wait interval x 800 seconds", which would
	// keep a hanging endpoint UP forever.
	defaultMaxRedirects  = 10
	requestTimeoutFactor = 0.8

	// Uptime Kuma only allows 20 logins per minute for all clients together, so
	// a failed connection must not be retried on the next reconcile
	initialReconnectBackoff = 5 * time.Second
	maxReconnectBackoff     = 5 * time.Minute

	// Bad credentials never fix themselves, so back off hard instead of
	// repeatedly logging in and eventually locking out the Kuma UI as well
	authErrorBackoff = 5 * time.Minute
)

// UpTimeKumaMonitorService manages monitors on an Uptime Kuma server (Kuma 2.x)
// via its socket.io API. One client connection is held for the lifetime of the
// service and lazily re-established after failures, honouring a backoff so a
// broken or misconfigured server does not exhaust Uptime Kuma's login limit.
type UpTimeKumaMonitorService struct {
	url      string
	username string
	password string

	client      *kuma.Client
	mutex       sync.Mutex
	reconnectAt time.Time
	backoff     time.Duration
}

func (service *UpTimeKumaMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	if oldMonitor.URL != newMonitor.URL || !reflect.DeepEqual(normalizeConfig(oldMonitor.Config), normalizeConfig(newMonitor.Config)) {
		log.Info(fmt.Sprintf("There are some new changes in %s monitor", newMonitor.Name))
		return false
	}
	return true
}

func (service *UpTimeKumaMonitorService) Setup(p config.Provider) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.url = p.ApiURL
	service.username = p.Username
	service.password = p.Password
	service.client = nil
	service.backoff = 0
	service.reconnectAt = time.Time{}
}

// ensureClient returns the connected client, (re)connecting on first use or
// after a previous failure invalidated the connection
func (service *UpTimeKumaMonitorService) ensureClient() (*kuma.Client, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.client != nil {
		return service.client, nil
	}

	if wait := time.Until(service.reconnectAt); wait > 0 {
		return nil, fmt.Errorf("not reconnecting to Uptime Kuma at %s yet, retrying in %s: the previous connection failed",
			service.url, wait.Round(time.Second))
	}

	// The client keeps the context it was created with for the whole lifetime of
	// the socket connection, so this must not be cancelled when this function
	// returns. WithConnectTimeout bounds the connection attempt itself.
	client, err := kuma.New(context.Background(), service.url, service.username, service.password,
		kuma.WithConnectTimeout(connectTimeout),
		kuma.WithLogLevel(kuma.LogLevelNone),
	)
	if err != nil {
		service.registerFailure(err)
		return nil, fmt.Errorf("unable to connect to Uptime Kuma at %s: %w", service.url, err)
	}

	service.client = client
	service.backoff = 0
	service.reconnectAt = time.Time{}

	return service.client, nil
}

// invalidateClient drops the connection so the next call reconnects, and starts
// the reconnect backoff
func (service *UpTimeKumaMonitorService) invalidateClient(err error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.client != nil {
		if disconnectErr := service.client.Disconnect(); disconnectErr != nil {
			log.Error(disconnectErr, "Failed to disconnect Uptime Kuma client")
		}
		service.client = nil
	}

	service.registerFailure(err)
}

// registerFailure records a failed connection attempt and the earliest time a
// new one may be made. Callers must hold the mutex.
func (service *UpTimeKumaMonitorService) registerFailure(err error) {
	if isAuthError(err) {
		service.backoff = authErrorBackoff
		log.Error(err, fmt.Sprintf("Authentication with Uptime Kuma at %s failed, not retrying for %s. Check the configured username and password, and that 2FA is disabled for that account",
			service.url, service.backoff))
	} else {
		if service.backoff == 0 {
			service.backoff = initialReconnectBackoff
		} else {
			service.backoff *= 2
		}
		if service.backoff > maxReconnectBackoff {
			service.backoff = maxReconnectBackoff
		}
		log.Info(fmt.Sprintf("Connection to Uptime Kuma at %s lost, not reconnecting for %s", service.url, service.backoff))
	}

	service.reconnectAt = time.Now().Add(service.backoff)
}

// isAuthError reports whether connecting failed because Uptime Kuma rejected the
// credentials, in which case reconnecting right away would only burn through the
// login rate limit.
func isAuthError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "incorrect username or password") ||
		strings.Contains(msg, "autherrincorrectcreds") ||
		strings.Contains(msg, "login:")
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
		service.invalidateClient(err)
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

	kumaMonitor, err := buildKumaMonitor(m, normalized)
	if err != nil {
		log.Error(err, "Monitor couldn't be added: "+m.Name)
		return
	}

	client, err := service.ensureClient()
	if err != nil {
		log.Error(err, "Monitor couldn't be added: "+m.Name)
		return
	}

	notificationIDs := parseNotificationIDs(normalized.Notifications)
	if err := validateNotificationIDs(client, notificationIDs); err != nil {
		log.Error(err, "Monitor couldn't be added: "+m.Name)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	if _, err := client.CreateMonitor(ctx, kumaMonitor); err != nil {
		service.invalidateClient(err)
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

	if normalized.MonitorType == "keyword" && len(normalized.KeywordValue) == 0 {
		log.Error(nil, "Monitor couldn't be updated, it is of type keyword but the `keywordValue` is missing: "+m.Name)
		return
	}

	client, err := service.ensureClient()
	if err != nil {
		log.Error(err, "Monitor couldn't be updated: "+m.Name)
		return
	}

	notificationIDs := parseNotificationIDs(normalized.Notifications)
	if err := validateNotificationIDs(client, notificationIDs); err != nil {
		log.Error(err, "Monitor couldn't be updated: "+m.Name)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// Uptime Kuma's editMonitor needs the full object, so fetch the current
	// monitor and override the fields managed by the controller
	current, err := client.GetMonitor(ctx, id)
	if err != nil {
		service.invalidateClient(err)
		log.Error(err, "Monitor couldn't be updated: "+m.Name)
		return
	}

	var updated monitor.Monitor
	if normalized.MonitorType == "keyword" {
		keywordMonitor := monitor.HTTPKeyword{}
		if err := current.As(&keywordMonitor); err != nil {
			log.Error(err, "Unable to read current keyword monitor: "+m.Name)
			return
		}
		keywordMonitor.Name = m.Name
		keywordMonitor.URL = m.URL
		keywordMonitor.Interval = int64(normalized.Interval)
		keywordMonitor.RetryInterval = int64(normalized.Interval)
		keywordMonitor.MaxRedirects = defaultMaxRedirects
		keywordMonitor.Timeout = requestTimeoutFor(normalized.Interval)
		keywordMonitor.NotificationIDs = notificationIDs
		keywordMonitor.Keyword = normalized.KeywordValue
		keywordMonitor.InvertKeyword = keywordInvert(normalized.KeywordExists)
		updated = &keywordMonitor
	} else {
		httpMonitor := monitor.HTTP{}
		if err := current.As(&httpMonitor); err != nil {
			log.Error(err, "Unable to read current http monitor: "+m.Name)
			return
		}
		httpMonitor.Name = m.Name
		httpMonitor.URL = m.URL
		httpMonitor.Interval = int64(normalized.Interval)
		httpMonitor.RetryInterval = int64(normalized.Interval)
		httpMonitor.MaxRedirects = defaultMaxRedirects
		httpMonitor.Timeout = requestTimeoutFor(normalized.Interval)
		httpMonitor.NotificationIDs = notificationIDs
		updated = &httpMonitor
	}

	if err := client.UpdateMonitor(ctx, updated); err != nil {
		service.invalidateClient(err)
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
		service.invalidateClient(err)
		log.Error(err, "Monitor couldn't be removed: "+m.Name)
		return
	}

	log.Info("Monitor Removed: " + m.Name)
}

// buildKumaMonitor builds a typed monitor for creation from the base monitor
// and the normalized provider config
func buildKumaMonitor(m models.Monitor, normalized *endpointmonitorv1alpha1.UptimeKumaConfig) (monitor.Monitor, error) {
	base := monitor.Base{
		Name:            m.Name,
		Interval:        int64(normalized.Interval),
		RetryInterval:   int64(normalized.Interval),
		NotificationIDs: parseNotificationIDs(normalized.Notifications),
		IsActive:        true,
	}
	details := monitor.HTTPDetails{
		URL:                 m.URL,
		Method:              "GET",
		AcceptedStatusCodes: []string{"200-299"},
		MaxRedirects:        defaultMaxRedirects,
		Timeout:             requestTimeoutFor(normalized.Interval),
	}

	if normalized.MonitorType == "keyword" {
		if len(normalized.KeywordValue) == 0 {
			// Creating the monitor with an empty keyword would leave a monitor
			// that is UP or DOWN for the wrong reason
			return nil, errors.New("monitor is of type keyword but the `keywordValue` is missing")
		}
		return &monitor.HTTPKeyword{
			Base:        base,
			HTTPDetails: details,
			HTTPKeywordDetails: monitor.HTTPKeywordDetails{
				Keyword:       normalized.KeywordValue,
				InvertKeyword: keywordInvert(normalized.KeywordExists),
			},
		}, nil
	}

	return &monitor.HTTP{
		Base:        base,
		HTTPDetails: details,
	}, nil
}

// requestTimeoutFor returns the per-request timeout for a check interval, using
// the same 80% ratio as the Uptime Kuma UI. Without it Uptime Kuma waits
// interval x 800 seconds, so a hanging endpoint would never go down.
func requestTimeoutFor(interval int) int64 {
	return int64(float64(interval) * requestTimeoutFactor)
}

// validateNotificationIDs checks the configured notification IDs against the
// ones the connected account can see. Uptime Kuma rejects a monitor with an
// unknown notification ID with a foreign key error but still creates it without
// any notification, which made the controller retry a failing update on every
// reconcile.
func validateNotificationIDs(client *kuma.Client, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	known := make(map[int64]bool)
	for _, notification := range client.GetNotifications(context.Background()) {
		known[notification.GetID()] = true
	}

	missing := make([]string, 0, len(ids))
	for _, id := range ids {
		if !known[id] {
			missing = append(missing, strconv.FormatInt(id, 10))
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("unknown Uptime Kuma notification id(s): %s, find the available ids in Uptime Kuma under Settings > Notifications",
			strings.Join(missing, ", "))
	}

	return nil
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

// keywordInvert maps KeywordExists to Uptime Kuma's invertKeyword flag.
//
// Uptime Kuma reports a keyword monitor as UP when the keyword was found and
// invertKeyword is false (`keywordFound === !invertKeyword` in its check
// handler), so:
//
//	keywordExists: "yes" (default) -> UP while the keyword is in the response
//	keywordExists: "no"           -> UP only while the keyword is absent, which
//	                                 alerts as soon as the keyword shows up
//	                                 (e.g. keywordValue: "404")
//
// This matches the UptimeRobot provider's keyword_type 1 and 2.
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
