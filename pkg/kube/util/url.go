package util

import (
	"errors"

	routes "github.com/openshift/client-go/route/clientset/versioned"
	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetURL(kubeClient kubernetes.Interface, ingressMonitor ingressmonitorv1alpha1.IngressMonitor) (string, error) {
	if len(ingressMonitor.Spec.URL) == 0 {
		return discoverURLFromRefs(clients, ingressMonitor)
	}
	return ingressMonitor.Spec.URL, nil
}

func discoverURLFromIngressRef(kubeClient kubernetes.Interface, ingressRef *v1alpha1.IngressURLSource, namespace string) (string, error) {
	ingress, err := kubeClient.ExtensionsV1beta1().Ingresses(namespace).Get(ingressRef.Name, metav1.GetOptions{})
	if err != nil {
		logger.Warn("Ingress not found with name " + ingressRef.Name)
		return "", err
	}
	return wrappers.NewIngressWrapper(ingress).GetURL(), nil
}

func discoverURLFromRouteRef(routesClient routes.Interface, routeRef *v1alpha1.RouteURLSource, namespace string) (string, error) {
	route, err := routesClient.RouteV1().Routes(namespace).Get(routeRef.Name, metav1.GetOptions{})
	if err != nil {
		logger.Warn("Route not found with name " + routeRef.Name)
		return "", err
	}

	return wrappers.NewRouteWrapper(route).GetURL(), nil
}

func discoverURLFromRefs(clients kube.Clients, ingressMonitor ingressmonitorv1alpha1.IngressMonitor) (string, error) {
	urlFrom := ingressMonitor.Spec.URLFrom
	if urlFrom == nil {
		logger.Warn("No URL sources set for ingressMonitor: " + ingressMonitor.Name)
		return "", errors.New("No URL sources set for ingressMonitor: " + ingressMonitor.Name)
	}

	if urlFrom.IngressRef != nil && !kube.IsOpenshift {
		return discoverURLFromIngressRef(clients.KubernetesClient, urlFrom.IngressRef, ingressMonitor.Namespace)
	}

	if urlFrom.RouteRef != nil && kube.IsOpenshift  {
		return discoverURLFromRouteRef(clients.RoutesClient, urlFrom.RouteRef, ingressMonitor.Namespace)
	}

	logger.Warn("Unsupported Ref set on ingressMonitor: " + ingressMonitor.Name)
	return "", errors.New("Unsupported Ref set on ingressMonitor: " + ingressMonitor.Name)
}