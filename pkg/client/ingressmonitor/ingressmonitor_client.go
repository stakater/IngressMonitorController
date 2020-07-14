// package ingressmonitorclient
//
// import (
//
// 	log "github.com/sirupsen/logrus"
// 	"github.com/stakater/IngressMonitorController/pkg/monitors"
// 	"github.com/stakater/IngressMonitorController/pkg/config"
// 	"github.com/stakater/IngressMonitorController/pkg/models"
// )
//
// // Client is a wrapper interface for the ingressmonitorclient to allow for easier testing
// type Client interface {
// 	GetAll() []models.Monitor
// 	Add(m models.Monitor)
// 	Update(m models.Monitor)
// 	GetByName(name string) (*models.Monitor, error)
// 	Remove(m models.Monitor)
// 	Setup(p config.Provider)
// }
//
// // Client wraps http client
// type ingresMonitorClient struct {
// 	config  	config.Config
// 	monitorServices []monitors.MonitorServiceProxy
// }
//
// // NewClient creates an API client
// func NewClient() Client {
// 	log.Info("DEBUG: Instantiating controller")
//
// 	config := config.GetControllerConfig()
//
// 	return &ingresMonitorClient{
// 		config:  config,
// 		monitorServices: setupMonitorServicesForProviders(config.Providers),
// 	}
// }
//
// func getControllerConfig() Config {
// 	config := config.GetControllerConfig()
// 	return config
// }
//
// func setupMonitorServicesForProviders(providers []config.Provider) []monitors.MonitorServiceProxy {
// 	if len(providers) < 1 {
// 		log.Panic("Cannot Instantiate controller with no providers")
// 	}
//
// 	log.Info("DEBUG: setupMonitorServicesForProviders providers", "providers", providers)
//
// 	monitorServices := []monitors.MonitorServiceProxy{}
//
// 	for index := 0; index < len(providers); index++ {
// 		monitorServices = append(monitorServices, monitors.CreateMonitorService(&providers[index]))
// 		log.Info("DEBUG: setupMonitorServicesForProviders added monitorServices", "monitorServices", monitorServices)
// 	}
//
// 	return monitorServices
// }
//
