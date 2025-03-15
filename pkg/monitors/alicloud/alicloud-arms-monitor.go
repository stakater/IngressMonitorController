package alicloud

import (
	"encoding/json"
	"errors"
	"fmt"
	cms20190101 "github.com/alibabacloud-go/cms-20190101/v9/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/openapi/ptr"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

var log = logf.Log.WithName("AliCloud")

var _ monitors.MonitorService = &AliCloudMonitorService{}

// AliCloudMonitorService struct contains parameters required by AliCloud go client
type AliCloudMonitorService struct {
	accessKey string
	secretKey string
	endpoint  string
	client    *cms20190101.Client
}

func (monitor *AliCloudMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {
	return reflect.DeepEqual(oldMonitor, newMonitor)
}

// Setup method will initialize a AliCloud's go client object by using the configuration parameters
func (monitor *AliCloudMonitorService) Setup(confProvider config.Provider) {

	// initializeCustomLog(os.Stdout)
	log.Info("AliCloud monitor's Setup has been called. AliCloud monitor initializing")

	// AliCloud go client apiKey
	monitor.accessKey = confProvider.ApiKey
	monitor.secretKey = confProvider.ApiToken
	monitor.endpoint = monitor.formatEndpoint(confProvider.ApiURL)

	// creating AliCloud go client
	client, err := monitor.newClient()
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize AliCloud client: %v", err))
	}
	monitor.client = client
	log.Info("AliCloud monitor has been initialized")
}

// GetAll function will return all monitors (AliCloud checks) object in an array
func (monitor *AliCloudMonitorService) GetAll() []models.Monitor {
	log.Info("AliCloud monitor's GetAll method has been called")

	describeSiteMonitorListRequest := &cms20190101.DescribeSiteMonitorListRequest{}
	request, err := monitor.getByListRequest(describeSiteMonitorListRequest)
	if err != nil {
		log.Error(err, "failed to get AliCloud checks", "describeSiteMonitorListRequest", describeSiteMonitorListRequest.String())
	}
	return request
}

func (monitor *AliCloudMonitorService) getByListRequest(describeSiteMonitorListRequest *cms20190101.DescribeSiteMonitorListRequest) ([]models.Monitor, error) {
	var allMonitors []models.Monitor
	var listResult *cms20190101.DescribeSiteMonitorListResponse
	runtime := &util.RuntimeOptions{}
	tryErr := func() (err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		listResult, err = monitor.client.DescribeSiteMonitorListWithOptions(describeSiteMonitorListRequest, runtime)
		return err
	}()

	if tryErr != nil {
		return allMonitors, monitor.formatTryError(tryErr)
	}

	if listResult.Body == nil || listResult.Body.TotalCount == nil || listResult.Body.SiteMonitors == nil {
		return allMonitors, fmt.Errorf("listResult.Body.SiteMonitors is nil")
	}

	// populating a allMonitors slice using the AliCloudChecks objects given in AliCloudChecks slice
	for _, aliCloudCheck := range listResult.Body.SiteMonitors.SiteMonitor {
		newMonitor := models.Monitor{
			URL:  *aliCloudCheck.Address,
			Name: *aliCloudCheck.TaskName,
			ID:   *aliCloudCheck.TaskId,
		}
		allMonitors = append(allMonitors, newMonitor)
	}
	return allMonitors, nil
}

func (monitor *AliCloudMonitorService) formatTryError(tryErr error) error {
	var err = &tea.SDKError{}
	if _t, ok := tryErr.(*tea.SDKError); ok {
		err = _t
	} else {
		err.Message = tea.String(tryErr.Error())
	}
	result, assertErr := util.AssertAsString(err.Message)
	if assertErr == nil {
		return errors.New(tea.StringValue(result))
	}
	return errors.New(err.Error())
}

// GetByName function will return a monitor(AliCloud check) object based on the name provided
func (monitor *AliCloudMonitorService) GetByName(monitorName string) (*models.Monitor, error) {

	log.Info("AliCloud monitor's GetByName method has been called")

	describeSiteMonitorListRequest := &cms20190101.DescribeSiteMonitorListRequest{
		Keyword: ptr.String(monitorName),
	}
	byListRequest, err := monitor.getByListRequest(describeSiteMonitorListRequest)
	if err != nil {
		return nil, err
	}
	if len(byListRequest) < 1 {
		return nil, nil
	}
	return &byListRequest[0], nil
}

func (monitor *AliCloudMonitorService) getSiteMonitorCreateInstance(aliCloudMonitor models.Monitor) *cms20190101.CreateSiteMonitorRequest {
	singleTask := &cms20190101.CreateSiteMonitorRequest{
		Address:  tea.String(aliCloudMonitor.URL),
		TaskName: tea.String(aliCloudMonitor.Name),
	}
	if aliCloudMonitor.Config != nil {
		providerConfig, ok := aliCloudMonitor.Config.(*endpointmonitorv1alpha1.AliCloudConfig)
		if !ok {
			panic(fmt.Errorf("unable to locate %v monitor", aliCloudMonitor.Name))
		}
		singleTask.TaskType = tea.String(providerConfig.TaskType)
		singleTask.OptionsJson = tea.String(providerConfig.OptionsJson)
	}

	if singleTask.TaskType == nil {
		singleTask.TaskType = ptr.String("HTTP")
	}
	return singleTask
}

func (monitor *AliCloudMonitorService) getSiteMonitorUpdateInstance(aliCloudMonitor models.Monitor) *cms20190101.ModifySiteMonitorRequest {
	singleTask := &cms20190101.ModifySiteMonitorRequest{
		Address:  tea.String(aliCloudMonitor.URL),
		TaskName: tea.String(aliCloudMonitor.Name),
		TaskId:   ptr.String(aliCloudMonitor.ID),
	}
	if aliCloudMonitor.Config != nil {
		providerConfig, ok := aliCloudMonitor.Config.(*endpointmonitorv1alpha1.AliCloudConfig)
		if !ok {
			panic(fmt.Errorf("unable to locate %v monitor", aliCloudMonitor.Name))
		}
		singleTask.OptionsJson = tea.String(providerConfig.OptionsJson)
	}
	return singleTask
}

// Add function method will add a monitor (AliCloud check)
func (monitor *AliCloudMonitorService) Add(aliCloudMonitor models.Monitor) {

	log.Info("AliCloud monitor's Add method has been called")

	createInstance := monitor.getSiteMonitorCreateInstance(aliCloudMonitor)
	runtime := &util.RuntimeOptions{}
	var createResult *cms20190101.CreateSiteMonitorResponse
	tryErr := func() (err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		createResult, err = monitor.client.CreateSiteMonitorWithOptions(createInstance, runtime)
		return err
	}()
	if tryErr != nil {
		log.Error(monitor.formatTryError(tryErr), "failed to create AliCloud check", "createInstance", createInstance.String())
		return
	}

	if createResult.Body == nil {
		log.Error(errors.New("createResult is nil"), "createResult is nil")
		return
	}

	if createResult.Body.Success == nil || *createResult.Body.Success != "true" {
		bodyJson, _ := json.Marshal(createResult.Body)
		log.Error(errors.New("failed to create AliCloud check"), "failed to create AliCloud check", "body", bodyJson)
		return
	}
	log.Info("Monitor addition request has been completed", "createResult", createResult.Body.CreateResultList)
}

// Update method will update a monitor (AliCloud check)
func (monitor *AliCloudMonitorService) Update(aliCloudMonitor models.Monitor) {

	log.Info("AliCloud's Update method has been called")

	getInstance := monitor.getSiteMonitorUpdateInstance(aliCloudMonitor)
	runtime := &util.RuntimeOptions{}
	var updateResult *cms20190101.ModifySiteMonitorResponse
	tryErr := func() (err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		updateResult, err = monitor.client.ModifySiteMonitorWithOptions(getInstance, runtime)
		return err
	}()
	if tryErr != nil {
		log.Error(monitor.formatTryError(tryErr), "failed to update AliCloud check", "getInstance", getInstance.String())
		return
	}

	if updateResult.Body == nil {
		log.Error(errors.New("createResult is nil"), "createResult is nil")
		return
	}

	if updateResult.Body.Success == nil || *updateResult.Body.Success != "true" {
		bodyJson, _ := json.Marshal(updateResult.Body)
		log.Error(errors.New("failed to update AliCloud check"), "failed to create AliCloud check", "body", bodyJson)
		return
	}
	log.Info("AliCloud's check Update request has been completed", "updateResult", updateResult.String())
}

// Remove method will remove a monitor (AliCloud check)
func (monitor *AliCloudMonitorService) Remove(aliCloudMonitor models.Monitor) {

	log.Info("AliCloud's Remove method has been called")

	deleteSiteMonitorsRequest := &cms20190101.DeleteSiteMonitorsRequest{
		TaskIds: tea.String(aliCloudMonitor.ID),
	}
	var deleteResult *cms20190101.DeleteSiteMonitorsResponse
	runtime := &util.RuntimeOptions{}
	tryErr := func() (err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		deleteResult, err = monitor.client.DeleteSiteMonitorsWithOptions(deleteSiteMonitorsRequest, runtime)
		return err
	}()

	if tryErr != nil {
		log.Error(monitor.formatTryError(tryErr), "failed to delete AliCloud check", "deleteSiteMonitorsRequest", deleteSiteMonitorsRequest.String())
		return
	}
	if deleteResult.Body == nil {
		log.Error(errors.New("createResult is nil"), "createResult is nil")
		return
	}

	if deleteResult.Body.Success == nil || *deleteResult.Body.Success != "true" {
		bodyJson, _ := json.Marshal(deleteResult.Body)
		log.Error(errors.New("failed to delete AliCloud check"), "failed to create AliCloud check", "body", bodyJson)
		return
	}
	log.Info("AliCloud's check Remove request has been completed", "deleteResult", deleteResult.Body)
}

func (monitor *AliCloudMonitorService) newClient() (*cms20190101.Client, error) {
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(monitor.accessKey),
		AccessKeySecret: tea.String(monitor.secretKey),
	}
	// Endpoint refers to: https://api.aliyun.com/product/Cms
	cfg.Endpoint = tea.String(monitor.endpoint)
	return cms20190101.NewClient(cfg)
}

func (monitor *AliCloudMonitorService) formatEndpoint(url string) string {
	if strings.Trim(url, " ") == "" {
		return "metrics.cn-qingdao.aliyuncs.com"
	}
	if strings.HasPrefix(url, "http://") {
		return strings.Split(url, "http://")[1]
	}
	if strings.HasPrefix(url, "https://") {
		return strings.Split(url, "https://")[1]
	}
	return url
}
