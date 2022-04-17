package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	EthUrl    string   `yaml:"eth-url"`
	Addresses []string `yaml:"addresses"`
	Tokens    []string `yaml:"tokens"`
}

func LoadConfig(filepath string) (config *Config, err error) {
	configData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	config = &Config{}
	err = yaml.Unmarshal(configData, config)
	return
}
