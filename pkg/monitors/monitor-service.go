package monitors

import (
	"strings"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
)

type MonitorService interface {
	GetAll() []models.Monitor
	Add(m models.Monitor)
	Update(m models.Monitor)
	GetByName(name string) (*models.Monitor, error)
	Remove(m models.Monitor)
	Setup(p config.Provider)
	Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool
}

func CreateMonitorService(p *config.Provider) MonitorServiceProxy {
	monitorService := (&MonitorServiceProxy{}).OfType(p.Name)
	monitorService.Setup(*p)
	return monitorService
}

func SetupMonitorServicesForProviders(providers []config.Provider) []MonitorServiceProxy {
	if len(providers) < 1 {
		panic("Cannot Instantiate controller with no providers")
	}

	monitorServices := []MonitorServiceProxy{}

	for index := 0; index < len(providers); index++ {
		monitorServices = append(monitorServices, CreateMonitorService(&providers[index]))
		log.Info("Configuration added for " + providers[index].Name)
	}

	return monitorServices
}

func SetupMonitorServicesForProvidersTest(providers []config.Provider) []MonitorServiceProxy {
	if len(providers) < 1 {
		panic("Cannot Instantiate controller with no providers")
	}
	// TODO: Fix provider specific implementation and then add them to this list
	allowedProviders := []string{"UptimeRobot", "StatusCake"}
	log.Info("Setting up monitor services for tests(CRDs) for supported providers: " + strings.Join(allowedProviders[:], ","))

	monitorServices := []MonitorServiceProxy{}

	for index := 0; index < len(providers); index++ {
		if contains(allowedProviders, providers[index].Name) {
			monitorServices = append(monitorServices, CreateMonitorService(&providers[index]))
			log.Info("Configuration added for " + providers[index].Name)
		}
	}

	return monitorServices
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
