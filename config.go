package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type TelegramConfig struct {
	Token    string `yaml:"token"`
	Username string `yaml:"username"`
}

type AliasConfig struct {
	Address string `yaml:"address"`
	Alias   string `yaml:"alias"`
}

type Config struct {
	EthUrl    string          `yaml:"eth-url"`
	Addresses []string        `yaml:"addresses"`
	Tokens    []string        `yaml:"tokens"`
	Telegram  *TelegramConfig `yaml:"telegram"`
	Aliases   []AliasConfig   `yaml:"aliases"`
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
