package main

import (
	"log"
	"net/url"
	"path"

	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	IngressForceHTTPSAnnotation = "monitor.stakater.com/forceHttps"
)

type IngressWrapper struct {
	ingress    *v1beta1.Ingress
	namespace  string
	kubeClient kubernetes.Interface
}

func (iw *IngressWrapper) supportsTLS() bool {
	if iw.ingress.Spec.TLS != nil && len(iw.ingress.Spec.TLS) > 0 {
		return true
	}
	return false
}

func (iw *IngressWrapper) tryGetTLSHost() (string, bool) {
	if iw.supportsTLS() {
		return "https://" + iw.ingress.Spec.TLS[0].Hosts[0], true
	}

	annotations := iw.ingress.GetAnnotations()
	if value, ok := annotations[IngressForceHTTPSAnnotation]; ok {
		if value == "true" {
			// Annotation exists and is enabled
			return "https://" + iw.ingress.Spec.Rules[0].Host, true
		}
	}

	return "", false
}

func (iw *IngressWrapper) getHost() string {
	return "http://" + iw.ingress.Spec.Rules[0].Host
}

func (iw *IngressWrapper) rulesExist() bool {
	if iw.ingress.Spec.Rules != nil && len(iw.ingress.Spec.Rules) > 0 {
		return true
	}
	return false
}

func (iw *IngressWrapper) getIngressSubPathWithPort() string {
	port := iw.getIngressPort()
	subPath := iw.getIngressSubPath()

	return port + subPath
}

func (iw *IngressWrapper) getIngressPort() string {
	rule := iw.ingress.Spec.Rules[0]
	if rule.HTTP != nil {
		if rule.HTTP.Paths != nil && len(rule.HTTP.Paths) > 0 {
			return rule.HTTP.Paths[0].Backend.ServicePort.StrVal
		}
	}
	return ""
}

func (iw *IngressWrapper) getIngressSubPath() string {
	rule := iw.ingress.Spec.Rules[0]
	if rule.HTTP != nil {
		if rule.HTTP.Paths != nil && len(rule.HTTP.Paths) > 0 {
			return rule.HTTP.Paths[0].Path
		}
	}
	return ""
}

func (iw *IngressWrapper) getURL() string {
	if !iw.rulesExist() {
		log.Println("No rules exist in ingress: " + iw.ingress.GetName())
		return ""
	}

	var URL string

	if host, exists := iw.tryGetTLSHost(); exists { // Get TLS Host if it exists
		URL = host
	} else {
		URL = iw.getHost() // Fallback for normal Host
	}

	// Convert url to url object
	u, err := url.Parse(URL)

	if err != nil {
		log.Printf("URL parsing error in getURL() :%v", err)
		return ""
	}

	// Append port + ingressSubPath
	u.Path = path.Join(u.Path, iw.getIngressSubPathWithPort())

	// Find pod by backtracking ingress -> service -> pod
	healthEndpoint, exists := iw.tryGetHealthEndpointFromIngress()

	// Health endpoint from pod successful
	if exists {
		u.Path = path.Join(u.Path, healthEndpoint)
	} else { // Try to get annotation and set
		annotations := iw.ingress.GetAnnotations()

		// Annotation for health Endpoint exists
		if value, ok := annotations[monitorHealthAnnotation]; ok {
			u.Path = path.Join(u.Path, value)
		}
	}

	return u.String()
}

func (iw *IngressWrapper) hasService() (string, bool) {
	ingress := iw.ingress
	if ingress.Spec.Rules[0].HTTP != nil &&
		ingress.Spec.Rules[0].HTTP.Paths != nil &&
		len(ingress.Spec.Rules[0].HTTP.Paths) > 0 &&
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName != "" {
		return ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName, true
	}
	return "", false
}

func (iw *IngressWrapper) tryGetHealthEndpointFromIngress() (string, bool) {

	serviceName, exists := iw.hasService()

	if !exists {
		return "", false
	}

	service, err := iw.kubeClient.Core().Services(iw.ingress.Namespace).Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("Get service from kubernetes cluster error:%v", err)
		return "", false
	}

	set := labels.Set(service.Spec.Selector)

	if pods, err := iw.kubeClient.Core().Pods(iw.ingress.Namespace).List(meta_v1.ListOptions{LabelSelector: set.AsSelector().String()}); err != nil {
		log.Printf("List Pods of service[%s] error:%v", service.GetName(), err)
	} else if len(pods.Items) > 0 {
		pod := pods.Items[0]

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
