// Package betterstack adds Better Stack (betterstack.com) support to
// IngressMonitorController.
//
// Better Stack exposes a plain JSON REST API under /api/v2/monitors with a
// bearer token, so this provider talks to it directly rather than through a
// vendor SDK — there is no maintained Go client for it.
package betterstack

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/http"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

const (
	// DefaultApiURL is Better Stack's public API root. Overridable through the
	// provider's apiURL so the whole surface can be pointed at a test double.
	DefaultApiURL = "https://uptime.betterstack.com"

	monitorsPath = "/api/v2/monitors"

	// Better Stack's own default when check_frequency is omitted. Stated
	// explicitly so a monitor's interval does not silently change if they move
	// the default.
	DefaultCheckFrequency = 180

	// A monitor created from an Ingress is an HTTP(S) check unless the CR says
	// otherwise. "status" means "up if the status code matches", which is what
	// every other provider in this repo defaults to.
	DefaultMonitorType = "status"

	// Better Stack paginates at 50 by default; ask for the maximum so GetAll
	// walks as few pages as it can.
	pageSize = 250

	// Upper bound on pages walked in one GetAll, so a misbehaving API cannot
	// hang a reconcile. 250 * 200 is far beyond any real account.
	maxPages = 200
)

var log = logf.Log.WithName("betterstack")

// BetterStackMonitorService is the MonitorService implementation for Better Stack.
type BetterStackMonitorService struct {
	apiToken string
	url      string
}

// monitorAttributes mirrors the subset of Better Stack's monitor object this
// controller manages. Fields are pointers so that an unset value is omitted
// entirely rather than sent as a zero — Better Stack treats an explicit null
// as "clear this", which would wipe settings made in their UI.
type monitorAttributes struct {
	URL                 *string   `json:"url,omitempty"`
	PronounceableName   *string   `json:"pronounceable_name,omitempty"`
	MonitorType         *string   `json:"monitor_type,omitempty"`
	CheckFrequency      *int      `json:"check_frequency,omitempty"`
	ExpectedStatusCodes *[]int    `json:"expected_status_codes,omitempty"`
	RequiredKeyword     *string   `json:"required_keyword,omitempty"`
	VerifySSL           *bool     `json:"verify_ssl,omitempty"`
	FollowRedirects     *bool     `json:"follow_redirects,omitempty"`
	RememberCookies     *bool     `json:"remember_cookies,omitempty"`
	RequestTimeout      *int      `json:"request_timeout,omitempty"`
	ConfirmationPeriod  *int      `json:"confirmation_period,omitempty"`
	RecoveryPeriod      *int      `json:"recovery_period,omitempty"`
	Regions             *[]string `json:"regions,omitempty"`
	PolicyID            *string   `json:"policy_id,omitempty"`
	Paused              *bool     `json:"paused,omitempty"`
	Email               *bool     `json:"email,omitempty"`
	SMS                 *bool     `json:"sms,omitempty"`
	Call                *bool     `json:"call,omitempty"`
	Push                *bool     `json:"push,omitempty"`
	TeamWait            *int      `json:"team_wait,omitempty"`
}

type monitorData struct {
	ID         string            `json:"id"`
	Attributes monitorAttributes `json:"attributes"`
}

type monitorResponse struct {
	Data monitorData `json:"data"`
}

type monitorListResponse struct {
	Data       []monitorData `json:"data"`
	Pagination struct {
		Next string `json:"next"`
	} `json:"pagination"`
}

func (s *BetterStackMonitorService) Setup(p config.Provider) {
	s.apiToken = p.ApiToken
	// The generic `apiKey` field is accepted as a fallback: several existing
	// providers in this repo use it, and an operator moving between them should
	// not have to know which name a given provider reads.
	if s.apiToken == "" {
		s.apiToken = p.ApiKey
	}

	s.url = strings.TrimSuffix(p.ApiURL, "/")
	if s.url == "" {
		s.url = DefaultApiURL
	}

	if s.apiToken == "" {
		log.Error(nil, "Better Stack provider configured without an API token; every request will be rejected")
	}
	log.Info("Better Stack monitor has been initialized")
}

func (s *BetterStackMonitorService) headers() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + s.apiToken,
		"Content-Type":  "application/json",
	}
}

// GetAll walks every page rather than returning only the first. GetByName is
// built on it, so stopping at page one would silently report an existing
// monitor as missing and the controller would create a duplicate.
func (s *BetterStackMonitorService) GetAll() ([]models.Monitor, error) {
	var monitors []models.Monitor

	// Pages are requested by number rather than by following pagination.next,
	// which is only used as a "there is more" flag. Better Stack returns next
	// as an absolute URL today, but the official Terraform provider pages by
	// number — following the URL would break silently if they ever return a
	// relative one, and this cannot.
	for page := 1; ; page++ {
		url := fmt.Sprintf("%s%s?page=%d&per_page=%d", s.url, monitorsPath, page, pageSize)
		client := http.CreateHttpClient(url)
		response := client.GetUrl(s.headers(), []byte{})

		if response.StatusCode != 200 {
			return nil, fmt.Errorf("betterstack: list monitors returned %d: %s",
				response.StatusCode, truncate(string(response.Bytes)))
		}

		var listResponse monitorListResponse
		if err := json.Unmarshal(response.Bytes, &listResponse); err != nil {
			return nil, fmt.Errorf("betterstack: cannot decode monitor list: %w", err)
		}

		for _, data := range listResponse.Data {
			monitors = append(monitors, toBaseMonitor(data))
		}

		if listResponse.Pagination.Next == "" {
			break
		}

		// Defensive stop: an API that always reports a next page would
		// otherwise spin forever inside a reconcile.
		if page >= maxPages {
			log.Error(nil, fmt.Sprintf("betterstack: stopped paging monitors after %d pages", maxPages))
			break
		}
	}

	return monitors, nil
}

// GetByName matches on the monitor's pronounceable_name, which is where Add
// puts the controller's monitor name. Better Stack has no name-filtered list
// endpoint, so this pages through all monitors.
func (s *BetterStackMonitorService) GetByName(name string) (*models.Monitor, error) {
	monitors, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	for _, monitor := range monitors {
		if monitor.Name == name {
			found := monitor
			return &found, nil
		}
	}

	// Not an error: the controller calls this to decide between Add and Update,
	// and "no such monitor" is the ordinary create path.
	return nil, nil
}

func (s *BetterStackMonitorService) Add(m models.Monitor) {
	body, err := json.Marshal(s.buildAttributes(m, true))
	if err != nil {
		log.Error(err, "betterstack: cannot encode monitor "+m.Name)
		return
	}

	client := http.CreateHttpClient(s.url + monitorsPath)
	response := client.PostUrl(s.headers(), body)

	// 201 is documented; 200 is accepted too so a change on their side does not
	// turn a successful create into an error log and a retry loop.
	if response.StatusCode != 201 && response.StatusCode != 200 {
		log.Error(nil, fmt.Sprintf("betterstack: create monitor %s returned %d: %s",
			m.Name, response.StatusCode, truncate(string(response.Bytes))))
		return
	}

	log.Info("Monitor " + m.Name + " has been added")
}

func (s *BetterStackMonitorService) Update(m models.Monitor) {
	if m.ID == "" {
		log.Error(nil, "betterstack: cannot update monitor "+m.Name+" without an id")
		return
	}

	// PATCH, not PUT: Better Stack replaces the whole resource on PUT, which
	// would clear every field this controller does not manage — escalation
	// policies, maintenance windows and regions set in their UI.
	body, err := json.Marshal(s.buildAttributes(m, false))
	if err != nil {
		log.Error(err, "betterstack: cannot encode monitor "+m.Name)
		return
	}

	client := http.CreateHttpClient(s.url + monitorsPath + "/" + m.ID)
	response := client.RequestWithHeaders("PATCH", body, s.headers())

	if response.StatusCode != 200 {
		log.Error(nil, fmt.Sprintf("betterstack: update monitor %s returned %d: %s",
			m.Name, response.StatusCode, truncate(string(response.Bytes))))
		return
	}

	log.Info("Monitor " + m.Name + " has been updated")
}

func (s *BetterStackMonitorService) Remove(m models.Monitor) {
	if m.ID == "" {
		log.Error(nil, "betterstack: cannot remove monitor "+m.Name+" without an id")
		return
	}

	client := http.CreateHttpClient(s.url + monitorsPath + "/" + m.ID)
	response := client.DeleteUrl(s.headers(), []byte{})

	// 204 is the documented success. 404 is treated as success as well: the
	// monitor is gone, which is the requested end state, and erroring would
	// leave the controller retrying a delete that can never succeed.
	if response.StatusCode != 204 && response.StatusCode != 200 && response.StatusCode != 404 {
		log.Error(nil, fmt.Sprintf("betterstack: delete monitor %s returned %d: %s",
			m.Name, response.StatusCode, truncate(string(response.Bytes))))
		return
	}

	log.Info("Monitor " + m.Name + " has been deleted")
}

// Equal reports whether the live monitor already matches the desired one. It
// compares only what this controller sets, so a field changed in Better Stack's
// UI that the CR says nothing about does not cause a permanent update loop.
func (s *BetterStackMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	if oldMonitor.URL != newMonitor.URL || oldMonitor.Name != newMonitor.Name {
		return false
	}

	oldConfig := getConfig(oldMonitor)
	newConfig := getConfig(newMonitor)

	// Both unconfigured: URL and name already matched, so nothing differs.
	if oldConfig == nil && newConfig == nil {
		return true
	}
	if oldConfig == nil || newConfig == nil {
		return false
	}

	return oldConfig.CheckFrequency == newConfig.CheckFrequency &&
		oldConfig.MonitorType == newConfig.MonitorType &&
		oldConfig.ExpectedStatusCodes == newConfig.ExpectedStatusCodes &&
		oldConfig.RequiredKeyword == newConfig.RequiredKeyword &&
		oldConfig.Paused == newConfig.Paused &&
		oldConfig.Email == newConfig.Email &&
		oldConfig.SMS == newConfig.SMS &&
		oldConfig.Call == newConfig.Call &&
		oldConfig.Push == newConfig.Push &&
		oldConfig.PolicyID == newConfig.PolicyID &&
		oldConfig.Regions == newConfig.Regions &&
		oldConfig.VerifySSL == newConfig.VerifySSL &&
		oldConfig.FollowRedirects == newConfig.FollowRedirects &&
		oldConfig.RememberCookies == newConfig.RememberCookies &&
		oldConfig.RequestTimeout == newConfig.RequestTimeout &&
		oldConfig.ConfirmationPeriod == newConfig.ConfirmationPeriod &&
		oldConfig.RecoveryPeriod == newConfig.RecoveryPeriod &&
		oldConfig.TeamWait == newConfig.TeamWait
}

// buildAttributes maps a Monitor plus its BetterStackConfig onto the API's
// shape. `create` marks the call that must carry defaults: on update, an
// omitted field is left as-is rather than reset.
func (s *BetterStackMonitorService) buildAttributes(m models.Monitor, create bool) monitorAttributes {
	attributes := monitorAttributes{
		URL:               strPtr(m.URL),
		PronounceableName: strPtr(m.Name),
	}

	providerConfig := getConfig(m)

	if providerConfig != nil && providerConfig.MonitorType != "" {
		attributes.MonitorType = strPtr(providerConfig.MonitorType)
	} else if create {
		attributes.MonitorType = strPtr(DefaultMonitorType)
	}

	if providerConfig != nil && providerConfig.CheckFrequency > 0 {
		attributes.CheckFrequency = intPtr(providerConfig.CheckFrequency)
	} else if create {
		attributes.CheckFrequency = intPtr(DefaultCheckFrequency)
	}

	if providerConfig == nil {
		return attributes
	}

	codes := parseStatusCodes(providerConfig.ExpectedStatusCodes)
	if len(codes) > 0 {
		attributes.ExpectedStatusCodes = &codes
	}
	if providerConfig.RequiredKeyword != "" {
		attributes.RequiredKeyword = strPtr(providerConfig.RequiredKeyword)
	}
	if providerConfig.PolicyID != "" {
		attributes.PolicyID = strPtr(providerConfig.PolicyID)
	}
	if regions := splitAndTrim(providerConfig.Regions); len(regions) > 0 {
		attributes.Regions = &regions
	}
	if providerConfig.RequestTimeout > 0 {
		attributes.RequestTimeout = intPtr(providerConfig.RequestTimeout)
	}
	if providerConfig.ConfirmationPeriod > 0 {
		attributes.ConfirmationPeriod = intPtr(providerConfig.ConfirmationPeriod)
	}
	if providerConfig.RecoveryPeriod > 0 {
		// Better Stack accepts only a fixed set here and 422s on anything else,
		// so a typo would otherwise mean no monitor at all.
		if isValidRecoveryPeriod(providerConfig.RecoveryPeriod) {
			attributes.RecoveryPeriod = intPtr(providerConfig.RecoveryPeriod)
		} else {
			log.Error(nil, fmt.Sprintf(
				"betterstack: ignoring recoveryPeriod %d; valid values are %v",
				providerConfig.RecoveryPeriod, validRecoveryPeriods))
		}
	}
	if providerConfig.TeamWait > 0 {
		attributes.TeamWait = intPtr(providerConfig.TeamWait)
	}

	// Booleans are tri-state in the CRD (unset / "true" / "false") so that not
	// naming one leaves Better Stack's own default or an operator's UI change
	// alone, instead of forcing Go's false.
	attributes.VerifySSL = parseBool(providerConfig.VerifySSL)
	attributes.FollowRedirects = parseBool(providerConfig.FollowRedirects)
	attributes.RememberCookies = parseBool(providerConfig.RememberCookies)

	// Better Stack rejects a monitor that both expects a 3xx status and follows
	// redirects or keeps cookies — "Cannot follow redirects when expecting a 3xx
	// status code" (verified against the live API 2026-09-10). A redirect
	// monitor is a real use case: probing an apex that 302s to the app is how
	// you check the redirect itself still works. Rather than let that fail at
	// create time with an error only visible in the operator log, turn both off
	// when the CR expects a 3xx and has not said otherwise.
	if expectsRedirectStatus(codes) {
		if attributes.FollowRedirects == nil {
			attributes.FollowRedirects = boolPtr(false)
		}
		if attributes.RememberCookies == nil {
			attributes.RememberCookies = boolPtr(false)
		}
	}
	attributes.Paused = parseBool(providerConfig.Paused)
	attributes.Email = parseBool(providerConfig.Email)
	attributes.SMS = parseBool(providerConfig.SMS)
	attributes.Call = parseBool(providerConfig.Call)
	attributes.Push = parseBool(providerConfig.Push)

	return attributes
}

func toBaseMonitor(data monitorData) models.Monitor {
	monitor := models.Monitor{ID: data.ID}
	if data.Attributes.URL != nil {
		monitor.URL = *data.Attributes.URL
	}
	if data.Attributes.PronounceableName != nil {
		monitor.Name = *data.Attributes.PronounceableName
	}
	return monitor
}

func getConfig(m models.Monitor) *endpointmonitorv1alpha1.BetterStackConfig {
	if m.Config == nil {
		return nil
	}
	providerConfig, ok := m.Config.(*endpointmonitorv1alpha1.BetterStackConfig)
	if !ok {
		return nil
	}
	return providerConfig
}

// parseStatusCodes turns "200,201,302" into []int. Unparseable entries are
// dropped with a log rather than failing the whole monitor: one typo should not
// stop the endpoint being watched at all.
func parseStatusCodes(raw string) []int {
	var codes []int
	for _, part := range splitAndTrim(raw) {
		code, err := strconv.Atoi(part)
		if err != nil {
			log.Error(err, "betterstack: ignoring unparseable expected status code "+part)
			continue
		}
		codes = append(codes, code)
	}
	return codes
}

func splitAndTrim(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// parseBool keeps the unset case distinct from false: "" yields nil, so the
// field is omitted from the request entirely.
func parseBool(raw string) *bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true":
		return boolPtr(true)
	case "false":
		return boolPtr(false)
	default:
		return nil
	}
}

// truncate keeps an unexpected HTML error page or a large payload from filling
// the log with one line.
func truncate(body string) string {
	const max = 300
	if len(body) <= max {
		return body
	}
	return body[:max] + "…"
}

// validRecoveryPeriods is the set Better Stack accepts; anything else is a 422.
var validRecoveryPeriods = []int{0, 60, 180, 300, 900, 1800, 3600, 7200}

func isValidRecoveryPeriod(v int) bool {
	for _, allowed := range validRecoveryPeriods {
		if v == allowed {
			return true
		}
	}
	return false
}

// expectsRedirectStatus reports whether any expected status code is a 3xx, in
// which case Better Stack forbids following redirects or keeping cookies.
func expectsRedirectStatus(codes []int) bool {
	for _, code := range codes {
		if code >= 300 && code < 400 {
			return true
		}
	}
	return false
}

func strPtr(v string) *string { return &v }
func intPtr(v int) *int       { return &v }
func boolPtr(v bool) *bool    { return &v }
