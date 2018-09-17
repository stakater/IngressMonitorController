package kube

import (
	"encoding/json"
	"os"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"github.com/stakater/IngressMonitorController/pkg/callbacks"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not create kubernetes client: %v", err)
	}

	return clientset
}

func buildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// GetClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func GetClientOutOfCluster() kubernetes.Interface {
	config, err := buildOutOfClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)

	return clientset
}

// IsRoute returns true if given resource is a route
func IsRoute(resource interface{}) bool {
	if _, ok := resource.(*routev1.Route); ok {
		return true
	}
	return false
}

// IsOpenShift returns true if cluster is openshift based
func IsOpenShift(c *kubernetes.Clientset) bool {
	res, err := c.RESTClient().Get().AbsPath("").DoRaw()
	if err != nil {
		return false
	}

	var rp v1.RootPaths
	err = json.Unmarshal(res, &rp)
	if err != nil {
		return false
	}
	for _, p := range rp.Paths {
		if p == "/oapi" {
			return true
		}
	}
	return false
}

// GetResourceActionFuncs provides the resource actions for ingress and routes
func GetResourceActionFuncs(resource interface{}) callbacks.ResourceActionFuncs {
	if IsRoute(resource) {
		return callbacks.ResourceActionFuncs{
			AnnotationFunc: callbacks.GetRouteAnnotation,
			NameFunc:       callbacks.GetRouteName,
			NamespaceFunc:  callbacks.GetRouteNamespace,
		}
	}

	return callbacks.ResourceActionFuncs{
		AnnotationFunc: callbacks.GetIngressAnnotation,
		NameFunc:       callbacks.GetIngressName,
		NamespaceFunc:  callbacks.GetIngressNamespace,
	}
}
