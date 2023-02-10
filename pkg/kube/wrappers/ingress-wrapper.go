package wrappers

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IngressWrapper struct {
	Ingress *v1.Ingress
	Client  client.Client
}

func NewIngressWrapper(ingress *v1.Ingress, client client.Client) *IngressWrapper {
	return &IngressWrapper{
		Ingress: ingress,
		Client:  client,
	}
}

func (iw *IngressWrapper) supportsTLS() bool {
	if iw.Ingress.Spec.TLS != nil && len(iw.Ingress.Spec.TLS) > 0 && iw.Ingress.Spec.TLS[0].Hosts != nil && len(iw.Ingress.Spec.TLS[0].Hosts) > 0 && len(iw.Ingress.Spec.TLS[0].Hosts[0]) > 0 {
		return true
	}
	return false
}

func (iw *IngressWrapper) tryGetTLSHost(forceHttps bool) (string, bool) {
	if iw.supportsTLS() {
		return "https://" + iw.Ingress.Spec.TLS[0].Hosts[0], true
	}

	if forceHttps {
		return "https://" + iw.Ingress.Spec.Rules[0].Host, true
	}

	return "", false
}

func (iw *IngressWrapper) getHost() string {
	return "http://" + iw.Ingress.Spec.Rules[0].Host
}

func (iw *IngressWrapper) rulesExist() bool {
	if iw.Ingress.Spec.Rules != nil && len(iw.Ingress.Spec.Rules) > 0 {
		return true
	}
	return false
}

func (iw *IngressWrapper) getIngressSubPath() string {
	rule := iw.Ingress.Spec.Rules[0]
	if rule.HTTP != nil {
		if rule.HTTP.Paths != nil && len(rule.HTTP.Paths) > 0 {
			path := rule.HTTP.Paths[0].Path

			// Remove * from path if exists
			path = strings.TrimRight(path, "*")
			// Remove regex caputure group from path if exists
			parsed := strings.Split(path, "(")
			path = parsed[0]

			return path
		}
	}
	return ""
}

func (iw *IngressWrapper) GetURL(forceHttps bool, healthEndpoint string) string {
	if !iw.rulesExist() {
		log.Info("No rules exist in ingress: " + iw.Ingress.GetName())
		return ""
	}

	var URL string

	if host, exists := iw.tryGetTLSHost(forceHttps); exists { // Get TLS Host if it exists
		URL = host
	} else {
		URL = iw.getHost() // Fallback for normal Host
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
		// ingressSubPath
		ingressSubPath := iw.getIngressSubPath()
		u.Path = path.Join(u.Path, ingressSubPath)

		// Find pod by backtracking ingress -> service -> pod
		healthEndpoint, exists := iw.tryGetHealthEndpointFromIngress()

		// Health endpoint from pod successful
		if exists {
			u.Path = path.Join(u.Path, healthEndpoint)
		}
	}
	return u.String()
}

func (iw *IngressWrapper) hasService() (string, bool) {
	ingress := iw.Ingress
	if ingress.Spec.Rules[0].HTTP != nil &&
		ingress.Spec.Rules[0].HTTP.Paths != nil &&
		len(ingress.Spec.Rules[0].HTTP.Paths) > 0 &&
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service != nil &&
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name != "" {
		return ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name, true
	}
	return "", false
}

func (iw *IngressWrapper) tryGetHealthEndpointFromIngress() (string, bool) {
	serviceName, exists := iw.hasService()

	if !exists {
		return "", false
	}

	service := &corev1.Service{}
	err := iw.Client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: iw.Ingress.Namespace}, service)
	if err != nil {
		log.Info(fmt.Sprintf("Get service from kubernetes cluster error:%v", err))
		return "", false
	}

	labels := labels.Set(service.Spec.Selector)

	podList := &corev1.PodList{}
	listOps := &client.ListOptions{
		Namespace:     iw.Ingress.Namespace,
		LabelSelector: labels.AsSelector(),
	}
	err = iw.Client.List(context.TODO(), podList, listOps)
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
