package statuscake

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

const (
	StatuscakeCheckRateAnnotation      = "statuscake.monitor.stakater.com/check-rate"
	StatuscakeTestTypeAnnotation       = "statuscake.monitor.stakater.com/test-type"
	StatuscakePausedAnnotation         = "statuscake.monitor.stakater.com/paused"
	StatuscakePingURLAnnotation        = "statuscake.monitor.stakater.com/ping-url"
	StatuscakeFollowRedirectAnnotation = "statuscake.monitor.stakater.com/follow-redirect"
	StatuscakePortAnnotation           = "statuscake.monitor.stakater.com/port"
	StatuscakeTriggerRateAnnotation    = "statuscake.monitor.stakater.com/trigger-rate"
	StatuscakeContactGroupAnnotation   = "statuscake.monitor.stakater.com/contact-group"
	StatuscakeBasicAuthUserAnnotation  = "statuscake.monitor.stakater.com/basic-auth-user"
)

// StatusCakeMonitorService is the service structure for StatusCake
type StatusCakeMonitorService struct {
	apiKey   string
	url      string
	username string
	cgroup   string
	client   *http.Client
}

// AnnotationInfo is the annotation information structure for AnnotationMap
type AnnotationInfo struct {
	name     string
	dataType string
}

// AnnotationMap holds all the enabled annotations for StatusCake
var AnnotationMap = map[string]AnnotationInfo{
	StatuscakeCheckRateAnnotation:      AnnotationInfo{"CheckRate", "int"},       // Int (0-24000)
	StatuscakeTestTypeAnnotation:       AnnotationInfo{"TestType", "string"},     // String (HTTP, TCP, PING)
	StatuscakePausedAnnotation:         AnnotationInfo{"Paused", "bool"},         // Int (0,1)
	StatuscakePingURLAnnotation:        AnnotationInfo{"PingURL", "string"},      // String (url)
	StatuscakeFollowRedirectAnnotation: AnnotationInfo{"FollowRedirect", "bool"}, // Int (0,1)
	StatuscakePortAnnotation:           AnnotationInfo{"Port", "int"},            // Int (TCP Port)
	StatuscakeTriggerRateAnnotation:    AnnotationInfo{"TriggerRate", "int"},     // Int (0-60)
	StatuscakeContactGroupAnnotation:   AnnotationInfo{"ContactGroup", "string"}, // String (0-60)
	StatuscakeBasicAuthUserAnnotation:  AnnotationInfo{"BasicUser", "string"},    // String (0-60)
}

// buildUpsertForm function is used to create the form needed to Add or update a monitor
func buildUpsertForm(m models.Monitor, cgroup string) url.Values {
	f := url.Values{}
	f.Add("WebsiteName", m.Name)
	unEscapedURL, _ := url.QueryUnescape(m.URL)
	f.Add("WebsiteURL", unEscapedURL)
	if val, ok := m.Annotations[StatuscakeCheckRateAnnotation]; ok {
		f.Add("CheckRate", val)
		delete(m.Annotations, StatuscakeCheckRateAnnotation)
	} else {
		f.Add("CheckRate", "300")
	}
	if val, ok := m.Annotations[StatuscakeTestTypeAnnotation]; ok {
		f.Add("TestType", val)
		delete(m.Annotations, StatuscakeTestTypeAnnotation)
	} else {
		f.Add("TestType", "HTTP")
	}
	if val, ok := m.Annotations[StatuscakeContactGroupAnnotation]; ok {
		f.Add("ContactGroup", val)
		delete(m.Annotations, StatuscakeContactGroupAnnotation)
	} else {
		if cgroup != "" {
			f.Add("ContactGroup", cgroup)
		}
	}
	// If a basic auth annotation is set
	if val, ok := m.Annotations[StatuscakeBasicAuthUserAnnotation]; ok {
		// Annotation should be set to the username
		// Environment variable should define the password
		// Mounted via a secret; key is the username, value is the password
		basicPass := os.Getenv(val)
		if basicPass != "" {
			f.Add("BasicUser", val)
			f.Add("BasicPass", basicPass)
			log.Println("Basic auth requirement detected. Setting username and password")
		} else {
			log.Println("Error reading basic auth password from environment variable")
		}
		delete(m.Annotations, StatuscakeBasicAuthUserAnnotation)
	}

	for key, value := range m.Annotations {
		if (AnnotationInfo{}) != AnnotationMap[key] {
			meta := AnnotationMap[key]
			switch strings.ToLower(meta.dataType) {
			case "int":
				f.Add(meta.name, value)
			case "string":
				f.Add(meta.name, value)
			case "bool":
				value = strings.ToLower(value)
				if value == "true" {
					f.Add(meta.name, "1")
				} else {
					f.Add(meta.name, "0")
				}
			}
		}
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
		log.Println(err)
		return nil
	}
	u.Path = "/API/Tests/"
	u.Scheme = "https"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	if resp.StatusCode == http.StatusOK {
		f := make([]StatusCakeMonitorMonitor, 0)
		err := json.NewDecoder(resp.Body).Decode(&f)
		if err != nil {
			log.Println(err)
			return nil
		}
		return StatusCakeMonitorMonitorsToBaseMonitorsMapper(f)
	}
	errorString := "GetAll Request failed"
	log.Println(errorString)
	return nil
}

// Add will create a new Monitor
func (service *StatusCakeMonitorService) Add(m models.Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = "/API/Tests/Update"
	u.Scheme = "https"
	data := buildUpsertForm(m, service.cgroup)
	req, err := http.NewRequest("PUT", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Println(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Added:", fa.InsertID)
		} else {
			log.Println("Monitor couldn't be added: " + m.Name)
			log.Println(fa.Message)
		}
	} else {
		errorString := "Insert Request failed for name: " + m.Name
		log.Println(errorString)
	}
}

// Update will update an existing Monitor
func (service *StatusCakeMonitorService) Update(m models.Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = "/API/Tests/Update"
	u.Scheme = "https"
	data := buildUpsertForm(m, service.cgroup)
	data.Add("TestID", m.ID)
	req, err := http.NewRequest("PUT", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Println(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Updated:", m.Name)
		} else {
			log.Println("Monitor couldn't be updated: " + m.Name)
			log.Println(fa.Message)
		}
	} else {
		errorString := "Update Request failed for name: " + m.Name
		log.Println(errorString)
	}
}

// Remove will delete an existing Monitor
func (service *StatusCakeMonitorService) Remove(m models.Monitor) {
	u, err := url.Parse(service.url)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = "/API/Tests/Details"
	u.Scheme = "https"
	query := u.Query()
	query.Set("TestID", m.ID)
	u.RawQuery = query.Encode()
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("API", service.apiKey)
	req.Header.Add("Username", service.username)
	resp, err := service.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		var fa StatusCakeUpsertResponse
		err := json.NewDecoder(resp.Body).Decode(&fa)
		if err != nil {
			log.Println(err)
			return
		}
		if fa.Success {
			log.Println("Monitor Deleted:", m.ID)
		} else {
			log.Println("Monitor couldn't be deleted: " + m.Name)
			log.Println(fa.Message)
		}
	} else {
		errorString := "Delete Request failed for name: " + m.Name
		log.Println(errorString)
	}
}
