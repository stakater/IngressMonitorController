package kube

import (
	"context"
	"encoding/json"
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	IsOpenshift = isOpenshift()
	log         = logf.Log.WithName("kube")
)

func getConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	// If file exists so use that config settings
	if _, err := os.Stat(kubeconfigPath); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		// Use Incluster Configuration
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

func CreateSingleClient() (client.Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		return nil, err
	}

	k8sClient, err := client.New(config, client.Options{Scheme: s})
	if err != nil {
		return nil, err
	}
	return k8sClient, err
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
		log.Error(err, "Unable to create Kubernetes client")
		os.Exit(1)
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

func GetCurrentKubernetesNamespace() string {
	// Read the namespace from the file
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil || len(namespace) == 0 {
		log.Error(err, "Failed to read namespace from file")
		return ""
	}
	return string(namespace)
}
