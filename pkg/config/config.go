package config

import (
	log "github.com/sirupsen/logrus"
	"os"
	"time"
	yaml "gopkg.in/yaml.v2"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/stakater/IngressMonitorController/pkg/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	IngressMonitorControllerSecretConfigKey               = "config.yaml"
	IngressMonitorControllerSecretDefaultName							= "imc-config"
)

var (
	IngressMonitorControllerConfig Config
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
	ApiURL            string      `yaml:"apiURL"`
	AlertContacts     string      `yaml:"alertContacts"`
	AlertIntegrations string      `yaml:"alertIntegrations"`
	Username          string      `yaml:"username"`
	Password          string      `yaml:"password"`
	AccountEmail      string      `yaml:"accountEmail"`
	AppInsightsConfig AppInsights `yaml:"appInsightsConfig"`
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

type EmailAction struct {
	SendToServiceOwners bool     `yaml:"send_to_service_owners"`
	CustomEmails        []string `yaml:"custom_emails"`
}

type WebhookAction struct {
	ServiceURI string `yaml:"service_uri"`
}

func LoadControllerConfig(client client.Client) {
	var config Config
	log.Info("Loading YAML Configuration from secret")

	// Retrieve operator namespace
	operatorNamespace, _ := os.LookupEnv("OPERATOR_NAMESPACE")
	if len(operatorNamespace) == 0 {
			log.Info("DEBUG: test")
			operatorNamespaceTemp, err := k8sutil.GetOperatorNamespace()
			if err != nil {
				if err == k8sutil.ErrNoNamespace {
					log.Info("Skipping leader election; not running in a cluster.")
				}
				log.Panic(err)
			}
			operatorNamespace = operatorNamespaceTemp
	}

	configSecretName, _ := os.LookupEnv("CONFIG_SECRET_NAME")
	if len(configSecretName) == 0 {
			configSecretName = IngressMonitorControllerSecretDefaultName
			log.Warn("CONFIG_SECRET_NAME is unset, using default value: imc-config")
	}

	// Retrieve config key from secret
	configKey, err := secret.LoadSecretData(client, configSecretName, operatorNamespace, IngressMonitorControllerSecretConfigKey)

	// Unmarshall
	err = yaml.Unmarshal([]byte(configKey), &config)
	if err != nil {
		log.Panic(err)
	}
	IngressMonitorControllerConfig = config
}

func GetControllerConfig() Config {
	return IngressMonitorControllerConfig
}
