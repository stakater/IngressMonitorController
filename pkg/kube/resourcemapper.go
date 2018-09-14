package kube

import (
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceMap are resources from where changes are going to be detected
var ResourceMap = map[string]runtime.Object{
	"ingresses": &v1beta1.Ingress{},
	"routes":    &routev1.Route{},
}
