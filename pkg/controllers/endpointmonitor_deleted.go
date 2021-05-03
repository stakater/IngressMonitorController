package controllers

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *EndpointMonitorReconciler) handleDelete(request reconcile.Request, instance *endpointmonitorv1alpha1.EndpointMonitor, monitorName string) (reconcile.Result, error) {
	log := r.Log.WithValues("endpointMonitor", request.Namespace)

	if instance == nil {
		// Instance not found, nothing to do
		return reconcile.Result{}, nil
	}

	if !config.GetControllerConfig().EnableMonitorDeletion {
		log.Info("Monitor deletion is disabled. Skipping deletion for monitor: " + monitorName)
		return reconcile.Result{}, nil
	}

	log.Info("Removing Monitor: " + monitorName)

	// Remove monitor if it exists
	for index := 0; index < len(r.MonitorServices); index++ {
		r.removeMonitorIfExists(r.MonitorServices[index], monitorName)
	}
	return reconcile.Result{}, nil
}

func (r *EndpointMonitorReconciler) removeMonitorIfExists(monitorService monitors.MonitorServiceProxy, monitorName string) {
	log := r.Log.WithValues("monitor", monitorName)

	monitor, _ := monitorService.GetByName(monitorName)
	// Monitor Exists
	if monitor != nil {
		// Monitor Exists, remove the monitor
		log.Info("Removing monitor with name: " + monitorName + " for provider: " + monitorService.GetType())
		monitorService.Remove(*monitor)
	} else {
		log.Info("Cannot find monitor with name: " + monitorName + " for provider: " + monitorService.GetType())
	}
}
