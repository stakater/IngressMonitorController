package ingressmonitor

import (
	"fmt"
	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/kube/util"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileIngressMonitor) handleCreate(request reconcile.Request, instance *ingressmonitorv1alpha1.IngressMonitor, monitorName string, monitorService monitors.MonitorServiceProxy) (reconcile.Result, error) {
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

	return reconcile.Result{RequeueAfter: defaultRequeueTime}, nil
}
