package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	// if len(currentNamespace) == 0 {
	// 	glog.Fatalf("Could not find the current namespace")
	// }

	currentNamespace := "test2"

	// creates the in-cluster config
	clusterConfig := createInClusterConfig()

	// creates the clientset
	clientset := createKubernetesClient(clusterConfig)

	config := getControllerConfig()

	// config := Config{Providers: []Provider{Provider{Name: "UptimeRobot", ApiKey: "u544483-b3647f3e973b66417071a555", ApiURL: "https://api.uptimerobot.com/v2/", AlertContacts: "0544483_0_0-2628365_0_0-2633263_0_0"}}, EnableMonitorDeletion: true}

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
		panic(err.Error())
	}
	return config
}

func createKubernetesClient(config *rest.Config) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
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
