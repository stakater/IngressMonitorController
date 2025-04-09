package alicloud

import (
	"fmt"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// To allow normal logging to be printed if tests fails
	// Dev mode is an extra feature to make output more readable
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := AliCloudMonitorService{}
	provider := util.GetProviderWithName(config, monitors.TypeAliCloud)
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error(nil, "Failed to find provider")
		return
	}
	service.Setup(*provider)
	m := models.Monitor{Name: "AliCloud-test", URL: "metrics.cn-qingdao.aliyuncs.com"}
	service.Add(m)

	mRes, err := service.GetByName("AliCloud-test")
	assert.Nil(t, err)
	assert.NotNil(t, mRes)

	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Errorf("URL and name should be the same. request: %+v response: %+v", m, mRes)
		return
	}

	// Cleanup
	service.Remove(*mRes)
	monitor, err := service.GetByName(mRes.Name)
	assert.Nil(t, monitor, fmt.Sprintf("Monitor should've been deleted %v %v", monitor, err))
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := AliCloudMonitorService{}

	provider := util.GetProviderWithName(config, monitors.TypeAliCloud)
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error(nil, "Failed to find provider")
		return
	}
	service.Setup(*provider)

	// Create initial record
	m := models.Monitor{Name: "AliCloud-test", URL: "https://aliCloud.com"}
	service.Add(m)

	mRes, err := service.GetByName("AliCloud-test")
	assert.Nil(t, err)
	assert.NotNil(t, mRes)

	if mRes.Name != m.Name || mRes.URL != m.URL {
		t.Errorf("URL and name should be the same. request: %+v response: %+v", m, mRes)
	}

	targetUrl := "https://facebook.com"
	// Update the record
	mRes.URL = targetUrl

	service.Update(*mRes)

	mRes, err = service.GetByName("AliCloud-test")
	assert.Nil(t, err)
	if mRes.Name != m.Name || mRes.URL != targetUrl {
		t.Errorf("URL and name should be the same. request: %+v response: %+v", m, mRes)
	}

	// Cleanup
	service.Remove(*mRes)
	monitor, err := service.GetByName(mRes.Name)
	assert.Nil(t, monitor, fmt.Sprintf("Monitor should've been deleted %v %v", monitor, err))
}
