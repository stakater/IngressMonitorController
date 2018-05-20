package main

import (
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	if len(currentNamespace) == 0 {
		currentNamespace = v1.NamespaceAll
		log.Println("Warning: KUBERNETES_NAMESPACE is unset, will monitor ingresses in all namespaces.")
	}

	// create the in-cluster config
	clusterConfig := createInClusterConfig()

	// create the clientset
	clientset := createKubernetesClient(clusterConfig)

	// fetche and create controller config from file
	config := getControllerConfig()

	// create the monitoring controller
	controller := NewMonitorController(currentNamespace, clientset, config)

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

func getControllerConfig() Config {
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if len(configFilePath) == 0 {
		configFilePath = "config.yaml"
	}

	config := ReadConfig(configFilePath)

	return config
}
