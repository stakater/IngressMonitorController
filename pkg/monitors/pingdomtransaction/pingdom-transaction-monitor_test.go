package pingdomtransaction

import (
	"context"
	"testing"

	pingdomNew "github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/openapi"
	"github.com/karlderkaefer/pingdom-golang-client/pkg/pingdom/openapi/ptr"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/util"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// To allow normal logging to be printed if tests fails
	// Dev mode is an extra feature to make output more readable
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
}

func TestAddMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := PingdomTransactionMonitorService{}
	provider := util.GetProviderWithName(config, "PingdomTransaction")
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error(nil, "Failed to find provider")
		return
	}

	service.Setup(*provider)
	m := models.Monitor{Name: "google-test", URL: "https://google1.com"}

	service.Add(m)

	mRes, err := service.GetByName("google-test")
	assert.NilError(t, err)

	defer func() {
		// Cleanup
		service.Remove(*mRes)
	}()

	if err != nil {
		t.Error("Error: " + err.Error())
	}

	assert.Equal(t, mRes.Name, m.Name)
	assert.Equal(t, mRes.URL, "https://google1.com")
}

func TestUpdateMonitorWithCorrectValues(t *testing.T) {
	config := config.GetControllerConfigTest()

	service := PingdomTransactionMonitorService{}

	provider := util.GetProviderWithName(config, "Pingdom")
	if provider == nil {
		// TODO: Currently forcing to pass the test as we dont have Pingdom account to test
		//       Fail this case in future when have a valid Pingdom account
		log.Error(nil, "Failed to find provider")
		return
	}
	service.Setup(*provider)

	// Create initial record
	m := models.Monitor{Name: "google-update-test", URL: "https://google.com"}
	service.Add(m)

	mRes, err := service.GetByName("google-update-test")
	assert.NilError(t, err)

	defer func() {
		// Cleanup
		service.Remove(*mRes)
	}()

	// Update the record
	mRes.URL = "https://facebook.com"

	service.Update(*mRes)

	mRes, err = service.GetByName("google-update-test")
	if err != nil {
		t.Error("Error: " + err.Error())
	}

	assert.Equal(t, mRes.Name, m.Name)
	assert.Equal(t, mRes.URL, "https://facebook.com")
}

func TestGetSecretFromTemplate(t *testing.T) {
	var tests = []struct {
		name               string
		content            string
		expectedSecretName string
		expectedSecretKey  string
	}{
		{
			name:               "With Secret",
			content:            "This is a sample content with {{secret:my-secret:my-key}} embedded in it.",
			expectedSecretName: "my-secret",
			expectedSecretKey:  "my-key",
		},
		{
			name:               "No Secret",
			content:            "This is a sample content without any secret.",
			expectedSecretName: "",
			expectedSecretKey:  "",
		},
		{
			name:               "Invalid Format",
			content:            "This is a sample content with invalid secret format {{secret:my-secret}}",
			expectedSecretName: "",
			expectedSecretKey:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			secretName, secretKey := parseSecretTemplate(tc.content)
			assert.Equal(t, secretName, tc.expectedSecretName)
			assert.Equal(t, secretKey, tc.expectedSecretKey)
		})
	}
}

func TestReplaceSecrets(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create and add a secret to the fake clientset
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("simple-password"),
		},
	}
	_, err := clientset.CoreV1().Secrets("default").Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error injecting secret add: %v", err)
	}

	// Define test cases
	testCases := []struct {
		name             string
		password         string
		value            string
		expectedPassword string
		expectedValue    string
		expectError      bool
	}{
		{
			name:             "Password field with secret",
			password:         "{{secret:my-secret:password}}",
			value:            "{{secret:my-secret:username}}",
			expectedPassword: "simple-password",
			expectedValue:    "admin",
			expectError:      false,
		},
		{
			name:             "Password field without secret",
			value:            "no-secret",
			password:         "no-secret",
			expectedPassword: "no-secret",
			expectedValue:    "no-secret",
			expectError:      false,
		},
		{
			name:             "Password field with invalid secret",
			password:         "{{secret:my-secret:invalidkey}}",
			value:            "no-secret",
			expectedPassword: "",
			expectedValue:    "no-secret",
			expectError:      true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := &pingdomNew.StepArgs{
				Password: ptr.String(tc.password),
				Value:    ptr.String(tc.value),
			}
			err := replaceSecretValuesInArgs(args, clientset, secret.Namespace)
			if tc.expectError {
				assert.Error(t, err, "failed to get secret: secret my-secret does not contain key invalidkey")
				return
			}
			assert.Equal(t, *args.Password, tc.expectedPassword)
			assert.Equal(t, *args.Value, tc.expectedValue)
		})
	}
}
