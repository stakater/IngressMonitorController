package endpointmonitor

import (
	"testing"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/endpointmonitor/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekubeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/apimachinery/pkg/types"
)

const (
	testName      = "test-endpointmonitor"
	testNamespace = "test-namespace"
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
			URL:        "https://www.google.com",
			ForceHTTPS: true,
		},
	}
)

func TestEndpointMonitorCreate(t *testing.T) {
	endpointMonitor := &EndpointMonitorInstance

	// Objects to track in the fake client.
	objs := []runtime.Object{
		endpointMonitor,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(endpointmonitorv1alpha1.SchemeGroupVersion, endpointMonitor)
	cl := fakekubeclient.NewFakeClient(objs...)
	r := &ReconcileEndpointMonitor{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      testName,
			Namespace: testNamespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		t.Error("reconcile did not return an empty Result")
	}

	// Check that the monitor is created

	// Check that the status of the CR is updated
}
