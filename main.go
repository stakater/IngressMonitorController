package main

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	// if len(currentNamespace) == 0 {
	// 	glog.Fatalf("Could not find the current namespace")
	// }

	currentNamespace := "tools"

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	controller := NewMonitorController(currentNamespace, clientset)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}

	for {
		ingresses, err := clientset.ExtensionsV1beta1().Ingresses(currentNamespace).List(metav1.ListOptions{})

		if err != nil {
			panic(err.Error())
		}

		clientset.ExtensionsV1beta1().Ingresses(currentNamespace).Watch(metav1.ListOptions{})

		for index := 0; index < len(ingresses.Items); index++ {
			ingress := ingresses.Items[index]
			if len(ingress.Spec.Rules) > 0 {
				rule := ingress.Spec.Rules[0]
				fmt.Println("Ingress: " + ingress.GetName() + " Host: " + rule.Host)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
