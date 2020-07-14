package ingressmonitor

import (
	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/monitors"
	"github.com/stakater/IngressMonitorController/pkg/models"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileIngressMonitor) handleUpdate(request reconcile.Request, instance *ingressmonitorv1alpha1.IngressMonitor, monitor models.Monitor, monitorService monitors.MonitorServiceProxy) (reconcile.Result, error) {
	log.Info("Updating Monitor: " + monitor.Name)

	log.Info("DEBUG: instance.Spec.URL: " + instance.Spec.URL)

	monitor.URL = instance.Spec.URL
	monitorService.Update(monitor)
	return reconcile.Result{}, nil
}