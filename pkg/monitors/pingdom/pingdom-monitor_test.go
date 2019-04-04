package pingdom

// import (
// 	"testing"

// 	"github.com/stakater/IngressMonitorController/pkg/config"
// 	"github.com/stakater/IngressMonitorController/pkg/models"
// )

// func TestAddPingdomMonitorWithCorrectValues(t *testing.T) {
// 	config := config.GetControllerConfig()

// 	service := PingdomMonitorService{}
// 	service.Setup(config.Providers[1])

// 	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
// 	service.Add(m)

// 	mRes, err := service.GetByName("google-test")

// 	if err != nil {
// 		t.Error("Error: " + err.Error())
// 	}
// 	if mRes != nil {
// 		if mRes.Name != m.Name || mRes.URL != m.URL {
// 			t.Error("URL and name should be the same")
// 		}
// 	}
// 	if mRes != nil {
// 		service.Remove(*mRes)
// 	}
// }

// func TestUpdateMonitorWithCorrectValues(t *testing.T) {
// 	config := config.GetControllerConfig()

// 	service := PingdomMonitorService{}
// 	service.Setup(config.Providers[1])

// 	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
// 	service.Add(m)

// 	mRes, err := service.GetByName("google-test")

// 	if err != nil {
// 		t.Error("Error: " + err.Error())
// 	}
// 	if mRes != nil {
// 		if mRes.Name != m.Name || mRes.URL != m.URL {
// 			t.Error("URL and name should be the same")
// 		}

// 		mRes.URL = "https://facebook.com"

// 		service.Update(*mRes)

// 		mRes, err = service.GetByName("google-test")

// 		if err != nil {
// 			t.Error("Error: " + err.Error())
// 		}
// 		if mRes.URL != "https://facebook.com" {
// 			t.Error("URL and name should be the same")
// 		}
// 		service.Remove(*mRes)
// 	}
// }

// func TestAddMonitorWithIncorrectValues(t *testing.T) {
// 	config := config.GetControllerConfig()

// 	service := PingdomMonitorService{}
// 	config.Providers[1].ApiKey = "dummy-api-key"
// 	service.Setup(config.Providers[1])

// 	m := models.Monitor{Name: "google-test", URL: "https://google.com"}
// 	service.Add(m)

// 	mRes, err := service.GetByName("google-test")

// 	if err != nil {
// 		t.Error("Error: " + err.Error())
// 	}

// 	if mRes != nil {
// 		t.Error("Monitor should not be added")
// 	}
// }

