package wrappers

import (
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakekubeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	routeTestUrl = "testurl.stackator.com/"
)

func createRouteObjectWithPath(routeName string, namespace string, url string, path string) *routev1.Route {
	route := util.CreateRouteObject(routeName, namespace, url)
	route.Spec.Path = path
	return route
}

func TestRouteWrapper_getURL(t *testing.T) {
	type fields struct {
		route  *routev1.Route
		Client client.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "TestGetUrlWithEmptyPath",
			fields: fields{
				route:  createRouteObjectWithPath("testRoute", "test", routeTestUrl, "/"),
				Client: fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithHelloPath",
			fields: fields{
				route:  createRouteObjectWithPath("testRoute", "test", routeTestUrl, "/hello"),
				Client: fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com/hello",
		},
		{
			name: "TestGetUrlWithNoPath",
			fields: fields{
				route:  util.CreateRouteObject("testRoute", "test", routeTestUrl),
				Client: fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iw := &RouteWrapper{
				Route:  tt.fields.route,
				Client: tt.fields.Client,
			}
			if got := iw.GetURL(false, ""); got != tt.want {
				t.Errorf("IngressWrapper.getURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
