package wrappers

import (
	"context"
	"net/url"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"github.com/stakater/IngressMonitorController/pkg/constants"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IngressWrapper struct {
	Ingress    *v1beta1.Ingress
	Namespace  string
	client client.Client
}

func NewIngressWrapper(ingress *v1beta1.Ingress, namespace string, client client.Client) *IngressWrapper {
	return &IngressWrapper{
		Ingress: ingress,
		Namespace: namespace,
		client: client,
	}
}

func (iw *IngressWrapper) supportsTLS() bool {
	if iw.Ingress.Spec.TLS != nil && len(iw.Ingress.Spec.TLS) > 0 && iw.Ingress.Spec.TLS[0].Hosts != nil && len(iw.Ingress.Spec.TLS[0].Hosts) > 0 && len(iw.Ingress.Spec.TLS[0].Hosts[0]) > 0 {
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

	service := &corev1.Service{}
	err := iw.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: iw.Ingress.Namespace}, service)
	if err != nil {
		log.Printf("Get service from kubernetes cluster error:%v", err)
		return "", false
	}

	labels := labels.Set(service.Spec.Selector)

	podList := &corev1.PodList{}
	listOps := &client.ListOptions{
		Namespace:     iw.Ingress.Namespace,
		LabelSelector: labels.AsSelector(),
	}
	err = iw.client.List(context.TODO(), podList, listOps)
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
