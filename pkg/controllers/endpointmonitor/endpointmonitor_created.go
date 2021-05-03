package endpointmonitor

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/kube/util"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileEndpointMonitor) handleCreate(request reconcile.Request, instance *endpointmonitorv1alpha1.EndpointMonitor, monitorName string, monitorService monitors.MonitorServiceProxy) error {
	log.Info("Creating Monitor: " + monitorName)

	url, err := util.GetMonitorURL(r.client, instance)
	if err != nil {
		return err
	}

	// Extract provider specific configuration
	providerConfig := monitorService.ExtractConfig(instance.Spec)

	// Create monitor Model
	monitor := models.Monitor{Name: monitorName, URL: url, Config: providerConfig}

	// Add monitor for provider
	monitorService.Add(monitor)

	return nil
}
