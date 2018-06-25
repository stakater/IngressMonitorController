package monitors

import (
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
