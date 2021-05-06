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
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/monitors"
	"github.com/stakater/IngressMonitorController/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
)

// EndpointMonitorReconciler reconciles a EndpointMonitor object
type EndpointMonitorReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	MonitorServices []monitors.MonitorServiceProxy
}

const (
	defaultRequeueTime = 60 * time.Second
)

//+kubebuilder:rbac:groups=endpointmonitor.stakater.com,resources=endpointmonitors,verbs=get;list;watch
//+kubebuilder:rbac:groups=endpointmonitor.stakater.com,resources=endpointmonitors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=endpointmonitor.stakater.com,resources=endpointmonitors/finalizers,verbs=update
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch
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

	for index := 0; index < len(r.MonitorServices); index++ {
		monitor := findMonitorByName(r.MonitorServices[index], monitorName)
		if monitor != nil {
			// Monitor already exists, update if required
			err = r.handleUpdate(req, instance, *monitor, r.MonitorServices[index])
		} else {
			// Monitor doesn't exist, create monitor
			if delay.Nanoseconds() > 0 {
				// Requeue request to add creation delay
				log.Info("Requeuing request to add monitor " + monitorName + " for " + fmt.Sprintf("%+v", config.GetControllerConfig().CreationDelay) + " seconds")
				return reconcile.Result{RequeueAfter: delay}, nil
			}
			err = r.handleCreate(req, instance, monitorName, r.MonitorServices[index])
		}
	}

	return reconcile.Result{RequeueAfter: defaultRequeueTime}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *EndpointMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&endpointmonitorv1alpha1.EndpointMonitor{}).
		Complete(r)
}
