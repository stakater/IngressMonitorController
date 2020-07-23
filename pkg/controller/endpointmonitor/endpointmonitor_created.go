package endpointmonitor

import (
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/kube/util"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	"fmt"
	"time"

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
	providerConfig := monitorService.ExtractConfig(instance.Spec)

	// Handle CreationDelay
	createTime := instance.CreationTimestamp
	delay := time.Until(createTime.Add(config.GetControllerConfig().CreationDelay))

	if delay.Nanoseconds() > 0 {
		// Requeue request to add creation delay
		log.Info("Requeuing request to add monitor " + monitorName + " for" + fmt.Sprintf("%+v", config.GetControllerConfig().CreationDelay) + " seconds")
		return reconcile.Result{RequeueAfter: delay}, nil
	}

	// Create monitor Model
	monitor := models.Monitor{Name: monitorName, URL: url, Config: providerConfig}

	// Add monitor for provider
	monitorService.Add(monitor)

	return reconcile.Result{RequeueAfter: RequeueTime}, nil
}
