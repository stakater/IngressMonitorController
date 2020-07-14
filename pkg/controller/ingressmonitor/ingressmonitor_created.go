package ingressmonitor

import (
	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/monitors"
	"github.com/stakater/IngressMonitorController/pkg/models"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileIngressMonitor) handleCreate(request reconcile.Request, instance *ingressmonitorv1alpha1.IngressMonitor, monitorName string, monitorService monitors.MonitorServiceProxy) (reconcile.Result, error) {
	log.Info("Creating Monitor: " + monitorName)

	log.Info("DEBUG: monitorName: " + monitorName)
	log.Info("DEBUG: instance.Spec.URL: " + instance.Spec.URL)

	m := models.Monitor{
		Name:        monitorName,
		URL:         instance.Spec.URL,
		// TODO: Add fields corresponding to annotations
		// TODO: Handle urlFrom
	}

	// Add monitor
	// TODO: Generate error in case of proper and handle it
	monitorService.Add(m)

	return reconcile.Result{}, nil
}