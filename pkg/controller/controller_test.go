package controller

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/kube"
	"github.com/stakater/IngressMonitorController/pkg/monitors/uptimerobot"
	"github.com/stakater/IngressMonitorController/pkg/util"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ingressNamePrefix = "ingress-imc-"
	podNamePrefix     = "pod-imc-"
	serviceNamePrefix = "service-imc-"
	letters           = []rune("abcdefghijklmnopqrstuvwxyz")
)

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestAddIngressWithNoAnnotationShouldNotCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	time.Sleep(5 * time.Second)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestAddIngressWithCorrectMonitorTemplate(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)
	monitorTemplate := "{{.IngressName}}-{{.Namespace}}-hello"

	controller := getControllerWithNamespace(namespace, true)

	controller.config.MonitorNameTemplate = monitorTemplate

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	time.Sleep(5 * time.Second)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	nameFormat, _ := util.GetNameTemplateFormat(monitorTemplate)
	monitorName := fmt.Sprintf(nameFormat, ingressName, namespace)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestInvalidMonitorTemplateShouldPanic(t *testing.T) {
	util.AssertPanic(t, func() {
		// Invalid monitor template
		monitorTemplate := "{.IngressName}}-{{.Namespace}"

		_, _ = util.GetNameTemplateFormat(monitorTemplate)

	})
}

func TestAddIngressWithAnnotationEnabledShouldCreateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(5 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestAddIngressWithAnnotationDisabledShouldNotCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, false)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestUpdateIngressWithAnnotationDisabledShouldNotCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress = addMonitorAnnotationToIngress(ingress, false)

	ingress, err = controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	monitorName := ingressName + "-" + namespace

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestAddIngressWithNameAnnotationShouldCreateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	monitorName := "monitor-friendly-name"
	ingress = addMonitorNameAnnotationToIngress(ingress, monitorName)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(5 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestUpdateIngressNameAnnotationShouldUpdateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := "name-annotation-ingress"

	controller := getControllerWithNamespace(namespace, true)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	monitorName := "monitor-friendly-name"
	ingress = addMonitorNameAnnotationToIngress(ingress, monitorName)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	updatedMonitorName := "monitor-friendly-name-updated"
	ingress = addMonitorNameAnnotationToIngress(ingress, updatedMonitorName)
	result, err = controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
	// Monitor with updated name should exist
	checkMonitorWithName(t, updatedMonitorName, true)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(5 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
	checkMonitorWithName(t, updatedMonitorName, false)
}

func TestUpdateIngressWithAnnotationEnabledShouldCreateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	ingress, err = controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	monitorName := ingressName + "-" + namespace

	time.Sleep(3 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestUpdateIngressWithAnnotationFromEnabledToDisabledShouldDeleteMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	ingress, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

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

	ingress = updateMonitorAnnotationInIngress(ingress, false)

	ingress, err = controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})
}

func TestUpdateIngressWithNewURLShouldUpdateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	newUrl := "facebook.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	ingress, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

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

	ingress, err = controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
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

	if monitor.URL != "http://"+newUrl {
		t.Error("Monitor did not update")
	}

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestUpdateIngressWithEnabledAnnotationShouldCreateMonitorAndDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, true)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, false)

	ingress, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", ingress.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress = updateMonitorAnnotationInIngress(ingress, true)

	ingress, err = controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	if err != nil {
		panic(err)
	}
	log.Printf("Updated ingress %q.\n", ingress.GetObjectMeta().GetName())

	monitorName := ingressName + "-" + namespace

	time.Sleep(3 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(3 * time.Second)

	// Should not exist
	checkMonitorWithName(t, monitorName, false)
}

func TestAddIngressWithAnnotationEnabledButDisableDeletionShouldCreateMonitorAndNotDelete(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, false)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	time.Sleep(5 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	// Delete the temporary monitor manually
	deleteMonitorWithName(t, monitorName)
}

func TestAddIngressWithAnnotationAssociatedWithServiceAndHasPodShouldCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)
	podName := podNamePrefix + randSeq(5)
	serviceName := serviceNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, false)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	pod := createPodObject(podName, namespace)

	pod = addReadinessProbeToPod(pod, "/health", 80)

	service := createServiceObject(serviceName, podName, namespace)

	if _, err := controller.kubeClient.Core().Pods(namespace).Create(pod); err != nil {
		panic(err)
	}

	if _, err := controller.kubeClient.Core().Services(namespace).Create(service); err != nil {
		panic(err)
	}

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	ingress = addServiceToIngress(ingress, serviceName, 80)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	monitorService := getMonitorService()

	monitor, err := monitorService.GetByName(monitorName)

	if err != nil {
		t.Error("An error occured while getting monitor")
	}

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	controller.kubeClient.Core().Pods(namespace).Delete(podName, &meta_v1.DeleteOptions{})

	controller.kubeClient.Core().Services(namespace).Delete(serviceName, &meta_v1.DeleteOptions{})

	time.Sleep(15 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	// Delete the temporary monitor manually
	deleteMonitorWithName(t, monitorName)

	if monitor.URL != "http://google.com/health" {
		t.Error("Monitor must have /health appended to the url")
	}
}

func TestAddIngressWithAnnotationAssociatedWithServiceAndHasPodButNoProbesShouldCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)
	podName := podNamePrefix + randSeq(5)
	serviceName := serviceNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, false)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	pod := createPodObject(podName, namespace)

	service := createServiceObject(serviceName, podName, namespace)

	if _, err := controller.kubeClient.Core().Pods(namespace).Create(pod); err != nil {
		panic(err)
	}

	if _, err := controller.kubeClient.Core().Services(namespace).Create(service); err != nil {
		panic(err)
	}

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	ingress = addServiceToIngress(ingress, serviceName, 80)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	monitorService := getMonitorService()

	monitor, err := monitorService.GetByName(monitorName)

	if err != nil {
		t.Error("An error occured while getting monitor")
	}

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	controller.kubeClient.Core().Pods(namespace).Delete(podName, &meta_v1.DeleteOptions{})

	controller.kubeClient.Core().Services(namespace).Delete(serviceName, &meta_v1.DeleteOptions{})

	time.Sleep(15 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	// Delete the temporary monitor manually
	deleteMonitorWithName(t, monitorName)

	if monitor.URL != "http://google.com" {
		t.Error("Monitor must not have /health appended to the url")
	}
}

func TestAddIngressWithHealthAnnotationAssociatedWithServiceAndHasPodShouldCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)
	podName := podNamePrefix + randSeq(5)
	serviceName := serviceNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, false)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	pod := createPodObject(podName, namespace)

	service := createServiceObject(serviceName, podName, namespace)

	if _, err := controller.kubeClient.Core().Pods(namespace).Create(pod); err != nil {
		panic(err)
	}

	if _, err := controller.kubeClient.Core().Services(namespace).Create(service); err != nil {
		panic(err)
	}

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	ingress = addHealthAnnotationToIngress(ingress, "/hello")

	ingress = addServiceToIngress(ingress, serviceName, 80)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	monitorService := getMonitorService()

	monitor, err := monitorService.GetByName(monitorName)

	if err != nil {
		t.Error("An error occured while getting monitor")
	}

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	controller.kubeClient.Core().Pods(namespace).Delete(podName, &meta_v1.DeleteOptions{})

	controller.kubeClient.Core().Services(namespace).Delete(serviceName, &meta_v1.DeleteOptions{})

	time.Sleep(15 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	// Delete the temporary monitor manually
	deleteMonitorWithName(t, monitorName)

	if monitor.URL != "http://google.com/hello" {
		t.Error("Monitor must have /health appended to the url")
	}
}

func TestAddIngressWithAnnotationAssociatedWithServiceAndHasNoPodShouldCreateMonitor(t *testing.T) {
	namespace := "test"
	url := "google.com"
	ingressName := ingressNamePrefix + randSeq(5)
	podName := podNamePrefix + randSeq(5)
	serviceName := serviceNamePrefix + randSeq(5)

	controller := getControllerWithNamespace(namespace, false)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	service := createServiceObject(serviceName, podName, namespace)

	if _, err := controller.kubeClient.Core().Services(namespace).Create(service); err != nil {
		panic(err)
	}

	ingress := util.CreateIngressObject(ingressName, namespace, url)

	ingress = addMonitorAnnotationToIngress(ingress, true)

	ingress = addServiceToIngress(ingress, serviceName, 80)

	result, err := controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)

	if err != nil {
		panic(err)
	}
	log.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())

	time.Sleep(5 * time.Second)

	monitorName := ingressName + "-" + namespace

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	monitorService := getMonitorService()

	monitor, err := monitorService.GetByName(monitorName)

	if err != nil {
		t.Error("An error occured while getting monitor")
	}

	controller.kubeClient.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, &meta_v1.DeleteOptions{})

	controller.kubeClient.Core().Services(namespace).Delete(serviceName, &meta_v1.DeleteOptions{})

	time.Sleep(15 * time.Second)

	// Should exist
	checkMonitorWithName(t, monitorName, true)

	// Delete the temporary monitor manually
	deleteMonitorWithName(t, monitorName)

	if monitor.URL != "http://google.com" {
		t.Error("Monitor should not have /health appended to the url since no pod exists")
	}
}

func addServiceToIngress(ingress *v1beta1.Ingress, serviceName string, servicePort int) *v1beta1.Ingress {
	ingress.Spec.Rules[0].HTTP = &v1beta1.HTTPIngressRuleValue{
		Paths: []v1beta1.HTTPIngressPath{
			v1beta1.HTTPIngressPath{
				Backend: v1beta1.IngressBackend{
					ServiceName: serviceName,
					ServicePort: intstr.FromInt(servicePort),
				},
			},
		},
	}

	return ingress
}

func addReadinessProbeToPod(pod *v1.Pod, path string, port int) *v1.Pod {
	pod.Spec.Containers[0].ReadinessProbe = &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path,
				Port: intstr.FromInt(port),
			},
		},
	}

	return pod
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

func getMonitorService() *uptimerobot.UpTimeMonitorService {
	config := config.GetControllerConfig()

	service := uptimerobot.UpTimeMonitorService{}
	service.Setup(config.Providers[0])

	return &service
}

func createServiceObject(serviceName string, podName string, namespace string) *v1.Service {
	service := &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"name":      serviceName,
				"namespace": namespace,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
			Selector: map[string]string{
				"name":      podName,
				"namespace": namespace,
			},
		},
	}

	return service
}

func createPodObject(podName string, namespace string) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"name":      podName,
				"namespace": namespace,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:  "test",
					Image: "tutum/hello-world",
				},
			},
		},
	}

	return pod
}

func addMonitorAnnotationToIngress(ingress *v1beta1.Ingress, annotationValue bool) *v1beta1.Ingress {
	if ingress.Annotations == nil {
		annotations := make(map[string]string)
		ingress.Annotations = annotations
	}
	ingress.Annotations["monitor.stakater.com/enabled"] = strconv.FormatBool(annotationValue)
	return ingress
}

func addMonitorNameAnnotationToIngress(ingress *v1beta1.Ingress, annotationValue string) *v1beta1.Ingress {
	if ingress.Annotations == nil {
		annotations := make(map[string]string)
		ingress.Annotations = annotations
	}
	ingress.Annotations["monitor.stakater.com/name"] = annotationValue
	return ingress
}

func addHealthAnnotationToIngress(ingress *v1beta1.Ingress, annotationValue string) *v1beta1.Ingress {
	if ingress.Annotations == nil {
		annotations := make(map[string]string)
		ingress.Annotations = annotations
	}
	ingress.Annotations["monitor.stakater.com/healthEndpoint"] = annotationValue
	return ingress
}

func updateMonitorAnnotationInIngress(ingress *v1beta1.Ingress, newValue bool) *v1beta1.Ingress {
	monitorAnnotation := "monitor.stakater.com/enabled"
	if _, ok := ingress.Annotations[monitorAnnotation]; ok {
		ingress.Annotations[monitorAnnotation] = strconv.FormatBool(newValue)
	}

	return ingress
}

func removeMonitorAnnotationFromIngress(ingress *v1beta1.Ingress) *v1beta1.Ingress {
	monitorAnnotation := "monitor.stakater.com/enabled"
	if _, ok := ingress.Annotations[monitorAnnotation]; ok {
		delete(ingress.Annotations, monitorAnnotation)
	}

	return ingress
}

func getControllerWithNamespace(namespace string, enableDeletion bool) *MonitorController {
	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = kube.GetClientOutOfCluster()
	} else {
		kubeClient = kube.GetClient()
	}

	// fetche and create controller config from file
	c := config.GetControllerConfig()

	// Only run controller with the 1st provider for tests
	c.Providers = []config.Provider{c.Providers[0]}

	c.EnableMonitorDeletion = enableDeletion

	// create the monitoring controller
	controller := NewMonitorController(namespace, kubeClient, c)

	return controller
}
