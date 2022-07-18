package wrappers

import (
	"testing"

	"github.com/stakater/IngressMonitorController/pkg/kube"
	"github.com/stakater/IngressMonitorController/pkg/util"
	"k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	testUrl = "testurl.stackator.com"
)

func createIngressObjectWithPath(ingressName string, namespace string, url string, path string) *v1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.Spec.Rules[0].IngressRuleValue = v1.IngressRuleValue{
		HTTP: &v1.HTTPIngressRuleValue{
			Paths: []v1.HTTPIngressPath{
				v1.HTTPIngressPath{
					Path: path,
				},
			},
		},
	}

	return ingress
}

func createIngressObjectWithAnnotations(ingressName string, namespace string, url string, annotations map[string]string) *v1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.ObjectMeta.SetAnnotations(annotations)

	return ingress
}

func createIngressObjectWithTLS(ingressName string, namespace string, url string, tlsHostname string) *v1.Ingress {
	ingress := util.CreateIngressObject(ingressName, namespace, url)
	ingress.Spec.TLS = []v1.IngressTLS{
		v1.IngressTLS{
			Hosts: []string{
				tlsHostname,
			},
		},
	}
	return ingress
}

func TestIngressWrapper_getURL(t *testing.T) {
	type fields struct {
		ingress    *v1.Ingress
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
				ingress:    createIngressObjectWithPath("testIngress", "test", testUrl, "/"),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/",
		},
		{
			name: "TestGetUrlWithHelloPath",
			fields: fields{
				ingress:    createIngressObjectWithPath("testIngress", "test", testUrl, "/hello"),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/hello",
		},
		{
			name: "TestGetUrlWithNoPath",
			fields: fields{
				ingress:    util.CreateIngressObject("testIngress", "test", testUrl),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com",
		},
		{
			name: "TestGetUrlWithForceHTTPSAnnotation",
			fields: fields{
				ingress:    createIngressObjectWithAnnotations("testIngress", "test", testUrl, map[string]string{"monitor.stakater.com/forceHttps": "true"}),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "https://testurl.stackator.com",
		},
		{
			name: "TestGetUrlWithForceHTTPSAnnotationOff",
			fields: fields{
				ingress:    createIngressObjectWithAnnotations("testIngress", "test", testUrl, map[string]string{"monitor.stakater.com/forceHttps": "false"}),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com",
		},
		{
			name: "TestGetUrlWithOverridePathAnnotation",
			fields: fields{
				ingress:    createIngressObjectWithAnnotations("testIngress", "test", testUrl, map[string]string{"monitor.stakater.com/overridePath": "/overriden-path"}),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/overriden-path",
		}, {
			name: "TestGetUrlWithWildCardInPath",
			fields: fields{
				ingress:    createIngressObjectWithPath("testIngress", "test", testUrl, "/*"),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/",
		}, {
			name: "TestGetUrlWithRegexCaptureGroupInPath",
			fields: fields{
				ingress:    createIngressObjectWithPath("testIngress", "test", testUrl, "/api(/|$)(.*)"),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com/api",
		}, {
			name: "TestGetUrlWithTLS",
			fields: fields{
				ingress:    createIngressObjectWithTLS("testIngress", "test", testUrl, "customtls.stackator.com"),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "https://customtls.stackator.com",
		}, {
			name: "TestGetUrlWithEmptyTLS",
			fields: fields{
				ingress:    createIngressObjectWithTLS("testIngress", "test", testUrl, ""),
				namespace:  "test",
				kubeClient: getTestKubeClient(),
			},
			want: "http://testurl.stackator.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iw := &IngressWrapper{
				Ingress:    tt.fields.ingress,
				Namespace:  tt.fields.namespace,
				KubeClient: tt.fields.kubeClient,
			}
			if got := iw.GetURL(); got != tt.want {
				t.Errorf("IngressWrapper.getURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getTestKubeClient() kubernetes.Interface {
	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = kube.GetClientOutOfCluster()
	} else {
		kubeClient = kube.GetClient()
	}

	return kubeClient
}
