package updown

import (
	"testing"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"gotest.tools/assert"

	// "github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/util"
	// "gotest.tools/assert"
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

func TestSetupMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

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

func TestGetAllMonitor(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)
	monitorSlice := UpdownService.GetAll()

	assert.Equal(t, len(monitorSlice), 0)

}

func TestGetByNameMonitor(t *testing.T) {
	config := config.GetControllerConfig()
	UpdownService := UpdownMonitorService{}
	provider := util.GetProviderWithName(config, "Updown")
	UpdownService.Setup(*provider)

	var nilMonitorModelObj *models.Monitor

	monitorObject, _ := UpdownService.GetByName("NoExistingCheck")
	assert.Equal(t, monitorObject, nilMonitorModelObj)

}

// func TestAddMonitor(t *testing.T) {
// 	config := config.GetControllerConfig()
// 	UpdownService := UpdownMonitorService{}
// 	provider := util.GetProviderWithName(config, "Updown")
// 	UpdownService.Setup(*provider)

// }

// func test_method(t *testing.T) {
// 	config := config.GetControllerConfig()
// 	UpdownService := UpdownMonitorService{}

// 	provider := util.GetProviderWithName(config, "Updown")
// 	provider = util.GetProviderWithName(config, "X")
// 	t.Log(reflect.TypeOf(*provider))
// 	// t.Log(json.Unmarshal(provider))
// 	// delete(*provider, "apiKey")
// 	t.Log("AAAAAAAAA", config, UpdownService, provider)
// 	UpdownService.Setup(*provider)
// 	panic("Error")
// 	// assert.Equal(t, 1, 1)
// 	Block{
//         Try: func() {
//             fmt.Println("I tried")
//             Throw("Oh,...sh...")
//         },
//         Catch: func(e Exception) {
//             fmt.Printf("Caught %v\n", e)
//         },
//         Finally: func() {
//             fmt.Println("Finally...")
//         },
//     }.Do()
//     fmt.Println("We went on")

// }
