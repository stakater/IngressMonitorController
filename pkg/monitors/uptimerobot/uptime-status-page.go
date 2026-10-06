package uptimerobot

import (
	"encoding/json"
	"fmt"
	Http "net/http"
	"strconv"

	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("uptime-monitor-test")

type UpTimeStatusPageService struct {
	apiKey string
	url    string
}

type UpTimeStatusPage struct {
	ID       string
	Name     string
	Monitors []string
}

func (statusPage *UpTimeStatusPageService) Setup(p config.Provider) {
	statusPage.apiKey = p.ApiKey
	statusPage.url = p.ApiURL
}

func (statusPageService *UpTimeStatusPageService) requestHeaders() map[string]string {
	return v3Headers(statusPageService.apiKey)
}

func (statusPageService *UpTimeStatusPageService) Add(statusPage UpTimeStatusPage) (string, error) {
	// monitorIds [0] is the v3 "all monitors" sentinel, mirroring the v2 "monitors=0" default
	monitorIds := []int{0}
	if statusPage.Monitors != nil {
		monitorIds = monitorIdsFromStrings(statusPage.Monitors)
	}

	body, err := json.Marshal(pspRequest{
		FriendlyName: statusPage.Name,
		MonitorIds:   monitorIds,
		Status:       "ENABLED",
	})
	if err != nil {
		return "", err
	}

	response := doRequestWithRetries("POST", apiURL(statusPageService.url, "/psps"), statusPageService.requestHeaders(), body)

	if isSuccess(response) {
		var psp UptimePublicStatusPage
		if err := json.Unmarshal(response.Bytes, &psp); err != nil {
			log.Error(err, "Unable to unmarshal JSON")
			return "", err
		}
		log.Info("Status Page Added: " + statusPage.Name)
		return strconv.Itoa(psp.ID), nil
	}

	return "", fmt.Errorf("Add Status Page Request failed. Status Code: %d. Error: %s", response.StatusCode, errorExcerpt(response.Bytes))
}

func (statusPageService *UpTimeStatusPageService) Remove(statusPage UpTimeStatusPage) {
	response := doRequestWithRetries("DELETE", apiURL(statusPageService.url, "/psps/"+statusPage.ID), statusPageService.requestHeaders(), nil)

	if isSuccess(response) {
		log.Info("Status Page Removed: " + statusPage.Name)
	} else {
		log.Info("Remove Status Page Request failed. Status Code: " + strconv.Itoa(response.StatusCode) + ". Error: " + errorExcerpt(response.Bytes))
	}
}

func (statusPageService *UpTimeStatusPageService) AddMonitorToStatusPage(statusPage UpTimeStatusPage, monitor models.Monitor) (string, error) {
	existingStatusPage, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		return "", fmt.Errorf("Updated Page Request failed. Error: %w", err)
	}
	if existingStatusPage == nil {
		return "", fmt.Errorf("Status Page not found: %s", statusPage.ID)
	}

	if coversAllMonitors(existingStatusPage.Monitors) || util.ContainsString(existingStatusPage.Monitors, monitor.ID) {
		log.Info("Status Page Already Up To Date: " + statusPage.ID)
		return statusPage.ID, nil
	}

	existingStatusPage.Monitors = append(existingStatusPage.Monitors, monitor.ID)
	return statusPageService.patchStatusPageMonitors(*existingStatusPage)
}

func (statusPageService *UpTimeStatusPageService) RemoveMonitorFromStatusPage(statusPage UpTimeStatusPage, monitor models.Monitor) (string, error) {
	existingStatusPage, err := statusPageService.Get(statusPage.ID)
	if err != nil {
		return "", fmt.Errorf("Updated Page Request failed. Error: %w", err)
	}
	if existingStatusPage == nil {
		return "", fmt.Errorf("Status Page not found: %s", statusPage.ID)
	}

	// Removing a single monitor from an all-monitors PSP is not expressible in v3; leave it untouched
	if coversAllMonitors(existingStatusPage.Monitors) {
		log.Info("Status Page covers all monitors, skipping removal: " + statusPage.ID)
		return statusPage.ID, nil
	}

	existingStatusPage.Monitors = remove(existingStatusPage.Monitors, monitor.ID)
	return statusPageService.patchStatusPageMonitors(*existingStatusPage)
}

// patchStatusPageMonitors replaces the full monitorIds set of a PSP (v3 PATCH semantics)
func (statusPageService *UpTimeStatusPageService) patchStatusPageMonitors(statusPage UpTimeStatusPage) (string, error) {
	body, err := json.Marshal(pspPatchRequest{
		FriendlyName: statusPage.Name,
		MonitorIds:   monitorIdsFromStrings(statusPage.Monitors),
	})
	if err != nil {
		return "", err
	}

	response := doRequestWithRetries("PATCH", apiURL(statusPageService.url, "/psps/"+statusPage.ID), statusPageService.requestHeaders(), body)

	if isSuccess(response) {
		var psp UptimePublicStatusPage
		if err := json.Unmarshal(response.Bytes, &psp); err != nil {
			log.Error(err, "Unable to unmarshal JSON")
			return "", err
		}
		log.Info("Status Page Updated: " + statusPage.Name)
		return strconv.Itoa(psp.ID), nil
	}

	return "", fmt.Errorf("Updated Page Request failed. Status Code: %d. Error: %s", response.StatusCode, errorExcerpt(response.Bytes))
}

func (statusPageService *UpTimeStatusPageService) Get(ID string) (*UpTimeStatusPage, error) {
	response := doRequestWithRetries("GET", apiURL(statusPageService.url, "/psps/"+ID), statusPageService.requestHeaders(), nil)

	if isSuccess(response) {
		var psp UptimePublicStatusPage
		if err := json.Unmarshal(response.Bytes, &psp); err != nil {
			log.Error(err, "Unable to unmarshal JSON")
			return nil, err
		}
		return UptimeStatusPageToBaseStatusPageMapper(psp), nil
	}

	if response.StatusCode == Http.StatusNotFound {
		return nil, nil
	}

	return nil, fmt.Errorf("Get Status Page Request failed for ID: %s. Status Code: %d. Error: %s", ID, response.StatusCode, errorExcerpt(response.Bytes))
}

func (statusPageService *UpTimeStatusPageService) GetAllStatusPages(name string) ([]UpTimeStatusPage, error) {
	psps, err := statusPageService.getPsps()
	if err != nil {
		return nil, err
	}

	statusPages := []UpTimeStatusPage{}
	for _, psp := range psps {
		if psp.FriendlyName == name {
			statusPages = append(statusPages, *UptimeStatusPageToBaseStatusPageMapper(psp))
		}
	}
	return statusPages, nil
}

// getPsps lists all PSPs, following nextLink pagination until exhausted
func (statusPageService *UpTimeStatusPageService) getPsps() ([]UptimePublicStatusPage, error) {
	psps, err := paginate[UptimePublicStatusPage](apiURL(statusPageService.url, "/psps"), statusPageService.requestHeaders())
	if err != nil {
		return nil, fmt.Errorf("GetPsps request failed: %w", err)
	}
	return psps, nil
}

// GetStatusPagesForMonitor returns the IDs of the PSPs a monitor belongs to,
// read directly from the monitor's own psps field (API v3)
func (statusPageService *UpTimeStatusPageService) GetStatusPagesForMonitor(ID string) ([]string, error) {
	response := doRequestWithRetries("GET", apiURL(statusPageService.url, "/monitors/"+ID), statusPageService.requestHeaders(), nil)

	if isSuccess(response) {
		var monitor UptimeMonitorMonitor
		if err := json.Unmarshal(response.Bytes, &monitor); err != nil {
			return nil, err
		}

		var pspIDs []string
		for _, psp := range monitor.Psps {
			pspIDs = append(pspIDs, strconv.Itoa(psp.ID))
		}
		return pspIDs, nil
	}

	if response.StatusCode == Http.StatusNotFound {
		return nil, nil
	}

	return nil, fmt.Errorf("GetStatusPagesForMonitor Request failed for ID: %s. Status Code: %d. Error: %s", ID, response.StatusCode, errorExcerpt(response.Bytes))
}

// coversAllMonitors reports whether the monitor list is the v3 [0] "all monitors" sentinel
func coversAllMonitors(monitors []string) bool {
	return len(monitors) == 1 && monitors[0] == "0"
}

func monitorIdsFromStrings(monitors []string) []int {
	ids := []int{}
	for _, monitor := range monitors {
		if id, err := strconv.Atoi(monitor); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func remove(s []string, i string) []string {
	j := 0
	for _, n := range s {
		if n != i {
			s[j] = n
			j++
		}
	}
	s = s[:j]
	return s
}
