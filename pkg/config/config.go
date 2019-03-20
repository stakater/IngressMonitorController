package config

import (
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Providers             []Provider `yaml:"providers"`
	EnableMonitorDeletion bool       `yaml:"enableMonitorDeletion"`
	MonitorNameTemplate   string     `yaml:"monitorNameTemplate"`
}

type Provider struct {
	Name          string `yaml:"name"`
	ApiKey        string `yaml:"apiKey"`
	ApiURL        string `yaml:"apiURL"`
	AlertContacts string `yaml:"alertContacts"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	AccountEmail  string `yaml:"accountEmail"`
}

func ReadConfig(filePath string) Config {
	var config Config
	// Read YML
	log.Println("Reading YAML Configuration")
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(err)
	}

	// Unmarshall
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		log.Panic(err)
	}

	return config
}

func GetControllerConfig() Config {
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if len(configFilePath) == 0 {
		configFilePath = "config.yaml"
	}

	config := ReadConfig(configFilePath)

	return config
}
