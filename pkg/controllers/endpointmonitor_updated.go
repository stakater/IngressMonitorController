package controllers

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/kube/util"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *EndpointMonitorReconciler) handleUpdate(request reconcile.Request, instance *endpointmonitorv1alpha1.EndpointMonitor, monitor models.Monitor, monitorService monitors.MonitorServiceProxy) error {
	url, err := util.GetMonitorURL(r.Client, instance)
	if err != nil {
		return err
	}

	// Extract provider specific configuration
	config := monitorService.ExtractConfig(instance.Spec)

	// Create monitor Model
	updatedMonitor := models.Monitor{Name: monitor.Name, ID: monitor.ID, URL: url, Config: config}

	// Compare and Update monitor for provider if required
	if !monitorService.Equal(monitor, updatedMonitor) {
		monitorService.Update(updatedMonitor)
	}
	return nil
}
