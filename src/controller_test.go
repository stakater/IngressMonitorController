package main

import (
	"testing"
	"time"

	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddIngressWithNoAnnotationShouldNotCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testIngress"

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := createIngress(ingressName, namespace, url, false)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestAddIngressWithAnnotationEnabledShouldCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testIngress"

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := createIngress(ingressName, namespace, url, true)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(5 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func checkMonitorWithName(t *testing.T, monitorName string, shouldExist bool) {
	service := getMonitorService()

	monitor, err := service.GetByName(monitorName)

	if err != nil {
		t.Error("An error occured while getting monitor")
	}

	if shouldExist {
		if monitor == nil {
			t.Error("Monitor does not exist but should have existed")
		}
	} else {
		t.Error("Monitor exists but shouldn't have existed")
	}

}

func getMonitorService() *UpTimeMonitorService {
	config := getControllerConfig()

	service := UpTimeMonitorService{}
	apiKey := config.Providers[0].ApiKey
	alertContacts := config.Providers[0].AlertContacts
	url := config.Providers[0].ApiURL
	service.Setup(apiKey, url, alertContacts)

	return &service
}

func createIngress(ingressName string, namespace string, url string, withAnnotation bool) *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      ingressName,
			Namespace: namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: url,
				},
			},
		},
	}

	if withAnnotation {
		annotations := make(map[string]string)
		annotations["monitor.stakater.com/enabled"] = "true"
		ingress.Annotations = annotations
	}

	return ingress
}

func getControllerWithNamespace(namespace string, enableDeletion bool) *MonitorController {
	// create the in-cluster config
	clusterConfig := createInClusterConfig()

	// create the clientset
	clientset := createKubernetesClient(clusterConfig)

	// fetche and create controller config from file
	config := getControllerConfig()

	config.EnableMonitorDeletion = enableDeletion

	// create the monitoring controller
	controller := NewMonitorController(namespace, clientset, config)

	return controller
}
