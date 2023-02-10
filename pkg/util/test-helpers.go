package util

import (
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/networking/v1"
)

func AssertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func CreateIngressObject(ingressName string, namespace string, url string) *v1.Ingress {
	ingress := &v1.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      ingressName,
			Namespace: namespace,
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					Host: url,
				},
			},
		},
	}

	return ingress
}

// CreateRouteObject creates an openshift route object
func CreateRouteObject(routeName string, namespace string, url string) *routev1.Route {
	route := &routev1.Route{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      routeName,
			Namespace: namespace,
		},
		Spec: routev1.RouteSpec{
			Host: url,
		},
	}
	return route
}

func GetProviderWithName(controllerConfig config.Config, name string) *config.Provider {
	for _, provider := range controllerConfig.Providers {
		if provider.Name == name {
			return &provider
		}
	}

	return nil
}
