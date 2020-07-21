package wrappers

import (
	"testing"

	"github.com/stakater/IngressMonitorController/pkg/util"
	"k8s.io/api/extensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakekubeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testUrl = "testurl.stackator.com"
)

func createIngressObjectWithPath(ingressName string, namespace string, url string, path string) *v1beta1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.Spec.Rules[0].IngressRuleValue = v1beta1.IngressRuleValue{
		HTTP: &v1beta1.HTTPIngressRuleValue{
			Paths: []v1beta1.HTTPIngressPath{
				{
					Path: path,
				},
			},
		},
	}

	return ingress
}

func createIngressObjectWithAnnotations(ingressName string, namespace string, url string, annotations map[string]string) *v1beta1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.ObjectMeta.SetAnnotations(annotations)

	return ingress
}

func createIngressObjectWithTLS(ingressName string, namespace string, url string, tlsHostname string) *v1beta1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.Spec.TLS = []v1beta1.IngressTLS{
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
		ingress   *v1beta1.Ingress
		namespace string
		Client    client.Client
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
				Client:  fakekubeclient.NewFakeClient(),
			},
			want: "http://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithHelloPath",
			fields: fields{
				ingress: createIngressObjectWithPath("testIngress", "test", testUrl, "/hello"),
				Client:  fakekubeclient.NewFakeClient(),
			},
			want: "http://testurl.stackator.com/hello",
		},
		{
			name: "TestGetUrlWithNoPath",
			fields: fields{
				ingress: util.CreateIngressObject("testIngress", "test", testUrl),
				Client:  fakekubeclient.NewFakeClient(),
			},
			want: "http://testurl.stackator.com",
		},
		{
			name: "TestGetUrlWithWildCardInPath",
			fields: fields{
				ingress: createIngressObjectWithPath("testIngress", "test", testUrl, "/*"),
				Client:  fakekubeclient.NewFakeClient(),
			},
			want: "http://testurl.stackator.com/",
		}, {
			name: "TestGetUrlWithTLS",
			fields: fields{
				ingress: createIngressObjectWithTLS("testIngress", "test", testUrl, "customtls.stackator.com"),
				Client:  fakekubeclient.NewFakeClient(),
			},
			want: "https://customtls.stackator.com",
		}, {
			name: "TestGetUrlWithEmptyTLS",
			fields: fields{
				ingress: createIngressObjectWithTLS("testIngress", "test", testUrl, ""),
				Client:  fakekubeclient.NewFakeClient(),
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
