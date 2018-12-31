package wrappers

import (
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stakater/IngressMonitorController/pkg/util"
	"k8s.io/client-go/kubernetes"
)

const (
	routeTestUrl = "testurl.stackator.com/"
)

func createRouteObjectWithPath(routeName string, namespace string, url string, path string) *routev1.Route {
	route := util.CreateRouteObject(routeName, namespace, url)
	route.Spec.Path = path
	return route
}

func createRouteObjectWithAnnotations(routeName string, namespace string, url string, annotations map[string]string) *routev1.Route {
	route := util.CreateRouteObject(routeName, namespace, url)
	route.ObjectMeta.SetAnnotations(annotations)

	return route
}

func TestRouteWrapper_getURL(t *testing.T) {
	type fields struct {
		route      *routev1.Route
		namespace  string
		kubeClient kubernetes.Interface
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "TestGetUrlWithEmptyPath",
			fields: fields{
				route:      createRouteObjectWithPath("testRoute", "test", routeTestUrl, "/"),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithHelloPath",
			fields: fields{
				route:      createRouteObjectWithPath("testRoute", "test", routeTestUrl, "/hello"),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/hello",
		},
		{
			name: "TestGetUrlWithNoPath",
			fields: fields{
				route:      util.CreateRouteObject("testRoute", "test", routeTestUrl),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithForceHTTPSAnnotation",
			fields: fields{
				route:      createRouteObjectWithAnnotations("testRoute", "test", routeTestUrl, map[string]string{"monitor.stakater.com/forceHttps": "true"}),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "https://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithForceHTTPSAnnotationOff",
			fields: fields{
				route:      createRouteObjectWithAnnotations("testRoute", "test", routeTestUrl, map[string]string{"monitor.stakater.com/forceHttps": "false"}),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithOverridePathAnnotation",
			fields: fields{
				route:      createRouteObjectWithAnnotations("testRoute", "test", routeTestUrl, map[string]string{"monitor.stakater.com/overridePath": "/overriden-path"}),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/overriden-path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iw := &RouteWrapper{
				Route:      tt.fields.route,
				Namespace:  tt.fields.namespace,
				KubeClient: tt.fields.kubeClient,
			}
			if got := iw.GetURL(); got != tt.want {
				t.Errorf("IngressWrapper.getURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
