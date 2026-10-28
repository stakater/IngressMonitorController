package uptimerobot

import (
	"encoding/json"
	"fmt"
	Http "net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/http"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

type UpTimeMonitorService struct {
	apiKey            string
	url               string
	alertContacts     string
	statusPageService UpTimeStatusPageService
}

// Default Interval for status checking
const DefaultInterval = 300

// DefaultTimeout is sent on every create/update; timeout is required by API v3
const DefaultTimeout = 30

const maxRateLimitRetries = 3

func (monitor *UpTimeMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	if !reflect.DeepEqual(monitor.processProviderConfig(oldMonitor), monitor.processProviderConfig(newMonitor)) {
		log.Info(fmt.Sprintf("There are some new changes in %s monitor", newMonitor.Name))
		return false
	}
	return true
}

func (monitor *UpTimeMonitorService) Setup(p config.Provider) {
	monitor.apiKey = p.ApiKey
	monitor.url = p.ApiURL
	monitor.alertContacts = p.AlertContacts
	monitor.statusPageService = UpTimeStatusPageService{}
	monitor.statusPageService.Setup(p)
}

func (monitor *UpTimeMonitorService) requestHeaders() map[string]string {
	return v3Headers(monitor.apiKey)
}

// v3Headers returns the headers required by API v3: Bearer auth + JSON bodies
func v3Headers(apiKey string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Content-Type":  "application/json",
	}
}

func apiURL(base string, path string) string {
	return strings.TrimSuffix(base, "/") + path
}

// doRequestWithRetries performs a request against API v3 and retries a bounded
// number of times on 429, honoring the Retry-After header
func doRequestWithRetries(method string, requestURL string, headers map[string]string, body []byte) http.HttpResponse {
	for attempt := 0; ; attempt++ {
		client := http.CreateHttpClient(requestURL)
		response := client.RequestWithHeaders(method, body, headers)
		if response.StatusCode != Http.StatusTooManyRequests || attempt >= maxRateLimitRetries {
			return response
		}
		delay := 10 * time.Second
		if retryAfter := response.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				delay = time.Duration(seconds) * time.Second
			}
		}
		log.Info("UptimeRobot rate limit hit, retrying after delay", "delay_seconds", int(delay.Seconds()), "attempt", attempt+1, "max_retries", maxRateLimitRetries)
		time.Sleep(delay)
	}
}

func errorExcerpt(bytes []byte) string {
	body := strings.TrimSpace(string(bytes))
	if len(body) > 200 {
		body = body[:200]
	}
	return body
}

func isSuccess(response http.HttpResponse) bool {
	return response.StatusCode >= Http.StatusOK && response.StatusCode <= 299
}

// paginate fetches all pages of a v3 cursor-paginated list endpoint,
// following nextLink until it is null
func paginate[T any](requestURL string, headers map[string]string) ([]T, error) {
	var results []T
	for {
		response := doRequestWithRetries("GET", requestURL, headers, nil)
		if !isSuccess(response) {
			return nil, fmt.Errorf("Request failed. Status Code: %d. Error: %s", response.StatusCode, errorExcerpt(response.Bytes))
		}
		var page struct {
			NextLink *string `json:"nextLink"`
			Data     []T     `json:"data"`
		}
		if err := json.Unmarshal(response.Bytes, &page); err != nil {
			return nil, fmt.Errorf("Unable to unmarshal paginated response: %w", err)
		}
		results = append(results, page.Data...)
		if page.NextLink == nil || *page.NextLink == "" {
			return results, nil
		}
		requestURL = *page.NextLink
	}
}

// getMonitors lists monitors, following nextLink pagination until exhausted
func (monitor *UpTimeMonitorService) getMonitors(query url.Values) ([]UptimeMonitorMonitor, error) {
	query.Set("limit", "200")
	monitors, err := paginate[UptimeMonitorMonitor](apiURL(monitor.url, "/monitors")+"?"+query.Encode(), monitor.requestHeaders())
	if err != nil {
		return nil, fmt.Errorf("GetMonitors request failed: %w", err)
	}
	return monitors, nil
}

func (monitor *UpTimeMonitorService) GetByName(name string) (*models.Monitor, error) {
	monitors, err := monitor.getMonitors(url.Values{"name": []string{name}})
	if err != nil {
		return nil, err
	}

	for _, m := range monitors {
		if m.FriendlyName == name {
			return UptimeMonitorMonitorToBaseMonitorMapper(m), nil
		}
	}

	return nil, nil
}

func (monitor *UpTimeMonitorService) GetAllByName(name string) ([]models.Monitor, error) {
	monitors, err := monitor.getMonitors(url.Values{"name": []string{name}})
	if err != nil {
		return nil, err
	}

	return UptimeMonitorMonitorsToBaseMonitorsMapper(monitors), nil
}

func (monitor *UpTimeMonitorService) GetAll() ([]models.Monitor, error) {
	monitors, err := monitor.getMonitors(url.Values{})
	if err != nil {
		return nil, err
	}

	return UptimeMonitorMonitorsToBaseMonitorsMapper(monitors), nil
}

func (monitor *UpTimeMonitorService) Add(m models.Monitor) {
	requestBody, err := json.Marshal(monitor.processProviderConfig(m))
	if err != nil {
		log.Error(err, "Monitor couldn't be added: "+m.Name)
		return
	}

	response := doRequestWithRetries("POST", apiURL(monitor.url, "/monitors"), monitor.requestHeaders(), requestBody)

	if isSuccess(response) {
		var created UptimeMonitorMonitor
		if err := json.Unmarshal(response.Bytes, &created); err != nil {
			log.Error(err, "Monitor couldn't be added: "+m.Name)
			return
		}
		log.Info("Monitor Added: " + m.Name)
		monitor.handleStatusPagesConfig(m, strconv.Itoa(created.ID))
	} else {
		log.Info("Monitor couldn't be added: " + m.Name + ". Status Code: " + strconv.Itoa(response.StatusCode) + ". Error: " + errorExcerpt(response.Bytes))
	}
}

func (monitor *UpTimeMonitorService) Update(m models.Monitor) {
	requestBody, err := json.Marshal(monitor.processProviderConfig(m))
	if err != nil {
		log.Error(err, "Monitor couldn't be updated: "+m.Name)
		return
	}

	response := doRequestWithRetries("PATCH", apiURL(monitor.url, "/monitors/"+m.ID), monitor.requestHeaders(), requestBody)

	if isSuccess(response) {
		var updated UptimeMonitorMonitor
		if err := json.Unmarshal(response.Bytes, &updated); err != nil {
			log.Error(err, "Monitor couldn't be updated: "+m.Name)
			return
		}
		monitorId := m.ID
		if updated.ID != 0 {
			monitorId = strconv.Itoa(updated.ID)
		}
		log.Info("Monitor Updated: " + m.Name)
		monitor.handleStatusPagesConfig(m, monitorId)
	} else {
		log.Info("Monitor couldn't be updated: " + m.Name + ". Status Code: " + strconv.Itoa(response.StatusCode) + ". Error: " + errorExcerpt(response.Bytes))
	}
}

func (monitor *UpTimeMonitorService) processProviderConfig(m models.Monitor) *UptimeMonitorMonitorRequest {
	request := &UptimeMonitorMonitorRequest{
		FriendlyName: m.Name,
		URL:          m.URL,
		Timeout:      DefaultTimeout,
	}

	// Retrieve provider configuration
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.UptimeRobotConfig)

	interval := DefaultInterval
	if providerConfig != nil && providerConfig.Interval > 0 {
		interval = providerConfig.Interval
	}
	request.Interval = interval

	alertContacts := monitor.alertContacts
	if providerConfig != nil && len(providerConfig.AlertContacts) != 0 {
		alertContacts = providerConfig.AlertContacts
	}
	request.AssignedAlertContacts = parseAlertContacts(alertContacts)

	if providerConfig != nil && len(providerConfig.MaintenanceWindows) != 0 {
		request.MaintenanceWindowsIds = parseMaintenanceWindows(providerConfig.MaintenanceWindows)
	}

	if providerConfig != nil && len(providerConfig.CustomHTTPStatuses) != 0 {
		request.SuccessHttpResponseCodes = parseSuccessHTTPStatuses(providerConfig.CustomHTTPStatuses)
	}

	monitorType := "http"
	if providerConfig != nil && len(providerConfig.MonitorType) != 0 {
		monitorType = providerConfig.MonitorType
	}

	if strings.EqualFold(monitorType, "keyword") {
		request.Type = "KEYWORD"

		request.KeywordValue = providerConfig.KeywordValue
		if len(providerConfig.KeywordValue) == 0 {
			log.Error(nil, "Monitor is of type Keyword but the `keyword-value` is missing")
		}

		keywordExists := "yes" // By default, alert when the keyword exists
		if len(providerConfig.KeywordExists) != 0 {
			keywordExists = providerConfig.KeywordExists
		}
		if strings.EqualFold(keywordExists, "no") {
			request.KeywordType = "ALERT_NOT_EXISTS"
		} else {
			request.KeywordType = "ALERT_EXISTS"
		}
	} else {
		request.Type = "HTTP" // By default monitor is of type HTTP
	}

	return request
}

func (monitor *UpTimeMonitorService) Remove(m models.Monitor) {
	// Detach the monitor from any status pages it belongs to before deleting
	if pspIDs, err := monitor.statusPageService.GetStatusPagesForMonitor(m.ID); err == nil {
		for _, pspID := range pspIDs {
			if _, err := monitor.statusPageService.RemoveMonitorFromStatusPage(UpTimeStatusPage{ID: pspID}, m); err != nil {
				log.Info("Monitor couldn't be removed from status page " + pspID + ": " + err.Error())
			}
		}
	}

	response := doRequestWithRetries("DELETE", apiURL(monitor.url, "/monitors/"+m.ID), monitor.requestHeaders(), nil)

	if isSuccess(response) {
		log.Info("Monitor Removed: " + m.Name)
	} else {
		log.Info("RemoveMonitor Request failed. Status Code: " + strconv.Itoa(response.StatusCode) + ". Error: " + errorExcerpt(response.Bytes))
	}
}

func (monitor *UpTimeMonitorService) handleStatusPagesConfig(monitorToAdd models.Monitor, monitorId string) {
	// Retrieve provider configuration
	providerConfig, _ := monitorToAdd.Config.(*endpointmonitorv1alpha1.UptimeRobotConfig)

	if providerConfig != nil && len(providerConfig.StatusPages) != 0 {
		IDs := strings.Split(providerConfig.StatusPages, "-")
		for i := range IDs {
			monitor.updateStatusPages(IDs[i], models.Monitor{ID: monitorId})
		}
	}
}

func (monitor *UpTimeMonitorService) updateStatusPages(statusPages string, monitorToAdd models.Monitor) {
	statusPage := UpTimeStatusPage{ID: statusPages}
	_, err := monitor.statusPageService.AddMonitorToStatusPage(statusPage, monitorToAdd)
	if err != nil {
		log.Info("Monitor couldn't be added to status page: " + err.Error())
	}
}

// parseAlertContacts parses the v2-style "id_threshold_recurrence-id2_t2_r2"
// format. The CRD/config field format is unchanged for backwards compatibility.
func parseAlertContacts(alertContacts string) []UptimeMonitorAlertContact {
	if alertContacts == "" {
		return nil
	}
	var contacts []UptimeMonitorAlertContact
	for _, contact := range strings.Split(alertContacts, "-") {
		parts := strings.Split(contact, "_")
		id, err := strconv.Atoi(parts[0])
		if err != nil || id == 0 {
			continue
		}
		contacts = append(contacts, UptimeMonitorAlertContact{
			AlertContactId: id,
			Threshold:      atoiOrZero(at(parts, 1)),
			Recurrence:     atoiOrZero(at(parts, 2)),
		})
	}
	return contacts
}

func at(parts []string, i int) string {
	if i < len(parts) {
		return parts[i]
	}
	return ""
}

func atoiOrZero(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return value
}

// parseMaintenanceWindows parses a dash-separated id list e.g. "123-456"
func parseMaintenanceWindows(maintenanceWindows string) []int {
	if maintenanceWindows == "" {
		return nil
	}
	var ids []int
	for _, part := range strings.Split(maintenanceWindows, "-") {
		if id, err := strconv.Atoi(part); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// parseSuccessHTTPStatuses maps the v2 "200:0_401:1_503:1" format to the v3
// successHttpResponseCodes list: only codes flagged ":1" count as UP.
// ponytail: v3 can only express success codes, so ":0" codes are simply left
// out of the success list, which makes them count as DOWN.
func parseSuccessHTTPStatuses(customHTTPStatuses string) []string {
	var codes []string
	for _, token := range strings.FieldsFunc(customHTTPStatuses, func(r rune) bool { return r == '_' || r == ',' }) {
		code, flag, found := strings.Cut(token, ":")
		if !found || flag == "1" {
			codes = append(codes, code)
		}
	}
	return codes
}
