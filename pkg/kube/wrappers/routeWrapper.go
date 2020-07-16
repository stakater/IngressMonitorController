package wrappers

import (
	"context"
	"net/url"
	"path"

	routev1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"
	"github.com/stakater/IngressMonitorController/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RouteWrapper struct {
	Route     *routev1.Route
	Namespace string
	client    client.Client
}

func NewRouteWrapper(route *routev1.Route, namespace string, client client.Client) *RouteWrapper {
	return &RouteWrapper{
		Route:     route,
		Namespace: namespace,
		client:    client,
	}
}

func (rw *RouteWrapper) supportsTLS() bool {
	if rw.Route.Spec.TLS != nil {
		return true
	}
	return false
}

func (rw *RouteWrapper) tryGetTLSHost() (string, bool) {
	if rw.supportsTLS() {
		return "https://" + rw.Route.Spec.Host, true
	}

	annotations := rw.Route.GetAnnotations()
	if value, ok := annotations[constants.ForceHTTPSAnnotation]; ok {
		if value == "true" {
			// Annotation exists and is enabled
			return "https://" + rw.Route.Spec.Host, true
		}
	}

	return "", false
}

func (rw *RouteWrapper) getHost() string {
	return "http://" + rw.Route.Spec.Host
}

func (rw *RouteWrapper) getRoutePort() string {
	if rw.Route.Spec.Port != nil && rw.Route.Spec.Port.TargetPort.String() != "" {
		return rw.Route.Spec.Port.TargetPort.String()
	}
	return ""
}

func (rw *RouteWrapper) getRouteSubPath() string {
	return rw.Route.Spec.Path
}

func (rw *RouteWrapper) hasService() (string, bool) {
	if rw.Route.Spec.To.Name != "" {
		return rw.Route.Spec.To.Name, true
	}
	return "", false
}

func (rw *RouteWrapper) tryGetHealthEndpointFromRoute() (string, bool) {
	serviceName, exists := rw.hasService()
	if !exists {
		return "", false
	}

	service := &corev1.Service{}
	err := rw.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: rw.Route.Namespace}, service)
	if err != nil {
		log.Printf("Get service from kubernetes cluster error:%v", err)
		return "", false
	}

	labels := labels.Set(service.Spec.Selector)

	podList := &corev1.PodList{}
	listOps := &client.ListOptions{
		Namespace:     rw.Route.Namespace,
		LabelSelector: labels.AsSelector(),
	}
	err = rw.client.List(context.TODO(), podList, listOps)
	if err != nil {
		log.Printf("List Pods of service[%s] error:%v", service.GetName(), err)
	} else if len(podList.Items) > 0 {
		pod := podList.Items[0]
		podContainers := pod.Spec.Containers

		if len(podContainers) == 1 {
			if podContainers[0].ReadinessProbe != nil && podContainers[0].ReadinessProbe.HTTPGet != nil {
				return podContainers[0].ReadinessProbe.HTTPGet.Path, true
			}
		} else {
			log.Printf("Pod has %d containers so skipping health endpoint", len(podContainers))
		}
	}

	return "", false
}

func (rw *RouteWrapper) GetURL() string {
	var URL string

	if host, exists := rw.tryGetTLSHost(); exists { // Get TLS Host if it exists
		URL = host
	} else {
		URL = rw.getHost() // Fallback for normal Host
	}

	// Convert url to url object
	u, err := url.Parse(URL)

	if err != nil {
		log.Printf("URL parsing error in getURL() :%v", err)
		return ""
	}

	annotations := rw.Route.GetAnnotations()

	if value, ok := annotations[constants.OverridePathAnnotation]; ok {
		u.Path = value
	} else {
		// Append subpath
		u.Path = path.Join(u.Path, rw.getRouteSubPath())

		// Find pod by backtracking route -> service -> pod
		healthEndpoint, exists := rw.tryGetHealthEndpointFromRoute()

		// Health endpoint from pod successful
		if exists {
			u.Path = path.Join(u.Path, healthEndpoint)
		}
	}
	return u.String()
}
