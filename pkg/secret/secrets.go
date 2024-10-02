package secret

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LoadSecretData loads a given secret key and returns it's data as a string.
func LoadSecretData(apiReader client.Reader, secretName, namespace, dataKey string) (string, error) {
	s := &corev1.Secret{}
	err := apiReader.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, s)
	if err != nil {
		return "", err
	}

	retStr, ok := s.Data[dataKey]
	if !ok {
		return "", fmt.Errorf("secret %s did not contain key %s", secretName, dataKey)
	}
	return string(retStr), nil
}

func ReadBasicAuthSecret(apiReader clientv1.SecretInterface, secretName string) (string, string, error) {
	secret, err := apiReader.Get(context.TODO(), secretName, metav1.GetOptions{})
	var username, password string
	if err != nil {
		return "", "", err
	}

	for key, value := range secret.Data {
		switch key {
		case "username":
			username = string(value)
		case "password":
			password = string(value)
		default:
			return "", "", fmt.Errorf("secret %s contained unkown key %s", secretName, key)
		}
	}

	return username, password, err
}
