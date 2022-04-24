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

type AccountConfig struct {
	Address string `yaml:"address"`
	Alias   string `yaml:"alias"`
}

type Config struct {
	EthUrl   string          `yaml:"eth-url"`
	Accounts []AccountConfig `yaml:"accounts"`
	Telegram *TelegramConfig `yaml:"telegram"`
}

func LoadConfig(configPath string) (config *Config, err error) {
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config = &Config{}
	if err = yaml.Unmarshal(configData, config); err != nil {
		return
	}

	if len(config.Accounts) == 0 {
		log.Fatalf("Config is missing accounts")
	}
	if config.EthUrl == "" {
		log.Fatalf("Config is missing eth-url")
	}

	return
}
