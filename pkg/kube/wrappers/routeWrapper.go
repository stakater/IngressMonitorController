package wrappers

import (
	"context"
	"fmt"
	"net/url"
	"path"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("route-wrapper")

type RouteWrapper struct {
	Route  *routev1.Route
	Client client.Client
}

func NewRouteWrapper(route *routev1.Route, client client.Client) *RouteWrapper {
	return &RouteWrapper{
		Route:  route,
		Client: client,
	}
}

func (rw *RouteWrapper) supportsTLS() bool {
	return rw.Route.Spec.TLS != nil
}

func (rw *RouteWrapper) tryGetTLSHost(forceHttps bool) (string, bool) {
	if rw.supportsTLS() {
		return "https://" + rw.Route.Spec.Host, true
	}

	if forceHttps {
		return "https://" + rw.Route.Spec.Host, true
	}

	return "", false
}

func (rw *RouteWrapper) getHost() string {
	return "http://" + rw.Route.Spec.Host
}

func (rw *RouteWrapper) getRouteSubPath() string {
	return rw.Route.Spec.EscapedPath()
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
	err := rw.Client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: rw.Route.Namespace}, service)
	if err != nil {
		log.Info(fmt.Sprintf("Get service from kubernetes cluster error:%v", err))
		return "", false
	}

	labels := labels.Set(service.Spec.Selector)

	podList := &corev1.PodList{}
	listOps := &client.ListOptions{
		Namespace:     rw.Route.Namespace,
		LabelSelector: labels.AsSelector(),
	}
	err = rw.Client.List(context.TODO(), podList, listOps)
	if err != nil {
		log.Info(fmt.Sprintf("List Pods of service[%s] error:%v", service.GetName(), err))
	} else if len(podList.Items) > 0 {
		pod := podList.Items[0]
		podContainers := pod.Spec.Containers

		if len(podContainers) == 1 {
			if podContainers[0].ReadinessProbe != nil && podContainers[0].ReadinessProbe.HTTPGet != nil {
				return podContainers[0].ReadinessProbe.HTTPGet.Path, true
			}
		} else {
			log.Info(fmt.Sprintf("Pod has %d containers so skipping health endpoint", len(podContainers)))
		}
	}

	return "", false
}

func (rw *RouteWrapper) GetURL(forceHttps bool, healthEndpoint string) string {
	var URL string

	if host, exists := rw.tryGetTLSHost(forceHttps); exists { // Get TLS Host if it exists
		URL = host
	} else {
		URL = rw.getHost() // Fallback for normal Host
	}

	// Convert url to url object
	u, err := url.Parse(URL)

	if err != nil {
		log.Info(fmt.Sprintf("URL parsing error in getURL() :%v", err))
		return ""
	}

	if len(healthEndpoint) != 0 {
		u.Path = healthEndpoint
	} else {
		// Append subpath
		u.Path = path.Join(u.EscapedPath(), rw.getRouteSubPath())

		// Find pod by backtracking route -> service -> pod
		healthEndpoint, exists := rw.tryGetHealthEndpointFromRoute()

		// Health endpoint from pod successful
		if exists {
			u.Path = path.Join(u.EscapedPath(), healthEndpoint)
		}
	}
	return u.String()
}
