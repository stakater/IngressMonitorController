package endpointmonitor

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/kube/util"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileEndpointMonitor) handleCreate(request reconcile.Request, instance *endpointmonitorv1alpha1.EndpointMonitor, monitorName string, monitorService monitors.MonitorServiceProxy) (reconcile.Result, error) {
	log.Info("Creating Monitor: " + monitorName)

	url, err := util.GetMonitorURL(r.client, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Extract provider specific configuration
	config := monitorService.ExtractConfig(instance.Spec)

	// Create monitor Model
	monitor := models.NewMonitor(monitorName, url, config)

	// Add monitor for provider
	monitorService.Add(monitor)

	return reconcile.Result{RequeueAfter: RequeueTime}, nil
}
