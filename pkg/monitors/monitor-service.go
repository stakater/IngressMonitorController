package monitors

import (
	log "github.com/sirupsen/logrus"

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
}

func CreateMonitorService(p *config.Provider) MonitorServiceProxy {
	monitorService := (&MonitorServiceProxy{}).OfType(p.Name)
	monitorService.Setup(*p)
	return monitorService
}

func SetupMonitorServicesForProviders(providers []config.Provider) []MonitorServiceProxy {
	if len(providers) < 1 {
		log.Panic("Cannot Instantiate controller with no providers")
	}

	log.Info("DEBUG: setupMonitorServicesForProviders providers", "providers", providers)

	monitorServices := []MonitorServiceProxy{}

	for index := 0; index < len(providers); index++ {
		monitorServices = append(monitorServices, CreateMonitorService(&providers[index]))
		log.Info("DEBUG: setupMonitorServicesForProviders added monitorServices", "monitorServices", monitorServices)
	}

	return monitorServices
}