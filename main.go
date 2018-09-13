package main

import (
	"log"
	"os"

	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/controller"
	"github.com/stakater/IngressMonitorController/pkg/kube"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	if len(currentNamespace) == 0 {
		currentNamespace = v1.NamespaceAll
		log.Println("Warning: KUBERNETES_NAMESPACE is unset, will monitor ingresses in all namespaces.")
	}

	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = kube.GetClientOutOfCluster()
	} else {
		kubeClient = kube.GetClient()
	}

	// fetche and create controller config from file
	config := config.GetControllerConfig()

	var resource = "ingresses"
	if kube.IsOpenShift(kubeClient.(*kubernetes.Clientset)) {
		resource = "routes"
	}

	// create the monitoring controller
	controller := controller.NewMonitorController(currentNamespace, kubeClient, config, resource)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}

func createInClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Panic(err.Error())
	}
	return config
}

func createKubernetesClient(config *rest.Config) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}
	return clientset
}
