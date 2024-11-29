package statuscake

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	statuscake "github.com/StatusCakeDev/statuscake-go"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/kube"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/secret"
)

var log = logf.Log.WithName("statuscake-monitor")
var rateLimiter = rate.NewLimiter(5, 1) // Allow 5 requests per second

// StatusCakeMonitorService is the service structure for StatusCake
type StatusCakeMonitorService struct {
	apiKey   string
	url      string
	username string
	cgroup   string
	client   *http.Client
}

func (monitor *StatusCakeMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	// TODO: Retrieve oldMonitor config and compare it here
	return false
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
		}
		if providerConfig.FollowRedirect {
			f.Add("follow_redirects", "1")
		}
		if providerConfig.EnableSSLAlert {
			f.Add("enable_ssl_alert", "1")
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
	if providerConfig != nil {
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
}

// GetByName function will Get a monitor by it's name
func (service *StatusCakeMonitorService) GetByName(name string) (*models.Monitor, error) {
	monitors := service.GetAll()
	if len(monitors) != 0 {
		for _, monitor := range monitors {
			if monitor.Name == name {
				return &monitor, nil
			}
		}
	}
	errorString := "GetByName Request failed for name: " + name
	return nil, errors.New(errorString)

}

// GetByID function will Get a monitor by it's ID
func (service *StatusCakeMonitorService) GetByID(id string) (*models.Monitor, error) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err, "Unable to Parse monitor URL")
		return nil, err
	}
	u.Path = fmt.Sprintf("/v1/uptime/%s", id)
	u.Scheme = "https"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Error(err, "Unable to retrieve monitor")
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))

	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to retrieve monitor")
		return nil, err
	}

	BodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "Unable to read response body")
	}
	bodyString := string(BodyBytes)

	if resp.StatusCode == http.StatusOK {

		// TODO use statuscake managed structs, rather than managing own structs

		var StatusCakeMonitorData statuscake.UptimeTestResponse
		err = json.Unmarshal(BodyBytes, &StatusCakeMonitorData)
		if err != nil {
			log.Error(err, "Unable to unmarshal response")
			return nil, err
		}
		return StatusCakeApiResponseDataToBaseMonitorMapper(StatusCakeMonitorData), nil
	}
	log.Info(fmt.Sprintf("Request failed with response: %s for id: %s", bodyString, id))

	return nil, errors.New("GetByID Request failed")
}

// doRequest function to handle requests to StatusCake and handle ratelimits.
func (service *StatusCakeMonitorService) doRequest(req *http.Request) (*http.Response, error) {
	// Wait for the rate limiter to allow a request
	err := rateLimiter.Wait(req.Context())
	if err != nil {
		log.Error(err, "Rate limiter wait failed")
		return nil, err
	}

	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "HTTP request failed")
		return nil, err
	}

	// Handle rate-limiting responses (HTTP 429)
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, err := strconv.Atoi(retryAfter)
			if err == nil {
				time.Sleep(time.Duration(seconds) * time.Second)
				return service.doRequest(req) // Retry after the specified delay
			}
		}
	}

	return resp, nil
}

// GetAll function will fetch all monitors
func (service *StatusCakeMonitorService) GetAll() []models.Monitor {
	var StatusCakeMonitorData []StatusCakeMonitorData
	page := 1
	for {
		res := service.fetchMonitors(page)
		if res != nil {
			StatusCakeMonitorData = append(StatusCakeMonitorData, res.StatusCakeData...)
			if page >= res.StatusCakeMetadata.PageCount {
				break
			}
		} else {
			return nil
		}
		page += 1
	}
	return StatusCakeMonitorMonitorsToBaseMonitorsMapper(StatusCakeMonitorData)
}

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
		log.Error(err, "Unable to retrieve monitor")
		return nil
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", service.apiKey))

	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to retrieve monitor")
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "Unable to read response body")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var StatusCakeMonitor StatusCakeMonitor
	err = json.Unmarshal(bodyBytes, &StatusCakeMonitor)
	if err != nil {
		log.Error(err, "Failed to unmarshal response")
		return nil
	}

	return &StatusCakeMonitor
}

// Add will create a new Monitor
func (service *StatusCakeMonitorService) Add(m models.Monitor) {
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
	resp, err := service.doRequest(req)
	if err != nil {
		log.Error(err, "Unable to make HTTP call")
		return
	}
	if resp.StatusCode == http.StatusCreated {
		log.Info("Monitor Added: " + m.Name)
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "Unable to read response")
			os.Exit(1)
		}
		log.Error(nil, "Insert Request failed for name: "+m.Name+" with status code "+strconv.Itoa(resp.StatusCode))
		log.Error(nil, string(bodyBytes))
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
}

// Remove will delete an existing Monitor
func (service *StatusCakeMonitorService) Remove(m models.Monitor) {
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
	if resp.StatusCode != http.StatusNoContent {
		log.Error(nil, fmt.Sprintf("Delete Request failed for Monitor: %s with id: %s", m.Name, m.ID))

	} else {
		_, err = service.GetByID(m.ID)
		if strings.Contains(err.Error(), "Request failed") {
			log.Info("Monitor Deleted: " + m.ID + m.Name)
		} else {
			log.Error(nil, fmt.Sprintf("Delete Request failed for Monitor: %s with id: %s", m.Name, m.ID))
		}
	}
}
