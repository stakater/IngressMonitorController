package kube

import (
	"context"
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"

	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	IsOpenshift = isOpenshift()
)

func getConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	//If file exists so use that config settings
	if _, err := os.Stat(kubeconfigPath); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		//Use Incluster Configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() (*kubernetes.Clientset, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// IsRoute returns true if given resource is a route
func IsRoute(resource interface{}) bool {
	if _, ok := resource.(*routev1.Route); ok {
		return true
	}
	return false
}

// IsOpenShift returns true if cluster is openshift based
func isOpenshift() bool {
	kubeClient, err := GetClient()
	if err != nil {
		log.Fatalf("Unable to create Kubernetes client error = %v", err)
	}

	res, err := kubeClient.RESTClient().Get().AbsPath("").DoRaw(context.TODO())
	if err != nil {
		log.Info("Failed to determine Environment, will try kubernetes")
		return false
	}

	var rp v1.RootPaths
	err = json.Unmarshal(res, &rp)
	if err != nil {
		log.Info("Failed to determine Environment, will try kubernetes")
		return false
	}
	for _, p := range rp.Paths {
		if p == "/apis/route.openshift.io" {
			log.Info("Environment is OpenShift")
			return true
		}
	}
	log.Info("Environment is Vanilla Kubernetes")
	return false
}
