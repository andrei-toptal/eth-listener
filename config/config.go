package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Telegram struct {
	Token    string `yaml:"token"`
	Username string `yaml:"username"`
}

type Config struct {
	EthUrl    string    `yaml:"eth-url"`
	Addresses []string  `yaml:"addresses"`
	Tokens    []string  `yaml:"tokens"`
	Telegram  *Telegram `yaml:"telegram"`
}

func LoadConfig(filepath string) (config *Config, err error) {
	configData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	config = &Config{}
	if err = yaml.Unmarshal(configData, config); err != nil {
		return
	}

	if config.EthUrl == "" {
		log.Fatalf("Config is missing eth-url")
	}
	if config.Telegram == nil {
		log.Fatalf("Config is missing telegram settings")
	}
	if config.Telegram.Token == "" {
		log.Fatalf("Config is missing telegram token")
	}

	return
}
