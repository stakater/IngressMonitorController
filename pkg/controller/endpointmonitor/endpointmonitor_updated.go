package endpointmonitor

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/kube/util"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileEndpointMonitor) handleUpdate(request reconcile.Request, instance *endpointmonitorv1alpha1.EndpointMonitor, monitor models.Monitor, monitorService monitors.MonitorServiceProxy) (reconcile.Result, error) {
	log.Info("Updating Monitor: " + monitor.Name)

	url, err := util.GetMonitorURL(r.client, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Extract provider specific configuration
	config := monitorService.ExtractConfig(instance.Spec)

	// Create monitor Model
	updatedMonitor := models.NewMonitor(monitor.Name, url, config)

	// Update monitor for provider
	monitorService.Update(updatedMonitor)
	return reconcile.Result{RequeueAfter: RequeueTime}, nil
}
