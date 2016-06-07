package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func loadConfig(filePath string) (configuration, error) {
	conf := configuration{}

	f, err := os.Open(filePath)
	if err != nil {
		return conf, err
	}
	defer f.Close()

	confBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(confBytes, &conf)

	return conf, err
}

type configuration struct {
	Versions map[string]string `yaml:"versions"`
}
