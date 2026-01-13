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
	apiKey          string
	url             string
	username        string
	cgroup          string
	client          *http.Client
	cacheLock       sync.Mutex
	monitorCache    map[string]*models.Monitor
	allMonitors     []models.Monitor // Cache for GetAll results
	cacheTime       time.Time
	cacheTTL        time.Duration // Time-to-live for cache entries
	stopCleanerChan chan struct{}
}

// Equal compares two monitors to determine if they have the same configuration
// Returns true if monitors are equal, false if they need to be updated
func (monitor *StatusCakeMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// Compare base monitor fields first
	if oldMonitor.URL != newMonitor.URL {
		log.Info("Monitor URL has changed",
			"name", newMonitor.Name,
			"oldURL", oldMonitor.URL,
			"newURL", newMonitor.URL)
		return false
	}

	// If old monitor has no config, we need to update
	if oldMonitor.Config == nil {
		log.Info("Old monitor has no config, considering different", "name", newMonitor.Name)
		return false
	}

	// Type-assert to get the actual config structs
	oldConfig, ok1 := oldMonitor.Config.(*endpointmonitorv1alpha1.StatusCakeConfig)
	newConfig, ok2 := newMonitor.Config.(*endpointmonitorv1alpha1.StatusCakeConfig)

	if !ok1 || !ok2 {
		// If type assertion fails, consider them different
		log.Info("Type assertion failed for configs",
			"name", newMonitor.Name,
			"ok1", ok1,
			"ok2", ok2)
		return false
	}

	// Log the full configuration objects only at high verbosity level
	if log.V(4).Enabled() {
		oldConfigJSON, _ := json.Marshal(oldConfig)
		newConfigJSON, _ := json.Marshal(newConfig)
		log.V(4).Info("Monitor configurations",
			"name", newMonitor.Name,
			"oldConfig", string(oldConfigJSON),
			"newConfig", string(newConfigJSON))
	}

	// Track differences to log them all before returning
	hasDifferences := false
	differences := []string{}

	// Check each field and record differences
	if oldConfig.Paused != newConfig.Paused {
		differences = append(differences, fmt.Sprintf("Paused: %v -> %v", oldConfig.Paused, newConfig.Paused))
		hasDifferences = true
	}

	if oldConfig.FollowRedirect != newConfig.FollowRedirect {
		differences = append(differences, fmt.Sprintf("FollowRedirect: %v -> %v", oldConfig.FollowRedirect, newConfig.FollowRedirect))
		hasDifferences = true
	}

	if oldConfig.EnableSSLAlert != newConfig.EnableSSLAlert {
		differences = append(differences, fmt.Sprintf("EnableSSLAlert: %v -> %v", oldConfig.EnableSSLAlert, newConfig.EnableSSLAlert))
		hasDifferences = true
	}

	if oldConfig.CheckRate != newConfig.CheckRate {
		differences = append(differences, fmt.Sprintf("CheckRate: %v -> %v", oldConfig.CheckRate, newConfig.CheckRate))
		hasDifferences = true
	}

	if oldConfig.TestType != newConfig.TestType {
		differences = append(differences, fmt.Sprintf("TestType: %v -> %v", oldConfig.TestType, newConfig.TestType))
		hasDifferences = true
	}

	if oldConfig.ContactGroup != newConfig.ContactGroup {
		differences = append(differences, fmt.Sprintf("ContactGroup: %v -> %v", oldConfig.ContactGroup, newConfig.ContactGroup))
		hasDifferences = true
	}

	if oldConfig.TestTags != newConfig.TestTags {
		differences = append(differences, fmt.Sprintf("TestTags: %v -> %v", oldConfig.TestTags, newConfig.TestTags))
		hasDifferences = true
	}

	if oldConfig.Port != newConfig.Port {
		differences = append(differences, fmt.Sprintf("Port: %v -> %v", oldConfig.Port, newConfig.Port))
		hasDifferences = true
	}

	if oldConfig.TriggerRate != newConfig.TriggerRate {
		differences = append(differences, fmt.Sprintf("TriggerRate: %v -> %v", oldConfig.TriggerRate, newConfig.TriggerRate))
		hasDifferences = true
	}

	if oldConfig.Confirmation != newConfig.Confirmation {
		differences = append(differences, fmt.Sprintf("Confirmation: %v -> %v", oldConfig.Confirmation, newConfig.Confirmation))
		hasDifferences = true
	}

	if oldConfig.FindString != newConfig.FindString {
		differences = append(differences, fmt.Sprintf("FindString: %v -> %v", oldConfig.FindString, newConfig.FindString))
		hasDifferences = true
	}

	if oldConfig.StatusCodes != newConfig.StatusCodes {
		differences = append(differences, fmt.Sprintf("StatusCodes: %v -> %v", oldConfig.StatusCodes, newConfig.StatusCodes))
		hasDifferences = true
	}

	if oldConfig.Regions != newConfig.Regions {
		differences = append(differences, fmt.Sprintf("Regions: %v -> %v", oldConfig.Regions, newConfig.Regions))
		hasDifferences = true
	}

	if oldConfig.UserAgent != newConfig.UserAgent {
		differences = append(differences, fmt.Sprintf("UserAgent: %v -> %v", oldConfig.UserAgent, newConfig.UserAgent))
		hasDifferences = true
	}

	if oldConfig.RawPostData != newConfig.RawPostData {
		differences = append(differences, fmt.Sprintf("RawPostData: %v -> %v", oldConfig.RawPostData, newConfig.RawPostData))
		hasDifferences = true
	}

	// Add missing fields that might be relevant to your comparison
	if oldConfig.BasicAuthUser != newConfig.BasicAuthUser {
		differences = append(differences, fmt.Sprintf("BasicAuthUser: %v -> %v", oldConfig.BasicAuthUser, newConfig.BasicAuthUser))
		hasDifferences = true
	}

	if oldConfig.BasicAuthSecret != newConfig.BasicAuthSecret {
		differences = append(differences, fmt.Sprintf("BasicAuthSecret: %v -> %v", oldConfig.BasicAuthSecret, newConfig.BasicAuthSecret))
		hasDifferences = true
	}

	// Log differences at a reasonable log level
	if hasDifferences && len(differences) > 0 {
		log.Info("Monitor needs update",
			"name", newMonitor.Name,
			"differences", strings.Join(differences, "; "))
	} else {
		log.V(2).Info("Monitor is up to date", "name", newMonitor.Name)
	}

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
	if providerConfig != nil && providerConfig.Timeout > 0 {
		f.Add("timeout", strconv.Itoa(providerConfig.Timeout))
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
	service.cacheTime = time.Time{} // Start with an empty time to force first load

	// Use a much longer TTL - 24 hours - to avoid unnecessary API calls
	service.cacheTTL = 24 * time.Hour

	service.stopCleanerChan = make(chan struct{})

	// Start a goroutine to clear the cache periodically
	go service.startCacheCleaner()
}

// Graceful shutdown
func (service *StatusCakeMonitorService) StopCacheCleaner() {
	close(service.stopCleanerChan)
}

// Update startCacheCleaner method
func (service *StatusCakeMonitorService) startCacheCleaner() {
	ticker := time.NewTicker(service.cacheTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Clear all cache entries
			service.cacheLock.Lock()
			service.monitorCache = make(map[string]*models.Monitor)
			service.allMonitors = nil // Clear the GetAll cache as well
			service.cacheTime = time.Now()
			log.V(1).Info("Cache reset due to time expiration", "cacheTTL", service.cacheTTL)
			service.cacheLock.Unlock()
		case <-service.stopCleanerChan:
			log.V(1).Info("Cache cleaner stopped")
			return
		}
	}
}

// GetByName function will Get a monitor by it's name
func (service *StatusCakeMonitorService) GetByName(name string) (*models.Monitor, error) {
	// First try the direct cache lookup by name
	service.cacheLock.Lock()
	if cached, found := service.monitorCache[name]; found {
		service.cacheLock.Unlock()
		return cached, nil
	}
	service.cacheLock.Unlock()

	// Then fall back to scanning through all monitors
	monitors := service.GetAll()
	for i := range monitors {
		if monitors[i].Name == name {
			// Important - store this result in the cache by name for future lookups
			monitor := monitors[i]

			// Store by name for future queries
			service.cacheLock.Lock()
			service.monitorCache[name] = &monitor
			service.cacheLock.Unlock()

			return &monitor, nil
		}
	}

	return nil, fmt.Errorf("monitor not found: %s", name)
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
		resp.Body.Close()

		resetTime := 5 * time.Second
		// Parse x-ratelimit-reset header
		if reset := resp.Header.Get("x-ratelimit-reset"); reset != "" {
			if seconds, err := strconv.Atoi(reset); err == nil && seconds > 0 {
				resetTime = time.Duration(seconds+1) * time.Second
			}
		}

		// Only log rate limits at higher verbosity level
		log.V(1).Info("Rate limit exceeded, waiting to retry",
			"method", req.Method,
			"path", req.URL.Path,
			"resetSeconds", resetTime.Seconds())

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

// GetAll function will fetch all monitors with intelligent cache rebuild
func (service *StatusCakeMonitorService) GetAll() []models.Monitor {
	service.cacheLock.Lock()

	// If cache is valid and we have data, return it
	if !service.cacheTime.IsZero() && time.Since(service.cacheTime) < service.cacheTTL && len(service.allMonitors) > 0 {
		monitors := service.allMonitors
		service.cacheLock.Unlock()
		log.V(1).Info("Returning cached monitors list", "count", len(monitors))
		return monitors
	}

	// If allMonitors was cleared but monitorCache has data, rebuild from monitorCache
	if len(service.monitorCache) > 0 &&
		len(service.allMonitors) == 0 {

		log.V(1).Info("Rebuilding monitor list from cache", "cacheSize", len(service.monitorCache))
		monitors := make([]models.Monitor, 0, len(service.monitorCache))
		for _, m := range service.monitorCache {
			monitors = append(monitors, *m)
		}

		service.allMonitors = monitors
		service.cacheTime = time.Now()
		service.cacheLock.Unlock()

		return monitors
	}
	service.cacheLock.Unlock()

	// Both caches invalid or empty, trigger batch load
	return service.fetchAllMonitors()
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

// Replace batchLoadAllMonitors with fetchAllMonitors (keeping the efficient implementation)
func (service *StatusCakeMonitorService) fetchAllMonitors() []models.Monitor {
	log.Info("Starting complete load of all monitors with full details")
	startTime := time.Now()

	// Step 1: Fetch basic monitor info to get all IDs
	var monitorIDs []string
	var basicMonitors = make(map[string]models.Monitor)
	page := 1

	for {
		res := service.fetchMonitors(page)
		if res == nil {
			break
		}

		for _, data := range res.StatusCakeData {
			monitorIDs = append(monitorIDs, data.TestID)
			basicMonitor := *StatusCakeMonitorMonitorToBaseMonitorMapper(data)
			basicMonitors[data.TestID] = basicMonitor
		}

		if page >= res.StatusCakeMetadata.PageCount {
			break
		}
		page++
		time.Sleep(300 * time.Millisecond) // Reduced delay between pagination
	}

	log.Info("Found monitors to fetch", "count", len(monitorIDs))

	// Step 2: Prepare for efficient detailed data loading
	var completeMonitors []models.Monitor
	var mutex sync.Mutex

	// Create a semaphore to limit concurrent API requests
	semaphore := make(chan struct{}, 5) // Max 5 concurrent requests
	var wg sync.WaitGroup

	// Step 3: Fetch complete data for all monitors concurrently but controlled

	for _, id := range monitorIDs {
		wg.Add(1)

		go func(monitorID string) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Fetch detailed monitor data
			u, err := url.Parse(service.url)
			if err != nil {
				log.Error(err, "Unable to parse URL", "monitorID", monitorID)
				return
			}

			u.Path = fmt.Sprintf("/v1/uptime/%s", monitorID)
			u.Scheme = "https"

			req, err := http.NewRequest("GET", u.String(), nil)
			if err != nil {
				log.Error(err, "Unable to create request", "monitorID", monitorID)
				return
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))

			// Use doRequest instead of direct client.Do to properly handle rate limits
			resp, err := service.doRequest(req)
			if err != nil {
				log.Error(err, "HTTP request failed", "monitorID", monitorID)
				return
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil || len(bodyBytes) == 0 {
				return
			}

			var data statuscake.UptimeTestResponse
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				log.Error(err, "Failed to unmarshal response", "monitorID", monitorID)
				return
			}

			detailedMonitor := StatusCakeApiResponseDataToBaseMonitorMapper(data)

			// THIS IS THE KEY PART THAT'S MISSING:
			// Add the detailed monitor to our results collection
			mutex.Lock()
			completeMonitors = append(completeMonitors, *detailedMonitor)
			service.monitorCache[monitorID] = detailedMonitor
			mutex.Unlock()
		}(id)

		// Larger delay between starting goroutines when multiple instances share API key
		time.Sleep(100 * time.Millisecond) // Increased from 50ms
	}

	// Wait for all fetches to complete
	wg.Wait()
	close(semaphore)

	// Step 4: Fill in any missing monitors with basic data as fallback
	for id, basicMonitor := range basicMonitors {
		found := false
		for _, completeMonitor := range completeMonitors {
			if completeMonitor.ID == id {
				found = true
				break
			}
		}

		if !found {
			log.Info("Using basic data as fallback for monitor", "id", id, "name", basicMonitor.Name)
			completeMonitors = append(completeMonitors, basicMonitor)

			// Also cache it
			mutex.Lock()
			service.monitorCache[id] = &basicMonitor
			mutex.Unlock()
		}
	}

	// Update cache with complete data
	service.cacheLock.Lock()
	service.allMonitors = completeMonitors
	service.cacheTime = time.Now()
	service.cacheLock.Unlock()

	log.Info("Completed loading all monitors with full details",
		"count", len(completeMonitors),
		"totalTime", time.Since(startTime).String())

	return completeMonitors
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
		if len(bodyBytes) > 0 {
			// Define a custom struct to match the actual response format
			var createResp struct {
				Data struct {
					NewID string `json:"new_id"`
				} `json:"data"`
			}

			err = json.Unmarshal(bodyBytes, &createResp)
			if err == nil && createResp.Data.NewID != "" {
				// Update the monitor ID from the response
				m.ID = createResp.Data.NewID
				log.Info("Monitor ID captured from API response", "id", m.ID, "name", m.Name)
			} else {
				// Try with generic map as a fallback
				var result map[string]interface{}
				err = json.Unmarshal(bodyBytes, &result)
				if err == nil {
					if data, ok := result["data"].(map[string]interface{}); ok {
						if newID, ok := data["new_id"].(string); ok {
							m.ID = newID
							log.Info("Monitor ID captured using fallback", "id", m.ID)
						}
					}
				}

				if m.ID == "" {
					log.Error(err, "Failed to capture monitor ID", "name", m.Name, "body", string(bodyBytes))
				}
			}
		}

		// Add to cache
		service.cacheLock.Lock()
		if m.ID != "" {
			service.monitorCache[m.ID] = &m
		}
		service.monitorCache[m.Name] = &m // Index by name too
		// Don't clear allMonitors, just append the new one
		if service.allMonitors != nil {
			service.allMonitors = append(service.allMonitors, m)
		}
		service.cacheLock.Unlock()

		log.Info("Monitor Added: " + m.Name + " with ID: " + m.ID)
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
		// Update cache with new data
		service.cacheLock.Lock()

		// Remove any old cache key for the old name if changed
		// so we don't leave stale entries lying around
		for k, v := range service.monitorCache {
			if v.ID == m.ID && k != m.ID && k != m.Name {
				delete(service.monitorCache, k)
			}
		}

		// Refresh the entries
		service.monitorCache[m.ID] = &m
		service.monitorCache[m.Name] = &m

		// Update allMonitors in place
		for i, mon := range service.allMonitors {
			if mon.ID == m.ID {
				service.allMonitors[i] = m
				break
			}
		}
		service.cacheLock.Unlock()

		log.Info("Monitor Updated: " + m.ID + " " + m.Name)
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "Unable to read response")
			os.Exit(1)
		}
		log.Error(nil, "Update Request failed for name: "+m.Name+" with status code "+strconv.Itoa(resp.StatusCode))
		log.Error(nil, string(bodyBytes))
	}
}

// Remove will delete an existing Monitor
func (service *StatusCakeMonitorService) Remove(m models.Monitor) {
	// 1) Determine monitor ID if not set
	if m.ID == "" {
		found, err := service.GetByName(m.Name)
		if err != nil || found == nil {
			log.Error(err, "Unable to find monitor ID for removal", "name", m.Name)
			return // Exit early if we can't find the ID
		}
		m.ID = found.ID
	}

	// Ensure we have an ID before trying to delete
	if m.ID == "" {
		log.Error(nil, "Cannot remove monitor without ID", "name", m.Name)
		return
	}

	// 2) Make DELETE call to StatusCake with proper ID
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err, "Failed to parse StatusCake URL")
		return
	}

	// Make sure ID is in the path
	u.Path = fmt.Sprintf("/v1/uptime/%s", m.ID)
	u.Scheme = "https"

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		log.Error(err, "Failed to create DELETE request")
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))

	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "HTTP call failed", "name", m.Name, "id", m.ID)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Error(nil,
			"Failed to remove monitor remotely",
			"statusCode", resp.StatusCode,
			"name", m.Name,
			"id", m.ID,
			"body", string(bodyBytes),
		)
	}

	// 3) Remove from cache regardless of API response
	service.cacheLock.Lock()
	defer service.cacheLock.Unlock()

	delete(service.monitorCache, m.ID)
	delete(service.monitorCache, m.Name)

	// Remove from allMonitors list too
	for i := 0; i < len(service.allMonitors); i++ {
		if service.allMonitors[i].ID == m.ID || service.allMonitors[i].Name == m.Name {
			service.allMonitors = append(service.allMonitors[:i], service.allMonitors[i+1:]...)
			i-- // Adjust index after removal
		}
	}

	log.V(1).Info("Monitor removed from cache", "name", m.Name, "id", m.ID)
}
