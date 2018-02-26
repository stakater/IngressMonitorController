package main

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Providers             []Provider `yaml:"providers"`
	EnableMonitorDeletion bool       `yaml:"enableMonitorDeletion"`
}

type Provider struct {
	Name          string `yaml:"name"`
	ApiKey        string `yaml:"apiKey"`
	ApiURL        string `yaml:"apiURL"`
	AlertContacts string `yaml:"alertContacts"`
}

func ReadConfig(filePath string) Config {
	var config Config
	// Read YML
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
