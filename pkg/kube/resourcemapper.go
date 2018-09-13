package kube

import (
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stakater/IngressMonitorController/pkg/callbacks"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

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

// ResourceMap are resources from where changes are going to be detected
var ResourceMap = map[string]runtime.Object{
	"ingresses": &v1beta1.Ingress{},
	"routes":    &routev1.Route{},
}
