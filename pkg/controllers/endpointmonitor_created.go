package controllers

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/kube/util"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *EndpointMonitorReconciler) handleCreate(request reconcile.Request, instance *endpointmonitorv1alpha1.EndpointMonitor, monitorName string, monitorService monitors.MonitorServiceProxy) error {
	log := r.Log.WithValues("endpointMonitor", instance.ObjectMeta.Namespace)

	log.Info("Creating Monitor: " + monitorName)

	url, err := util.GetMonitorURL(r.Client, instance)
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
