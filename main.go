package main

import (
	log "github.com/sirupsen/logrus"
	"os"

	routeClient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"github.com/stakater/IngressMonitorController/pkg/config"
	"github.com/stakater/IngressMonitorController/pkg/controller"
	"github.com/stakater/IngressMonitorController/pkg/kube"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

func init() {
	if logLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		if level, err := log.ParseLevel(logLevel); err != nil {
			log.SetLevel(level)
		}
	}
	if logFormat, ok := os.LookupEnv("LOG_FORMAT"); ok {
		switch strings.ToLower(logFormat) {
		case "json":
			log.SetFormatter(&log.JSONFormatter{})
		default:
			log.SetFormatter(&log.TextFormatter{})
		}
	}
}

func main() {
	currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	if len(currentNamespace) == 0 {
		currentNamespace = v1.NamespaceAll
		log.Println("Warning: KUBERNETES_NAMESPACE is unset, will monitor ingresses in all namespaces.")
	}

	var kubeClient kubernetes.Interface
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = kube.GetClientOutOfCluster()
	} else {
		kubeClient = kube.GetClient()
	}

	// fetche and create controller config from file
	config := config.GetControllerConfig()

	var resource = "ingresses"
	var restClient rest.Interface
	var osClient *routeClient.RouteV1Client
	if kube.IsOpenShift(kubeClient.(*kubernetes.Clientset)) {
		resource = "routes"
		// Create an OpenShift build/v1 client.
		osClient, err = routeClient.NewForConfig(cfg)
		if err != nil {
			log.Panic(err.Error())
		}
		restClient = osClient.RESTClient()
	} else {
		restClient = kubeClient.ExtensionsV1beta1().RESTClient()
	}

	// create the monitoring controller
	controller := controller.NewMonitorController(currentNamespace, kubeClient, config, resource, restClient)

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
