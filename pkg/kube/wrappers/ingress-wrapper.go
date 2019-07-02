package wrappers

import (
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/stakater/IngressMonitorController/pkg/constants"
	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type IngressWrapper struct {
	Ingress    *v1beta1.Ingress
	Namespace  string
	KubeClient kubernetes.Interface
}

func (iw *IngressWrapper) supportsTLS() bool {
	if iw.Ingress.Spec.TLS != nil && len(iw.Ingress.Spec.TLS) > 0 {
		return true
	}
	return false
}

func (iw *IngressWrapper) tryGetTLSHost() (string, bool) {
	if iw.supportsTLS() {
		return "https://" + iw.Ingress.Spec.TLS[0].Hosts[0], true
	}

	annotations := iw.Ingress.GetAnnotations()
	if value, ok := annotations[constants.ForceHTTPSAnnotation]; ok {
		if value == "true" {
			// Annotation exists and is enabled
			return "https://" + iw.Ingress.Spec.Rules[0].Host, true
		}
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
			if strings.ContainsAny(rule.HTTP.Paths[0].Path, "*") {
				return strings.TrimRight(rule.HTTP.Paths[0].Path, "*")
			} else {
				return rule.HTTP.Paths[0].Path
			}
		}
	}
	return ""
}

func (iw *IngressWrapper) GetURL() string {
	if !iw.rulesExist() {
		log.Println("No rules exist in ingress: " + iw.Ingress.GetName())
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

	annotations := iw.Ingress.GetAnnotations()

	if value, ok := annotations[constants.OverridePathAnnotation]; ok {
		u.Path = value
	} else {
		// ingressSubPath
		ingressSubPath := iw.getIngressSubPath()
		u.Path = path.Join(u.Path, ingressSubPath)

		// Find pod by backtracking ingress -> service -> pod
		healthEndpoint, exists := iw.tryGetHealthEndpointFromIngress()

		// Health endpoint from pod successful
		if exists {
			u.Path = path.Join(u.Path, healthEndpoint)
		} else { // Try to get annotation and set

			// Annotation for health Endpoint exists
			if value, ok := annotations[constants.MonitorHealthAnnotation]; ok {
				u.Path = path.Join(u.Path, value)
			}
		}
	}

	return u.String()
}

func (iw *IngressWrapper) hasService() (string, bool) {
	ingress := iw.Ingress
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

	service, err := iw.KubeClient.Core().Services(iw.Ingress.Namespace).Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("Get service from kubernetes cluster error:%v", err)
		return "", false
	}

	set := labels.Set(service.Spec.Selector)

	if pods, err := iw.KubeClient.Core().Pods(iw.Ingress.Namespace).List(meta_v1.ListOptions{LabelSelector: set.AsSelector().String()}); err != nil {
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
