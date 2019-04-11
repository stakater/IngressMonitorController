package updown

import (
	"testing"
	"time"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
	"github.com/stretchr/testify/assert"
)

type Block struct {
	Try     func()
	Catch   func(Exception)
	Finally func()
}

type Exception interface{}

func Throw(up Exception) {
	panic(up)
}

func (tcf Block) Do() {
	if tcf.Finally != nil {

		defer tcf.Finally()
	}
	if tcf.Catch != nil {
		defer func() {
			if r := recover(); r != nil {
				tcf.Catch(r)
			}
		}()
	}
	tcf.Try()
}

const (
	CheckURL         = "https://updown.io"
	CheckName        = "Updown-site-check"
	UpdatedCheckName = "Update-Updown-site-check"
)

func TestSetupMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}

	provider := util.GetProviderWithName(config, "Updown")

	Block{
		Try: func() {
			UpdownService.Setup(*provider)
		},
		Catch: func(e Exception) {},
	}.Do()

}

func TestSetupMonitorWithIncorrectValues(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}

	provider := util.GetProviderWithName(config, "InvalidProviderName")

	Block{
		Try: func() {
			UpdownService.Setup(*provider)
		},
		Catch: func(e Exception) {},
	}.Do()

}

func TestGetAllMonitorWhileNoCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	monitorSlice := UpdownService.GetAll()

	assert.Equal(t, 0, len(monitorSlice))

}

func TestGetByNameMonitorWhileNoCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	var nilMonitorModelObj *models.Monitor
	monitorObject, _ := UpdownService.GetByName("NoExistingCheck")

	assert.Equal(t, monitorObject, nilMonitorModelObj)

}

func TestAddMonitorWhileNoCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	newMonitor := models.Monitor{
		URL:  CheckURL,
		Name: CheckName,
	}

	UpdownService.Add(newMonitor)
}

func TestAddMonitorWhileCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	newMonitor := models.Monitor{
		URL:  CheckURL,
		Name: CheckName,
	}

	UpdownService.Add(newMonitor)
}

func TestGetAllMonitorWhileCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	time.Sleep(30 * time.Second)
	monitorSlice2 := UpdownService.GetAll()
	firstElement := 0

	assert.Equal(t, 1, len(monitorSlice2))
	assert.Equal(t, monitorSlice2[firstElement].Name, CheckName)
	assert.Equal(t, monitorSlice2[firstElement].URL, CheckURL)

}

func TestGetByNameMonitorWhileCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	firstElement := 0
	var nilMonitorModelObj *models.Monitor
	monitorSlice := UpdownService.GetAll()
	monitorObject, _ := UpdownService.GetByName(monitorSlice[firstElement].ID)

	assert.NotEqual(t, &monitorObject, nilMonitorModelObj)

}

func TestUpdateMonitorWhileCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	firstElement := 0
	monitorSlice := UpdownService.GetAll()
	updatedMonitor := models.Monitor{
		URL:  CheckURL,
		Name: UpdatedCheckName,
		ID:   monitorSlice[firstElement].ID}

	UpdownService.Update(updatedMonitor)

}

func TestGetAllMonitorWhileCheckUpdated(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	time.Sleep(10 * time.Second)
	monitorSlice := UpdownService.GetAll()
	firstElement := 0

	assert.NotEqual(t, monitorSlice[firstElement].Name, UpdatedCheckName)

}

func TestRemoveMonitorWhileCheckExists(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	firstElement := 0
	monitorSlice := UpdownService.GetAll()
	updatedMonitor := models.Monitor{
		URL:  monitorSlice[firstElement].URL,
		Name: monitorSlice[firstElement].Name,
		ID:   monitorSlice[firstElement].ID}

	UpdownService.Remove(updatedMonitor)

}

func TestGetAllMonitorWhenCheckAreRemoved(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	time.Sleep(30 * time.Second)
	monitorSlice1 := UpdownService.GetAll()

	assert.Equal(t, 0, len(monitorSlice1))

}
