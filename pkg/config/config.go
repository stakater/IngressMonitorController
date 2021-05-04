package config

import (
	"io/ioutil"
	"os"
	"time"

	util "github.com/stakater/operator-utils/util"
	yaml "gopkg.in/yaml.v2"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/stakater/IngressMonitorController/pkg/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	IngressMonitorControllerSecretConfigKey   = "config.yaml"
	IngressMonitorControllerSecretDefaultName = "imc-config"
)

var (
	IngressMonitorControllerConfig Config
	log                            = logf.Log.WithName("config")
)

type Config struct {
	Providers             []Provider    `yaml:"providers"`
	EnableMonitorDeletion bool          `yaml:"enableMonitorDeletion"`
	MonitorNameTemplate   string        `yaml:"monitorNameTemplate"`
	ResyncPeriod          int           `yaml:"resyncPeriod,omitempty"`
	CreationDelay         time.Duration `yaml:"creationDelay,omitempty"`
}

// UnmarshalYAML interface to deserialize specific types
func (c *Config) UnmarshalYAML(data []byte) error {
	type Alias Config
	aux := struct {
		CreationDelay string `yaml:"creationDelay,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := yaml.Unmarshal(data, &aux); err != nil {
		return err
	}

	delay, err := time.ParseDuration(aux.CreationDelay)
	if err != nil {
		return err
	}
	c.CreationDelay = delay

	return nil
}

type Provider struct {
	Name              string      `yaml:"name"`
	ApiKey            string      `yaml:"apiKey"`
	ApiToken          string      `yaml:"apiToken"`
	ApiURL            string      `yaml:"apiURL"`
	AlertContacts     string      `yaml:"alertContacts"`
	AlertIntegrations string      `yaml:"alertIntegrations"`
	TeamAlertContacts string      `yaml:"teamAlertContacts"`
	Username          string      `yaml:"username"`
	Password          string      `yaml:"password"`
	AccountEmail      string      `yaml:"accountEmail"`
	AppInsightsConfig AppInsights `yaml:"appInsightsConfig"`
	GcloudConfig      Gcloud      `yaml:"gcloudConfig"`
}

type AppInsights struct {
	Name          string        `yaml:"name"`
	Location      string        `yaml:"location"`
	ResourceGroup string        `yaml:"resourceGroup"`
	Frequency     int32         `yaml:"frequency"`
	GeoLocation   []interface{} `yaml:"geoLocation"`
	EmailAction   EmailAction   `yaml:"emailAction"`
	WebhookAction WebhookAction `yaml:"webhookAction"`
}

type Gcloud struct {
	ProjectID string `yaml:"projectId"`
}

type EmailAction struct {
	SendToServiceOwners bool     `yaml:"send_to_service_owners"`
	CustomEmails        []string `yaml:"custom_emails"`
}

type WebhookAction struct {
	ServiceURI string `yaml:"service_uri"`
}

func LoadControllerConfig(apiReader client.Reader) {
	var config Config
	log.Info("Loading YAML Configuration from secret")

	// Retrieve operator namespace
	operatorNamespace, _ := os.LookupEnv("OPERATOR_NAMESPACE")
	if len(operatorNamespace) == 0 {
		operatorNamespaceTemp, err := util.GetOperatorNamespace()
		if err != nil {
			log.Error(err, "Unable to get operator namespace")
			panic("Unable to get operator namespace")
		}
		operatorNamespace = operatorNamespaceTemp
	}

	configSecretName, _ := os.LookupEnv("CONFIG_SECRET_NAME")
	if len(configSecretName) == 0 {
		configSecretName = IngressMonitorControllerSecretDefaultName
		log.Info("CONFIG_SECRET_NAME is unset, using default value: imc-config")
	}

	// Retrieve config key from secret
	configKey, err := secret.LoadSecretData(apiReader, configSecretName, operatorNamespace, IngressMonitorControllerSecretConfigKey)
	if err != nil {
		panic(err)
	}

	// Unmarshall
	err = yaml.Unmarshal([]byte(configKey), &config)
	if err != nil {
		panic(err)
	}
	IngressMonitorControllerConfig = config
}

func GetControllerConfig() Config {
	return IngressMonitorControllerConfig
}

func GetControllerConfigTest() Config {
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if len(configFilePath) == 0 {
		configFilePath = "../../../.local/test-config.yaml"
	}

	config := ReadConfig(configFilePath)

	return config
}

func ReadConfig(filePath string) Config {
	var config Config
	// Read YML
	log.Info("Reading YAML Configuration: ", filePath)
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	// Unmarshall
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
	IngressMonitorControllerConfig = config
	return config
}
