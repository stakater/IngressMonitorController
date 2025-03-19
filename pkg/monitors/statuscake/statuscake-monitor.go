package statuscake

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	statuscake "github.com/StatusCakeDev/statuscake-go"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/kube"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/secret"
)

var log = logf.Log.WithName("statuscake-monitor")
var backoffTime = 10 * time.Second // Default backoff time for errors

// StatusCakeMonitorService is the service structure for StatusCake
type StatusCakeMonitorService struct {
	apiKey       string
	url          string
	username     string
	cgroup       string
	client       *http.Client
	cacheLock    sync.Mutex
	monitorCache map[string]*models.Monitor
	allMonitors  []models.Monitor // Cache for GetAll results
	cacheTime    time.Time
	cacheTTL     time.Duration // Time-to-live for cache entries
}

// Equal compares two monitors to determine if they have the same configuration
// Returns true if monitors are equal, false if they need to be updated
func (monitor *StatusCakeMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// If old monitor has no config, we need to update
	if oldMonitor.Config == nil {
		return false
	}

	// Type-assert to get the actual config structs
	oldConfig, ok1 := oldMonitor.Config.(*endpointmonitorv1alpha1.StatusCakeConfig)
	newConfig, ok2 := newMonitor.Config.(*endpointmonitorv1alpha1.StatusCakeConfig)

	if !ok1 || !ok2 {
		// If type assertion fails, consider them different
		return false
	}

	// Track differences to log them all before returning
	hasDifferences := false

	// Compare boolean fields - always compare these regardless of zero value
	if oldConfig.Paused != newConfig.Paused {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "Paused",
			"old", oldConfig.Paused, "new", newConfig.Paused)
		hasDifferences = true
	}

	if oldConfig.FollowRedirect != newConfig.FollowRedirect {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "FollowRedirect",
			"old", oldConfig.FollowRedirect, "new", newConfig.FollowRedirect)
		hasDifferences = true
	}

	if oldConfig.EnableSSLAlert != newConfig.EnableSSLAlert {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "EnableSSLAlert",
			"old", oldConfig.EnableSSLAlert, "new", newConfig.EnableSSLAlert)
		hasDifferences = true
	}

	// For non-boolean fields, only compare if the new value is non-zero
	// CheckRate
	if newConfig.CheckRate != 0 && oldConfig.CheckRate != newConfig.CheckRate {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "CheckRate",
			"old", oldConfig.CheckRate, "new", newConfig.CheckRate)
		hasDifferences = true
	}

	// TestType
	if newConfig.TestType != "" && oldConfig.TestType != newConfig.TestType {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "TestType",
			"old", oldConfig.TestType, "new", newConfig.TestType)
		hasDifferences = true
	}

	// ContactGroup
	if newConfig.ContactGroup != "" && oldConfig.ContactGroup != newConfig.ContactGroup {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "ContactGroup",
			"old", oldConfig.ContactGroup, "new", newConfig.ContactGroup)
		hasDifferences = true
	}

	// TestTags
	if newConfig.TestTags != "" && oldConfig.TestTags != newConfig.TestTags {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "TestTags",
			"old", oldConfig.TestTags, "new", newConfig.TestTags)
		hasDifferences = true
	}

	// Port
	if newConfig.Port != 0 && oldConfig.Port != newConfig.Port {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "Port",
			"old", oldConfig.Port, "new", newConfig.Port)
		hasDifferences = true
	}

	// TriggerRate
	if newConfig.TriggerRate != 0 && oldConfig.TriggerRate != newConfig.TriggerRate {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "TriggerRate",
			"old", oldConfig.TriggerRate, "new", newConfig.TriggerRate)
		hasDifferences = true
	}

	// Confirmation
	if newConfig.Confirmation != 0 && oldConfig.Confirmation != newConfig.Confirmation {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "Confirmation",
			"old", oldConfig.Confirmation, "new", newConfig.Confirmation)
		hasDifferences = true
	}

	// FindString
	if newConfig.FindString != "" && oldConfig.FindString != newConfig.FindString {
		log.Info("Difference detected", "monitor", oldMonitor.Name, "field", "FindString",
			"old", oldConfig.FindString, "new", newConfig.FindString)
		hasDifferences = true
	}

	// Only return after checking all fields
	return !hasDifferences
}

// buildUpsertForm function is used to create the form needed to Add or update a monitor
func buildUpsertForm(m models.Monitor, cgroup string) url.Values {
	f := url.Values{}
	f.Add("name", m.Name)
	unEscapedURL, _ := url.QueryUnescape(m.URL)
	f.Add("website_url", unEscapedURL)

	// Retrieve provider configuration
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.StatusCakeConfig)

	if providerConfig != nil && providerConfig.CheckRate > 0 {
		f.Add("check_rate", strconv.Itoa(providerConfig.CheckRate))
	} else {
		f.Add("check_rate", "300")
	}

	if providerConfig != nil && len(providerConfig.TestType) > 0 {
		f.Add("test_type", providerConfig.TestType)
	} else {
		f.Add("test_type", "HTTP")
	}

	if providerConfig != nil && len(providerConfig.ContactGroup) > 0 {
		contactGroups := convertStringToArray(providerConfig.ContactGroup)
		for _, contactgroups := range contactGroups {
			f.Add("contact_groups[]", contactgroups)
		}
	} else {
		if cgroup != "" {
			contactGroups := convertStringToArray(cgroup)
			for _, contactgroups := range contactGroups {
				f.Add("contact_groups[]", contactgroups)
			}
		}
	}

	if providerConfig != nil && len(providerConfig.TestTags) > 0 {
		testTags := convertStringToArray(providerConfig.TestTags)
		for _, testTag := range testTags {
			f.Add("tags[]", testTag)
		}
	}

	if providerConfig != nil && len(providerConfig.Regions) > 0 {
		regions := convertStringToArray(providerConfig.Regions)
		for _, region := range regions {
			f.Add("regions[]", region)
		}
	}

	if providerConfig != nil && len(providerConfig.BasicAuthUser) > 0 {
		// This value is mandatory
		// Environment variable should define the password
		// Mounted via a secret; key is the username, value is the password
		basicPass := os.Getenv(providerConfig.BasicAuthUser)
		if basicPass != "" {
			f.Add("basic_username", providerConfig.BasicAuthUser)
			f.Add("basic_password", basicPass)
			log.Info("Basic auth requirement detected. Setting username and password")
		} else {
			log.Info("Error reading basic auth password from environment variable")
		}
	}

	if providerConfig != nil && len(providerConfig.BasicAuthSecret) > 0 {
		k8sClient, err := kube.GetClient()
		if err != nil {
			panic(err)
		}

		namespace := kube.GetCurrentKubernetesNamespace()
		username, password, err := secret.ReadBasicAuthSecret(k8sClient.CoreV1().Secrets(namespace), providerConfig.BasicAuthSecret)

		if err != nil {
			log.Error(err, "Could not read the secret")
		} else {
			f.Add("basic_username", username)
			f.Add("basic_password", password)
			log.Info("Basic auth requirement detected. Setting username and password")
		}
	}

	if providerConfig != nil && len(providerConfig.StatusCodes) > 0 {
		f.Add("status_codes_csv", providerConfig.StatusCodes)

	} else {
		statusCodes := []string{
			"204", // No content
			"205", // Reset content
			"206", // Partial content
			"303", // See other
			"305", // Use proxy
			// https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#4xx_Client_errors
			// https://support.cloudflare.com/hc/en-us/articles/115003014512/
			"400",
			"401",
			"402",
			"403",
			"404",
			"405",
			"406",
			"407",
			"408",
			"409",
			"410",
			"411",
			"412",
			"413",
			"414",
			"415",
			"416",
			"417",
			"418",
			"421",
			"422",
			"423",
			"424",
			"425",
			"426",
			"428",
			"429",
			"431",
			"444",
			"451",
			"499",
			// https://support.cloudflare.com/hc/en-us/articles/115003011431/
			"500",
			"501",
			"502",
			"503",
			"504",
			"505",
			"506",
			"507",
			"508",
			"509",
			"510",
			"511",
			"520",
			"521",
			"522",
			"523",
			"524",
			"525",
			"526",
			"527",
			"530",
			"598",
			"599",
		}
		f.Add("status_codes_csv", strings.Join(statusCodes, ","))
	}

	if providerConfig != nil {
		if providerConfig.Paused {
			f.Add("paused", "1")
		} else {
			f.Add("paused", "0")
		}
		if providerConfig.FollowRedirect {
			f.Add("follow_redirects", "1")
		} else {
			f.Add("follow_redirects", "0")
		}
		if providerConfig.EnableSSLAlert {
			f.Add("enable_ssl_alert", "1")
		} else {
			f.Add("enable_ssl_alert", "0")
		}
	}

	// Shifted to contact groups api
	// TODO: create proper structs to cater contact groups api
	/*
		if providerConfig != nil && len(providerConfig.PingURL) > 0 {
			f.Add("ping_url", providerConfig.PingURL)
		}
	*/
	if providerConfig != nil && providerConfig.TriggerRate > 0 {
		f.Add("trigger_rate", strconv.Itoa(providerConfig.TriggerRate))
	}
	if providerConfig != nil && providerConfig.Port > 0 {
		f.Add("port", strconv.Itoa(providerConfig.Port))
	}
	if providerConfig != nil && providerConfig.Confirmation > 0 {
		f.Add("confirmation", strconv.Itoa(providerConfig.Confirmation))
	}
	if providerConfig != nil && len(providerConfig.FindString) > 0 {
		f.Add("find_string", providerConfig.FindString)
	}
	if providerConfig != nil && len(providerConfig.RawPostData) > 0 {
		f.Add("post_raw", providerConfig.RawPostData)
	}
	if providerConfig != nil && len(providerConfig.UserAgent) > 0 {
		f.Add("user_agent", providerConfig.UserAgent)
	}
	return f
}

// convertValuesToString changes multiple values returned by same key to string for validation purposes
func convertUrlValuesToString(vals url.Values, key string) string {
	var valuesArray []string
	for k, v := range vals {
		if k == key {
			valuesArray = append(valuesArray, v...)
		}
	}
	return strings.Join(valuesArray, ",")
}

// convertStringToArray function is used to convert string to []string
func convertStringToArray(stringValues string) []string {
	stringArray := strings.Split(stringValues, ",")
	return stringArray
}

// Setup function is used to initialise the StatusCake service
func (service *StatusCakeMonitorService) Setup(p config.Provider) {
	service.apiKey = p.ApiKey
	service.url = p.ApiURL
	service.username = p.Username
	service.cgroup = p.AlertContacts
	service.client = &http.Client{}
	service.monitorCache = make(map[string]*models.Monitor)
	service.cacheTime = time.Now()

	// TO DO This actually should be a parameter in the statuscake provider config

	service.cacheTTL = 24 * time.Hour // Cache TTL set to 24 hours

	// Start a goroutine to clear the cache periodically
	go service.startCacheCleaner()
}

// Update startCacheCleaner method
func (service *StatusCakeMonitorService) startCacheCleaner() {
	ticker := time.NewTicker(service.cacheTTL)
	defer ticker.Stop()

	for {
		<-ticker.C
		// Clear all cache entries
		service.cacheLock.Lock()
		service.monitorCache = make(map[string]*models.Monitor)
		service.allMonitors = nil // Clear the GetAll cache as well
		service.cacheTime = time.Now()
		log.Info("Cache reset due to time expiration", "cacheTTL", service.cacheTTL)
		service.cacheLock.Unlock()
	}
}

// GetByName function will Get a monitor by it's name
func (service *StatusCakeMonitorService) GetByName(name string) (*models.Monitor, error) {
	// Always fetch fresh data for GetByName to ensure we have current state
	// This is especially important after deletion operations
	monitors := service.fetchAllMonitors()

	// Due to API eventual consistency, check multiple times with delays
	// if we're looking for a monitor that might have been deleted
	found := false
	var targetMonitor *models.Monitor

	for _, monitor := range monitors {
		if monitor.Name == name {
			found = true
			targetMonitor = &monitor
			break
		}
	}

	// Update cache with fresh data
	service.cacheLock.Lock()
	service.allMonitors = monitors
	service.cacheTime = time.Now()
	service.cacheLock.Unlock()

	if found {
		return targetMonitor, nil
	}

	errorString := "GetByName Request failed for name: " + name
	return nil, errors.New(errorString)
}

// GetByID function with improved error handling
func (service *StatusCakeMonitorService) GetByID(id string) (*models.Monitor, error) {
	// Check the cache first
	service.cacheLock.Lock()
	if cached, found := service.monitorCache[id]; found {
		service.cacheLock.Unlock()
		return cached, nil
	}
	service.cacheLock.Unlock()

	// Cache miss; fetch from StatusCake
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err, "Unable to Parse monitor URL")
		return nil, err
	}
	u.Path = fmt.Sprintf("/v1/uptime/%s", id)
	u.Scheme = "https"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Error(err, "Unable to create request")
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))

	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to make HTTP call")
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "Unable to read response body")
		return nil, err
	}

	// More detailed logging for debugging
	logDetails := map[string]interface{}{
		"monitorID":   id,
		"statusCode":  resp.StatusCode,
		"statusText":  http.StatusText(resp.StatusCode),
		"contentType": resp.Header.Get("Content-Type"),
		"requestID":   resp.Header.Get("X-Request-ID"),
	}

	// Only include body in logs if there's an issue
	if resp.StatusCode != http.StatusOK {
		logDetails["responseBody"] = string(bodyBytes)
		log.Info("GetByID request failed", "details", logDetails)
		return nil, errors.New("GetByID Request failed")
	}

	// Only try to unmarshal if we have data
	if len(bodyBytes) > 0 {
		var StatusCakeMonitorData statuscake.UptimeTestResponse
		err = json.Unmarshal(bodyBytes, &StatusCakeMonitorData)
		if err != nil {
			log.Error(err, "Unable to unmarshal response", "body", string(bodyBytes))
			return nil, err
		}

		monitor := StatusCakeApiResponseDataToBaseMonitorMapper(StatusCakeMonitorData)

		// Store in cache
		service.cacheLock.Lock()
		service.monitorCache[id] = monitor
		service.cacheLock.Unlock()

		return monitor, nil
	}

	return nil, errors.New("empty response from statusCake")
}

// doRequest function with smarter rate limiting for multiple controllers
func (service *StatusCakeMonitorService) doRequest(req *http.Request) (*http.Response, error) {
	// Use a context with timeout to prevent hanging connections
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Add jitter to prevent controllers from syncing up
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(jitter)

	// Set content-type header for POST and PUT requests if not already set
	if (req.Method == "POST" || req.Method == "PUT") && req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// Add a unique identifier to help with debugging across instances
	instanceID := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	req.Header.Add("X-Instance-ID", instanceID)

	resp, err := service.client.Do(req)
	if err != nil {
		log.Error(err, "HTTP request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"instanceID", instanceID)
		return nil, err
	}

	// Handle 429 Too Many Requests with a proper retry based on headers
	if resp.StatusCode == http.StatusTooManyRequests {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		resetTime := 5 * time.Second
		// Parse x-ratelimit-reset header
		if reset := resp.Header.Get("x-ratelimit-reset"); reset != "" {
			if seconds, err := strconv.Atoi(reset); err == nil && seconds > 0 {
				resetTime = time.Duration(seconds+1) * time.Second
			}
		}

		log.Info("Rate limit exceeded, waiting to retry",
			"method", req.Method,
			"url", req.URL.String(),
			"resetSeconds", resetTime.Seconds(),
			"instanceID", instanceID,
			"body", string(bodyBytes), // Use bodyBytes so it isn't "declared and not used"
		)

		// Sleep for the reset duration + jitter
		time.Sleep(resetTime + jitter)

		newReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
		if err != nil {
			return nil, err
		}
		for name, values := range req.Header {
			for _, value := range values {
				newReq.Header.Add(name, value)
			}
		}
		return service.doRequest(newReq)
	}

	// Check for any other problematic responses
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent {

		// Get response body for logging purposes
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close() // Close this response

		sleepTime := backoffTime

		log.Info("API error encountered, backing off",
			"status", resp.StatusCode,
			"method", req.Method,
			"url", req.URL.String(),
			"body", string(bodyBytes),
			"seconds", sleepTime.Seconds(),
			"instanceID", instanceID)

		// Sleep and retry with exponential backoff
		sleepTime = sleepTime + jitter
		time.Sleep(sleepTime)

		// Create a new request since we've consumed the body
		newReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
		if err != nil {
			return nil, err
		}

		// Copy headers
		for name, values := range req.Header {
			for _, value := range values {
				newReq.Header.Add(name, value)
			}
		}

		return service.doRequest(newReq) // Retry after backoff
	}

	// Check remaining rate limit and slow down if needed
	if remaining := resp.Header.Get("x-ratelimit-remaining"); remaining != "" {
		if rem, err := strconv.Atoi(remaining); err == nil && rem <= 1 {
			if reset := resp.Header.Get("x-ratelimit-reset"); reset != "" {
				if seconds, err := strconv.Atoi(reset); err == nil && seconds > 0 {
					log.V(1).Info("Rate limit nearly exhausted, slowing down",
						"method", req.Method,
						"remaining", remaining,
						"resetSeconds", seconds)

					time.Sleep(500 * time.Millisecond)
				}
			}
		}
	}

	// Even with "successful" responses, check if body is empty and retry if so
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		log.Error(err, "Unable to read response body")
		return nil, err
	}

	// Create a new response with the same data but with a ReadCloser body
	newResp := *resp
	newResp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Empty response with 200 status might indicate rate limiting
	if len(bodyBytes) == 0 && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		log.V(1).Info("Empty body received with success status code - possible rate limiting",
			"method", req.Method,
			"url", req.URL.String(),
			"status", resp.StatusCode,
			"instanceID", instanceID)

		// Wait before retry
		time.Sleep(3*time.Second + jitter)

		// Create new request for retry
		newReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
		if err != nil {
			return &newResp, nil // Return the empty response if new request can't be created
		}

		// Copy headers
		for name, values := range req.Header {
			for _, value := range values {
				newReq.Header.Add(name, value)
			}
		}

		return service.doRequest(newReq) // Retry after backoff
	}

	return &newResp, nil
}

// GetAll function will fetch all monitors
func (service *StatusCakeMonitorService) GetAll() []models.Monitor {
	// Check if we have a cached version
	service.cacheLock.Lock()

	// If the cache is still valid and we have data, return it
	if !service.cacheTime.IsZero() && time.Since(service.cacheTime) < service.cacheTTL && len(service.allMonitors) > 0 {
		monitors := service.allMonitors
		service.cacheLock.Unlock()
		log.V(1).Info("Returning cached monitors list", "count", len(monitors))
		return monitors
	}
	service.cacheLock.Unlock()

	// Otherwise fetch fresh data
	monitors := service.fetchAllMonitors()

	// Update the cache
	service.cacheLock.Lock()
	service.allMonitors = monitors
	service.cacheLock.Unlock()

	return monitors
}

// Add a new method to fetch all monitors from the API
func (service *StatusCakeMonitorService) fetchAllMonitors() []models.Monitor {
	var monitors []models.Monitor
	page := 1
	for {
		res := service.fetchMonitors(page)
		if res == nil {
			break
		}
		for _, data := range res.StatusCakeData {
			monitor, err := service.GetByID(data.TestID)
			if err == nil {
				monitors = append(monitors, *monitor)
			} else {
				monitors = append(monitors, *StatusCakeMonitorMonitorToBaseMonitorMapper(data))
			}
		}
		if page >= res.StatusCakeMetadata.PageCount {
			break
		}
		page++

		// Add a short delay to avoid hammering the API
		time.Sleep(500 * time.Millisecond)
	}
	return monitors
}

// fetchMonitors with improved error handling
func (service *StatusCakeMonitorService) fetchMonitors(page int) *StatusCakeMonitor {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err, "Unable to Parse monitor URL")
		return nil
	}
	u.Path = "/v1/uptime/"
	query := u.Query()
	query.Add("limit", "100")
	query.Add("page", strconv.Itoa(page))
	u.RawQuery = query.Encode()
	u.Scheme = "https"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Error(err, "Unable to create request")
		return nil
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))

	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to make HTTP call")
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close() // Always close the body

	if err != nil {
		log.Error(err, "Unable to read response body")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Error(nil, "fetchMonitors request failed",
			"statusCode", resp.StatusCode,
			"url", req.URL.String(),
			"body", string(bodyBytes))
		return nil
	}

	// Only try to parse if we have content
	if len(bodyBytes) == 0 {
		log.Error(nil, "Empty response body")
		return nil
	}

	var StatusCakeMonitor StatusCakeMonitor
	err = json.Unmarshal(bodyBytes, &StatusCakeMonitor)
	if err != nil {
		log.Error(err, "Failed to unmarshal response", "body", string(bodyBytes))
		return nil
	}

	return &StatusCakeMonitor
}

// Add will create a new Monitor
func (service *StatusCakeMonitorService) Add(m models.Monitor) {
	// First check if a monitor with this name already exists
	existingMonitor, err := service.GetByName(m.Name)
	if err == nil && existingMonitor != nil {
		// Monitor already exists, update ID in our model and return
		log.Info("Monitor already exists, skipping creation", "name", m.Name, "id", existingMonitor.ID)
		m.ID = existingMonitor.ID
		return
	}

	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err, "Unable to Parse monitor URL")
		return
	}
	u.Path = "/v1/uptime"
	u.Scheme = "https"
	data := buildUpsertForm(m, service.cgroup)
	req, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Error(err, "Unable to create http request")
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to make HTTP call")
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "Unable to read response body")
		return
	}

	// For StatusCake API, 201 Created is the success code for monitor creation
	if resp.StatusCode == http.StatusCreated {
		// Try to parse the response, but don't fail if we can't
		var createResp statuscake.UptimeTestResponse
		if len(bodyBytes) > 0 {
			err = json.Unmarshal(bodyBytes, &createResp)
			if err == nil && createResp.Data.ID != "" {
				// Update the monitor ID from the response
				m.ID = createResp.Data.ID
			}
		}

		// Add to cache
		service.cacheLock.Lock()
		service.monitorCache[m.ID] = &m
		service.allMonitors = nil // Invalidate GetAll cache
		service.cacheLock.Unlock()

		log.Info("Monitor Added: " + m.Name)
	} else {
		// Log the full error details for debugging
		log.Error(nil, "Insert Request failed",
			"name", m.Name,
			"statusCode", resp.StatusCode,
			"body", string(bodyBytes),
			"url", req.URL.String())
	}
}

// Update will update an existing Monitor
func (service *StatusCakeMonitorService) Update(m models.Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err, "Unable to Parse monitor URL")
		return
	}
	u.Path = fmt.Sprintf("/v1/uptime/%s", m.ID)
	u.Scheme = "https"
	data := buildUpsertForm(m, service.cgroup)
	req, err := http.NewRequest("PUT", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Error(err, "Unable to create http request")
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))
	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to make HTTP call")
		return
	}
	if resp.StatusCode == http.StatusNoContent {
		// Remove stale monitor entry from cache so future calls re-fetch
		service.cacheLock.Lock()
		delete(service.monitorCache, m.ID)
		service.cacheLock.Unlock()

		log.Info("Monitor Updated: " + m.ID + m.Name)
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "Unable to read response")
			os.Exit(1)
		}
		log.Error(nil, "Update Request failed for name: "+m.Name+" with status code "+strconv.Itoa(resp.StatusCode))
		log.Error(nil, string(bodyBytes))
	}

	// Invalidate the GetAll cache
	service.cacheLock.Lock()
	service.allMonitors = nil
	service.cacheLock.Unlock()
}

// Remove will delete an existing Monitor
func (service *StatusCakeMonitorService) Remove(m models.Monitor) {
	// Immediately clear cache entries on deletion request
	service.cacheLock.Lock()
	delete(service.monitorCache, m.ID)
	service.allMonitors = nil // Force refresh of GetAll cache
	service.cacheLock.Unlock()

	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err, "Unable to Parse monitor URL")
		return
	}
	u.Path = fmt.Sprintf("/v1/uptime/%s", m.ID)
	u.Scheme = "https"

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		log.Error(err, "Unable to create http request")
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))
	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to make HTTP call")
		return
	}

	if resp.StatusCode == http.StatusNoContent {
		// Add longer delay to handle eventual consistency in StatusCake API
		time.Sleep(5 * time.Second)

		// Explicitly invalidate cache again after the delay
		service.cacheLock.Lock()
		delete(service.monitorCache, m.ID)
		service.allMonitors = nil
		service.cacheTime = time.Time{} // Reset cache time to force refresh
		service.cacheLock.Unlock()

		// Log deletion with ID first for better log parsing
		log.Info("Monitor Deleted: " + m.ID + m.Name)
	} else {
		log.Error(nil, fmt.Sprintf("Delete Request failed for Monitor: %s with id: %s", m.Name, m.ID))
	}

	// Always reset caches regardless of deletion success to ensure tests pass
	service.cacheLock.Lock()
	delete(service.monitorCache, m.ID)
	service.allMonitors = nil
	service.cacheTime = time.Time{} // Reset cache time to force refresh
	service.cacheLock.Unlock()
}
