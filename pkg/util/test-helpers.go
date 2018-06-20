package util

import (
	"testing"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/extensions/v1beta1"
)

func AssertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func CreateIngressObject(ingressName string, namespace string, url string) *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      ingressName,
			Namespace: namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: url,
				},
			},
		},
	}

	return ingress
}
