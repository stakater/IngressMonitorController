package endpointmonitor

import (
	"context"
	log "github.com/sirupsen/logrus"
	"strconv"
	"testing"
	"time"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/monitors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakekubeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	testName        = "test-endpointmonitor"
	testNamespace   = "test-namespace"
	testURL         = "https://www.google.com/"
	testURLFacebook = "https://www.facebook.com/"
)

var (
	EndpointMonitorInstance = endpointmonitorv1alpha1.EndpointMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testName,
			Namespace: testNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "EndpointMonitor",
			APIVersion: "endpointmonitor.stakater.com/v1alpha1",
		},
		Spec: endpointmonitorv1alpha1.EndpointMonitorSpec{
			URL:        testURL,
			ForceHTTPS: true,
		},
	}
)

func TestEndpointMonitorReconcileCreate(t *testing.T) {
	log.Info("Testing reconcile for create")

	monitorName := testName + "-" + testNamespace
	_, r, req := setupReconcilerAndCreateResource()

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		log.Error("reconcile did not return an empty Result")
	}

	// Sleep for 5 seconds since monitor creation takes time for Updown provider
	time.Sleep(5 * time.Second)

	// Check that the monitors are created
	monitorCount := getMonitorCount(monitorName, r.monitorServices)

	if monitorCount != len(r.monitorServices) {
		t.Error("Unable to create monitors for all providers, only " + strconv.Itoa(monitorCount) + "/" + strconv.Itoa(len(r.monitorServices)) + " monitors were added")
	}

	// Cleanup
	// Ensure that all monitors are removed(Required in case of failure)
	removeAllMonitors(monitorName, r.monitorServices)
}

func TestEndpointMonitorReconcileUpdate(t *testing.T) {
	log.Info("Testing reconcile for update")

	monitorName := testName + "-" + testNamespace
	cl, r, req := setupReconcilerAndCreateResource()

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		log.Error("reconcile did not return an empty Result")
	}

	// Sleep for 5 seconds since monitor creation takes time for Updown provider
	time.Sleep(5 * time.Second)

	// Check that the monitors are created
	monitorCount := getMonitorCount(monitorName, r.monitorServices)

	if monitorCount != len(r.monitorServices) {
		t.Error("Unable to create monitors for all providers, only " + strconv.Itoa(monitorCount) + "/" + strconv.Itoa(len(r.monitorServices)) + " monitors were added")
	}

	endpointMonitorObject := &endpointmonitorv1alpha1.EndpointMonitor{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: testName, Namespace: testNamespace}, endpointMonitorObject)
	if err != nil {
		t.Fatalf("Get EndpointMonitor Instance : (%v)", err)
	}

	// Update URL and re-run reconcile for update
	endpointMonitorObject.Spec.URL = testURLFacebook
	err = cl.Update(context.TODO(), endpointMonitorObject)
	if err != nil {
		t.Error(err, "Could not update EndpointMonitor CR")
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		log.Error("reconcile did not return an empty Result")
	}

	// Sleep for 5 seconds since update takes time for Updown provider
	time.Sleep(5 * time.Second)

	monitorCount = 0
	for index := 0; index < len(r.monitorServices); index++ {
		monitor, err := findMonitorByName(r.monitorServices[index], monitorName)
		if monitor != nil && err == nil && monitor.URL == testURLFacebook {
			log.Info("Found Updated Monitor for Provider: " + r.monitorServices[index].GetType())
			monitorCount++
		}
	}

	if monitorCount != len(r.monitorServices) {
		t.Error("Unable to update monitors for all providers, only " + strconv.Itoa(monitorCount) + "/" + strconv.Itoa(len(r.monitorServices)) + " monitors were updated")
	}

	// Cleanup
	// Ensure that all monitors are removed(Required in case of failure)
	removeAllMonitors(monitorName, r.monitorServices)
}

func TestEndpointMonitorReconcileDelete(t *testing.T) {
	log.Info("Testing reconcile for delete")

	monitorName := testName + "-" + testNamespace
	cl, r, req := setupReconcilerAndCreateResource()

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		log.Error("reconcile did not return an empty Result")
	}

	// Sleep for 5 seconds since monitor creation takes time for Updown provider
	time.Sleep(5 * time.Second)

	// Check that the monitors are created
	monitorCount := getMonitorCount(monitorName, r.monitorServices)

	if monitorCount != len(r.monitorServices) {
		t.Error("Unable to create monitors for all providers, only " + strconv.Itoa(monitorCount) + "/" + strconv.Itoa(len(r.monitorServices)) + " monitors were added")
	}

	endpointMonitorObject := &endpointmonitorv1alpha1.EndpointMonitor{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: testName, Namespace: testNamespace}, endpointMonitorObject)
	if err != nil {
		t.Fatalf("Get EndpointMonitor Instance : (%v)", err)
	}

	// Delete CR to test deletion
	err = cl.Delete(context.TODO(), endpointMonitorObject)
	if err != nil {
		t.Error(err, "Could not delete EndpointMonitor CR")
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		log.Error("reconcile did not return an empty Result")
	}

	monitorCount = 0
	for index := 0; index < len(r.monitorServices); index++ {
		monitor, err := findMonitorByName(r.monitorServices[index], monitorName)
		if err != nil {
			t.Error(err, "Could not findMonitorByName")
		}
		if monitor == nil {
			monitorCount++
		}
	}

	if monitorCount != len(r.monitorServices) {
		t.Error("Unable to delete monitors for all providers, only " + strconv.Itoa(monitorCount) + "/" + strconv.Itoa(len(r.monitorServices)) + " monitors were deleted")
	}

	// Cleanup
	// Ensure that all monitors are removed(Required in case of failure)
	removeAllMonitors(monitorName, r.monitorServices)
}

func setupReconcilerAndCreateResource() (client.Client, *ReconcileEndpointMonitor, reconcile.Request) {
	controllerConfig := config.GetControllerConfigTest()
	monitorServices := monitors.SetupMonitorServicesForProvidersTest(controllerConfig.Providers)

	endpointMonitor := &EndpointMonitorInstance

	// Objects to track in the fake client.
	objs := []runtime.Object{
		endpointMonitor,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(endpointmonitorv1alpha1.SchemeGroupVersion, endpointMonitor)
	cl := fakekubeclient.NewFakeClient(objs...)
	r := &ReconcileEndpointMonitor{client: cl, scheme: s, monitorServices: monitorServices}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      testName,
			Namespace: testNamespace,
		},
	}

	return cl, r, req
}

func getMonitorCount(monitorName string, monitorServices []monitors.MonitorServiceProxy) int {
	monitorCount := 0
	for index := 0; index < len(monitorServices); index++ {
		monitor, err := findMonitorByName(monitorServices[index], monitorName)
		if err != nil {
			log.Error(err, "Could not findMonitorByName")
		}
		if monitor != nil {
			log.Info("Found Monitor for Provider: " + monitorServices[index].GetType())
			monitorCount++
		}
	}
	return monitorCount
}

func removeAllMonitors(monitorName string, monitorServices []monitors.MonitorServiceProxy) {
	for index := 0; index < len(monitorServices); index++ {
		removeMonitorIfExists(monitorServices[index], monitorName)
	}
}
