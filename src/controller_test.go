package main

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddIngressWithNoAnnotationShouldNotCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	time.Sleep(5 * time.Second)

	ingress := createIngressObject(ingressName, namespace, url)

	result, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestAddIngressWithCorrectMonitorTemplate(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"
	monitorTemplate := "{{.IngressName}}-{{.Namespace}}-hello"

	controller := getControllerWithNamespace(namespace, true)

	controller.config.MonitorNameTemplate = monitorTemplate

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	time.Sleep(5 * time.Second)

	ingress := createIngressObject(ingressName, namespace, url)

	result, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	nameFormat, _ := getNameTemplateFormat(monitorTemplate)
	monitorName := fmt.Sprintf(nameFormat, ingressName, namespace)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestInvalidMonitorTemplateShouldPanic(t *testing.T) {
	assertPanic(t, func() {
		// Invalid monitor template
		monitorTemplate := "{.IngressName}}-{{.Namespace}"

		_, _ = getNameTemplateFormat(monitorTemplate)

	})
}

func TestAddIngressWithAnnotationEnabledShouldCreateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotation(ingress, true)

	result, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(5 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestAddIngressWithAnnotationDisabledShouldNotCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotation(ingress, false)

	result, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestUpdateIngressWithAnnotationDisabledShouldNotCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress = addMonitorAnnotation(ingress, false)

	ingress, err = controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	monitorName := ingressName + "-" + namespace

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestUpdateIngressWithAnnotationEnabledShouldCreateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress = addMonitorAnnotation(ingress, true)

	ingress, err = controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	monitorName := ingressName + "-" + namespace

	time.Sleep(3 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestUpdateIngressWithAnnotationFromEnabledToDisabledShouldDeleteMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotation(ingress, true)

	ingress, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	ingress = updateMonitorAnnotation(ingress, false)

	ingress, err = controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestUpdateIngressWithNewURLShouldUpdateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	newUrl := "facebook.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotation(ingress, true)

	ingress, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	// Update url
	ingress.Spec.Rules[0].Host = newUrl

	ingress, err = controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(3 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	service := getMonitorService()
	monitor, err := service.GetByName(monitorName)

	if err != nil {
		t.Error("Cannot Fetch monitor")
	}

	if monitor == nil {
		t.Error("Monitor with name" + monitorName + " does not exist")
	}

	if monitor.url != "http://"+newUrl {
		t.Error("Monitor did not update")
	}

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestUpdateIngressWithEnabledAnnotationShouldCreateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, true)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotation(ingress, false)

	ingress, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress = updateMonitorAnnotation(ingress, true)

	ingress, err = controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	monitorName := ingressName + "-" + namespace

	time.Sleep(3 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestAddIngressWithAnnotationEnabledButDisableDeletionShouldCreateMonitorAndNotDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "testingingress"

	controller := getControllerWithNamespace(namespace, false)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := createIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotation(ingress, true)

	result, err := controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.clientset.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(5 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	// Delete the temporary monitor manually
	deleteMonitorWithName(t, monitorName)
}

func deleteMonitorWithName(t *testing.T, monitorName string) {
	service := getMonitorService()

	monitor, err := service.GetByName(monitorName)

	if err != nil {
		t.Error("An error occured while getting monitor")
	}

	if monitor == nil {
		t.Error("Monitor does not exist but should have existed")
	} else {
		service.Remove(*monitor)
	}
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
		if monitor != nil {
			t.Error("Monitor exists but shouldn't have existed")
		}
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

func createIngressObject(ingressName string, namespace string, url string) *v1beta1.Ingress {
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

	return ingress
}

func addMonitorAnnotation(ingress *v1beta1.Ingress, annotationValue bool) *v1beta1.Ingress {
	annotations := make(map[string]string)
	annotations["monitor.stakater.com/enabled"] = strconv.FormatBool(annotationValue)
	ingress.Annotations = annotations

	return ingress
}

func updateMonitorAnnotation(ingress *v1beta1.Ingress, newValue bool) *v1beta1.Ingress {
	monitorAnnotation := "monitor.stakater.com/enabled"
	if _, ok := ingress.Annotations[monitorAnnotation]; ok {
		ingress.Annotations[monitorAnnotation] = strconv.FormatBool(newValue)
	}

	return ingress
}

func removeMonitorAnnotation(ingress *v1beta1.Ingress) *v1beta1.Ingress {
	monitorAnnotation := "monitor.stakater.com/enabled"
	if _, ok := ingress.Annotations[monitorAnnotation]; ok {
		delete(ingress.Annotations, monitorAnnotation)
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
