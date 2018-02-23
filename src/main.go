package main

import (
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
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		panic(err.Error())
	}

	config := Config{providers: []Provider{Provider{name: "UptimeRobot", apiKey: "u544483-b3647f3e973b66417071a555", apiURL: "https://api.uptimerobot.com/v2/", alertContacts: "0544483_0_0-2628365_0_0-2633263_0_0"}}, enableMonitorDeletion: true}

	controller := NewMonitorController(currentNamespace, clientset, config)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
