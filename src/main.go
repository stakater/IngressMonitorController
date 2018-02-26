package main

import (
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	if len(currentNamespace) == 0 {
		log.Fatal("Could not find the current namespace")
	}

	// creates the in-cluster config
	clusterConfig := createInClusterConfig()

	// creates the clientset
	clientset := createKubernetesClient(clusterConfig)

	config := getControllerConfig()

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
