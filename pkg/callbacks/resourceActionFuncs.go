package callbacks

import (
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/extensions/v1beta1"
)

// AnnotationFunc is a generic function to return annotations for resource
type AnnotationFunc func(interface{}) map[string]string

// NameFunc is a generic function to return name of resource
type NameFunc func(interface{}) string

// NamespaceFunc is a generic function to return namespace of resource
type NamespaceFunc func(interface{}) string

// ResourceActionFuncs provides generic functions to return name, namespace and annotations etc.
type ResourceActionFuncs struct {
	AnnotationFunc AnnotationFunc
	NameFunc       NameFunc
	NamespaceFunc  NamespaceFunc
}

// GetIngressAnnotation returns the ingress annotations
func GetIngressAnnotation(resource interface{}) map[string]string {
	return resource.(*v1beta1.Ingress).GetAnnotations()
}

// GetIngressName returns the ingress name
func GetIngressName(resource interface{}) string {
	return resource.(*v1beta1.Ingress).GetName()
}

// GetIngressNamespace returns the ingress namespace
func GetIngressNamespace(resource interface{}) string {
	return resource.(*v1beta1.Ingress).GetNamespace()
}

// GetRouteAnnotation returns the route annotations
func GetRouteAnnotation(resource interface{}) map[string]string {
	return resource.(*routev1.Route).GetAnnotations()
}

// GetRouteName returns the route name
func GetRouteName(resource interface{}) string {
	return resource.(*routev1.Route).GetName()
}

// GetRouteNamespace returns the route namespace
func GetRouteNamespace(resource interface{}) string {
	return resource.(*routev1.Route).GetNamespace()
}
