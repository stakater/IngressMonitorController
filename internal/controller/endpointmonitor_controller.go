/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
)

var log = logf.Log.WithName("endpointmonitor-controller")

// EndpointMonitorReconciler reconciles a EndpointMonitor object
type EndpointMonitorReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	MonitorServices []*monitors.MonitorServiceProxy
}

//+kubebuilder:rbac:groups=endpointmonitor.stakater.com,resources=endpointmonitors,verbs=get;list;watch
//+kubebuilder:rbac:groups=endpointmonitor.stakater.com,resources=endpointmonitors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=endpointmonitor.stakater.com,resources=endpointmonitors/finalizers,verbs=update
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *EndpointMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("endpointmonitor", req.NamespacedName)

	// Fetch the EndpointMonitor instance
	instance := &endpointmonitorv1alpha1.EndpointMonitor{}

	var monitorName string
	format, err := util.GetNameTemplateFormat(config.GetControllerConfig().MonitorNameTemplate)
	if err != nil {
		log.Error(err, "Failed to parse MonitorNameTemplate, using default template `{{.Name}}-{{.Namespace}}`")
		monitorName = req.Name + "-" + req.Namespace
	} else {
		monitorName = fmt.Sprintf(format, req.Name, req.Namespace)
	}

	err = r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return r.handleDelete(req, instance, monitorName)
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Handle CreationDelay
	createTime := instance.CreationTimestamp
	delay := time.Until(createTime.Add(config.GetControllerConfig().CreationDelay))

	monitorService := r.GetMonitorOfType(instance.Spec)
	monitor := findMonitorByName(monitorService, monitorName)
	if monitor != nil {
		// Monitor already exists, update if required
		err = r.handleUpdate(req, instance, *monitor, monitorService)
	} else {
		// Monitor doesn't exist, create monitor
		if delay.Nanoseconds() > 0 {
			// Requeue request to add creation delay
			log.Info("Requeuing request to add monitor " + monitorName + " for " + fmt.Sprintf("%+v", config.GetControllerConfig().CreationDelay) + " seconds")
			return reconcile.Result{RequeueAfter: delay}, nil
		}
		err = r.handleCreate(req, instance, monitorName, monitorService)
	}
	return reconcile.Result{RequeueAfter: config.ReconciliationRequeueTime}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *EndpointMonitorReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: maxConcurrentReconciles,
		}).
		For(&endpointmonitorv1alpha1.EndpointMonitor{}).
		Complete(r)
}

func (r *EndpointMonitorReconciler) GetMonitorOfType(spec endpointmonitorv1alpha1.EndpointMonitorSpec) *monitors.MonitorServiceProxy {
	if len(r.MonitorServices) == 0 {
		panic("No monitor services found")
	}
	if spec.PingdomTransactionConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypePingdomTransaction)
	}
	if spec.PingdomConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypePingdom)
	}
	if spec.UptimeRobotConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypeUptimeRobot)
	}
	if spec.StatusCakeConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypeStatusCake)
	}
	if spec.UptimeConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypeUptime)
	}
	if spec.UpdownConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypeUpdown)
	}
	if spec.AppInsightsConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypeAppInsights)
	}
	if spec.GCloudConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypeGCloud)
	}
	if spec.GrafanaConfig != nil {
		return r.GetMonitorServiceOfType(monitors.TypeGrafana)
	}
	// If none of the above, return the first monitor service
	return r.MonitorServices[0]
}

func (r *EndpointMonitorReconciler) GetMonitorServiceOfType(monitorType string) *monitors.MonitorServiceProxy {
	for _, monitorService := range r.MonitorServices {
		if monitorService.GetType() == monitorType {
			return monitorService
		}
	}
	log.Info("Error could not find monitor service " + monitorType + " in list of monitor services")
	return nil
}
