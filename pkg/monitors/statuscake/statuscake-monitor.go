package statuscake

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

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
	f.Add("WebsiteName", m.Name)
	unEscapedURL, _ := url.QueryUnescape(m.URL)
	f.Add("WebsiteURL", unEscapedURL)

	// Retrieve provider configuration
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.StatusCakeConfig)

	if providerConfig != nil && providerConfig.CheckRate > 0 {
		f.Add("CheckRate", strconv.Itoa(providerConfig.CheckRate))
	} else {
		f.Add("CheckRate", "300")
	}

	if providerConfig != nil && len(providerConfig.TestType) > 0 {
		f.Add("TestType", providerConfig.TestType)
	} else {
		f.Add("TestType", "HTTP")
	}

	if providerConfig != nil && len(providerConfig.ContactGroup) > 0 {
		f.Add("ContactGroup", providerConfig.ContactGroup)
	} else {
		if cgroup != "" {
			f.Add("ContactGroup", cgroup)
		}
	}

	if providerConfig != nil && len(providerConfig.TestTags) > 0 {
		f.Add("TestTags", providerConfig.TestTags)
	}

	if providerConfig != nil && len(providerConfig.BasicAuthUser) > 0 {
		// This value is mandatory
		// Environment variable should define the password
		// Mounted via a secret; key is the username, value is the password
		basicPass := os.Getenv(providerConfig.BasicAuthUser)
		if basicPass != "" {
			f.Add("BasicUser", providerConfig.BasicAuthUser)
			f.Add("BasicPass", basicPass)
			log.Println("Basic auth requirement detected. Setting username and password")
		} else {
			log.Println("Error reading basic auth password from environment variable")
		}
	}

	if providerConfig != nil && len(providerConfig.StatusCodes) > 0 {
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
		f.Add("StatusCodes", strings.Join(statusCodes, ","))
	}

	if providerConfig != nil {
		if providerConfig.Paused {
			f.Add("Paused", "1")
		}
		if providerConfig.FollowRedirect {
			f.Add("FollowRedirect", "1")
		}
		if providerConfig.EnableSSLAlert {
			f.Add("EnableSSLAlert", "1")
		}
		if providerConfig.RealBrowser {
			f.Add("RealBrowser", "1")
		}
	}

	if providerConfig != nil && len(providerConfig.PingURL) > 0 {
		f.Add("PingURL", providerConfig.PingURL)
	}

	if providerConfig != nil && len(providerConfig.NodeLocations) > 0 {
		f.Add("NodeLocations", providerConfig.NodeLocations)
	}

	if providerConfig != nil && providerConfig.TriggerRate > 0 {
		f.Add("TriggerRate", strconv.Itoa(providerConfig.TriggerRate))
	}

	if providerConfig != nil && providerConfig.Confirmation > 0 {
		f.Add("TriggerRate", strconv.Itoa(providerConfig.Confirmation))
	}
	return f
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
	for _, monitor := range monitors {
		if monitor.Name == name {
			return &monitor, nil
		}
	}
	errorString := "GetByName Request failed for name: " + name
	return nil, errors.New(errorString)
}

// GetAll function will fetch all monitors
func (service *StatusCakeMonitorService) GetAll() []models.Monitor {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err)
		return nil
	}
	u.Path = "/API/Tests/"
	u.Scheme = "https"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Error(err)
		return nil
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Error(err)
		return nil
	}
	if resp.StatusCode == http.StatusOK {
		f := make([]StatusCakeMonitorMonitor, 0)
		err := json.NewDecoder(resp.Body).Decode(&f)
		if err != nil {
			log.Error(err)
			return nil
		}
		return StatusCakeMonitorMonitorsToBaseMonitorsMapper(f)
	}
	errorString := "GetAll Request failed"
	log.Error(errorString)
	return nil
}

// Add will create a new Monitor
func (service *StatusCakeMonitorService) Add(m models.Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err)
		return
	}
	u.Path = "/API/Tests/Update"
	u.Scheme = "https"
	data := buildUpsertForm(m, service.cgroup)
	req, err := http.NewRequest("PUT", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Error(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Error(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Added:", fa.InsertID)
		} else {
			log.Println("Monitor couldn't be added: " + m.Name)
			log.Println(fa.Message)
		}
	} else {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		errorString := "Insert Request failed for name: " + m.Name + " with status code " + strconv.Itoa(resp.StatusCode)
		log.Error(errorString)
		log.Error(bodyString)
	}
}

// Update will update an existing Monitor
func (service *StatusCakeMonitorService) Update(m models.Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err)
		return
	}
	u.Path = "/API/Tests/Update"
	u.Scheme = "https"
	data := buildUpsertForm(m, service.cgroup)
	data.Add("TestID", m.ID)
	req, err := http.NewRequest("PUT", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Error(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Error(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Updated:", m.Name)
		} else {
			log.Warn("Monitor couldn't be updated: " + m.Name)
			log.Warn(fa.Message)
		}
	} else {
		errorString := "Update Request failed for name: " + m.Name
		log.Error(errorString)
	}
}

// Remove will delete an existing Monitor
func (service *StatusCakeMonitorService) Remove(m models.Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Error(err)
		return
	}
	u.Path = "/API/Tests/Details"
	u.Scheme = "https"
	query := u.Query()
	query.Set("TestID", m.ID)
	u.RawQuery = query.Encode()
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		log.Error(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Error(err)
			return
		}
		if fa.Success {
			log.Info("Monitor Deleted:", m.ID)
		} else {
			log.Warn("Monitor couldn't be deleted: " + m.Name)
			log.Warn(fa.Message)
		}
	} else {
		errorString := "Delete Request failed for name: " + m.Name
		log.Error(errorString)
	}
}
