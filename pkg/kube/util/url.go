package util

import (
	"context"
	"errors"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stakater/IngressMonitorController/pkg/kube"
	"github.com/stakater/IngressMonitorController/pkg/kube/wrappers"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/api/v1alpha1"
)

var log = logf.Log.WithName("config")

func GetMonitorURL(client client.Client, ingressMonitor *endpointmonitorv1alpha1.EndpointMonitor) (string, error) {
	if len(ingressMonitor.Spec.URL) == 0 {
		return discoverURLFromRefs(client, ingressMonitor)
	}
	if ingressMonitor.Spec.URLFrom != nil {
		log.V(1).Info("Both url and urlFrom fields are specified. Using url over urlFrom")
	}
	if len(ingressMonitor.Spec.HealthEndpoint) > 0 {
		log.V(1).Info("Ignoring HealthEndpoint since url field is specified")
	}
	return ingressMonitor.Spec.URL, nil
}

func discoverURLFromIngressRef(client client.Client, ingressRef *endpointmonitorv1alpha1.IngressURLSource, namespace string, forceHttps bool, healthEndpoint string) (string, error) {
	ingressObject := &v1.Ingress{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: ingressRef.Name, Namespace: namespace}, ingressObject)
	if err != nil {
		log.V(1).Info("Ingress not found with name " + ingressRef.Name)
		return "", err
	}

	ingressWrapper := wrappers.NewIngressWrapper(ingressObject, client)
	return ingressWrapper.GetURL(forceHttps, healthEndpoint), nil
}

func discoverURLFromRouteRef(client client.Client, routeRef *endpointmonitorv1alpha1.RouteURLSource, namespace string, forceHttps bool, healthEndpoint string) (string, error) {
	routeObject := &routev1.Route{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: routeRef.Name, Namespace: namespace}, routeObject)
	if err != nil {
		log.V(1).Info("Route not found with name " + routeRef.Name)
		return "", err
	}

	routeWrapper := wrappers.NewRouteWrapper(routeObject, client)
	return routeWrapper.GetURL(forceHttps, healthEndpoint), nil
}

func discoverURLFromRefs(client client.Client, ingressMonitor *endpointmonitorv1alpha1.EndpointMonitor) (string, error) {
	urlFrom := ingressMonitor.Spec.URLFrom
	if urlFrom == nil {
		log.V(1).Info("No URL sources set for ingressMonitor: " + ingressMonitor.Name)
		return "", errors.New("No URL sources set for ingressMonitor: " + ingressMonitor.Name)
	}

	if urlFrom.IngressRef != nil && !kube.IsOpenshift {
		return discoverURLFromIngressRef(client, urlFrom.IngressRef, ingressMonitor.Namespace, ingressMonitor.Spec.ForceHTTPS, ingressMonitor.Spec.HealthEndpoint)
	}
	if urlFrom.RouteRef != nil && kube.IsOpenshift {
		return discoverURLFromRouteRef(client, urlFrom.RouteRef, ingressMonitor.Namespace, ingressMonitor.Spec.ForceHTTPS, ingressMonitor.Spec.HealthEndpoint)
	}

	log.V(1).Info("Unsupported Ref set on ingressMonitor: " + ingressMonitor.Name)
	return "", errors.New("Unsupported Ref set on ingressMonitor: " + ingressMonitor.Name)
}
