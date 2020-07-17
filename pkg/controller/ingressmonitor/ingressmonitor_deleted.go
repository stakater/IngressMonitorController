package ingressmonitor

import (
	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileIngressMonitor) handleDelete(request reconcile.Request, instance *ingressmonitorv1alpha1.IngressMonitor, monitorName string) (reconcile.Result, error) {
	if instance == nil {
		// Instance not found, nothing to do
		return reconcile.Result{}, nil
	}

	log.Info("Removing Monitor: " + monitorName)

	// Remove monitor if it exists
	for index := 0; index < len(r.monitorServices); index++ {
		removeMonitorIfExists(r.monitorServices[index], monitorName)
	}
	return reconcile.Result{}, nil
}

func removeMonitorIfExists(monitorService monitors.MonitorServiceProxy, monitorName string) {
	monitor, _ := monitorService.GetByName(monitorName)
	// Monitor Exists
	if monitor != nil {
		// Monitor Exists, remove the monitor
		monitorService.Remove(*monitor)
	} else {
		log.Info("Cannot find monitor with name: " + monitorName + " for provider: " + monitorService.GetType())
	}
}
