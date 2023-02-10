package wrappers

import (
	"testing"

	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	v1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakekubeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testUrl = "testurl.stackator.com"
)

func createIngressObjectWithPath(ingressName string, namespace string, url string, path string) *v1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.Spec.Rules[0].IngressRuleValue = v1.IngressRuleValue{
		HTTP: &v1.HTTPIngressRuleValue{
			Paths: []v1.HTTPIngressPath{
				{
					Path: path,
				},
			},
		},
	}

	return ingress
}

func createIngressObjectWithTLS(ingressName string, namespace string, url string, tlsHostname string) *v1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.Spec.TLS = []v1.IngressTLS{
		{
			Hosts: []string{
				tlsHostname,
			},
		},
	}
	return ingress
}

func TestIngressWrapper_getURL(t *testing.T) {
	type fields struct {
		ingress *v1.Ingress
		Client  client.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "TestGetUrlWithEmptyPath",
			fields: fields{
				ingress: createIngressObjectWithPath("testIngress", "test", testUrl, "/"),
				Client:  fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithHelloPath",
			fields: fields{
				ingress: createIngressObjectWithPath("testIngress", "test", testUrl, "/hello"),
				Client:  fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com/hello",
		},
		{
			name: "TestGetUrlWithNoPath",
			fields: fields{
				ingress: util.CreateIngressObject("testIngress", "test", testUrl),
				Client:  fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com",
		},
		{
			name: "TestGetUrlWithWildCardInPath",
			fields: fields{
				ingress: createIngressObjectWithPath("testIngress", "test", testUrl, "/*"),
				Client:  fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com/",
		}, {
			name: "TestGetUrlWithRegexCaptureGroupInPath",
			fields: fields{
				ingress: createIngressObjectWithPath("testIngress", "test", testUrl, "/api(/|$)(.*)"),
				Client:  fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com/api",
		}, {
			name: "TestGetUrlWithTLS",
			fields: fields{
				ingress: createIngressObjectWithTLS("testIngress", "test", testUrl, "customtls.stackator.com"),
				Client:  fakekubeclient.NewClientBuilder().Build(),
			},
			want: "https://customtls.stackator.com",
		}, {
			name: "TestGetUrlWithEmptyTLS",
			fields: fields{
				ingress: createIngressObjectWithTLS("testIngress", "test", testUrl, ""),
				Client:  fakekubeclient.NewClientBuilder().Build(),
			},
			want: "http://testurl.stackator.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iw := &IngressWrapper{
				Ingress: tt.fields.ingress,
				Client:  tt.fields.Client,
			}
			if got := iw.GetURL(false, ""); got != tt.want {
				t.Errorf("IngressWrapper.getURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
