package main

import (
	"testing"

	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

const (
	testUrl = "testurl.stackator.com"
)

func createIngressObjectWithPath(ingressName string, namespace string, url string) *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      ingressName,
			Namespace: namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: url,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								v1beta1.HTTPIngressPath{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: "test",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}

func TestIngressWrapper_getURL(t *testing.T) {
	type fields struct {
		ingress    *v1beta1.Ingress
		namespace  string
		kubeClient kubernetes.Interface
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "TestGetUrlWithPath",
			fields: fields{
				ingress:    createIngressObjectWithPath("testIngress", "test", testUrl),
				namespace:  "test",
				kubeClient: nil,
			},
			want: "want",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iw := &IngressWrapper{
				ingress:    tt.fields.ingress,
				namespace:  tt.fields.namespace,
				kubeClient: tt.fields.kubeClient,
			}
			if got := iw.getURL(); got != tt.want {
				t.Errorf("IngressWrapper.getURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
