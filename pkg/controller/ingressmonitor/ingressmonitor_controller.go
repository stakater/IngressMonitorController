package ingressmonitor

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"

	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/models"
	"github.com/stakater/IngressMonitorController/pkg/monitors"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	controllerName     = "ingressmonitor-controller"
	defaultRequeueTime = 60 * time.Second
)

// Add creates a new IngressMonitor Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	config := config.GetControllerConfig()
	return &ReconcileIngressMonitor{
		client:          mgr.GetClient(),
		scheme:          mgr.GetScheme(),
		monitorServices: monitors.SetupMonitorServicesForProviders(config.Providers),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IngressMonitor
	err = c.Watch(&source.Kind{Type: &ingressmonitorv1alpha1.IngressMonitor{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileIngressMonitor implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIngressMonitor{}

// ReconcileIngressMonitor reconciles a IngressMonitor object
type ReconcileIngressMonitor struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client          client.Client
	scheme          *runtime.Scheme
	monitorServices []monitors.MonitorServiceProxy
}

// Reconcile reads that state of the cluster for a IngressMonitor object and makes changes based on the state read
// and what is in the IngressMonitor.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIngressMonitor) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling IngressMonitor")

	// Fetch the IngressMonitor instance
	instance := &ingressmonitorv1alpha1.IngressMonitor{}

	monitorName := request.Name + "-" + request.Namespace

	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return r.handleDelete(request, instance, monitorName)
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	for index := 0; index < len(r.monitorServices); index++ {
		log.Info("DEBUG: Iterating through monitorServices ", "monitorServices[index]", r.monitorServices[index])
		monitor := findMonitorByName(r.monitorServices[index], monitorName)
		if monitor != nil {
			// Monitor already exists, update if required
			r.handleUpdate(request, instance, *monitor, r.monitorServices[index])
		} else {
			// Monitor doesn't exist, create monitor
			r.handleCreate(request, instance, monitorName, r.monitorServices[index])
		}
	}

	return reconcile.Result{}, nil
}

func findMonitorByName(monitorService monitors.MonitorServiceProxy, monitorName string) *models.Monitor {

	log.Info("DEBUG: monitorService ", "monitorService", monitorService)
	log.Info("DEBUG: monitorName ", "monitorName", monitorName)

	monitor, _ := monitorService.GetByName(monitorName)
	// Monitor Exists
	if monitor != nil {
		return monitor
	}
	return nil
}
