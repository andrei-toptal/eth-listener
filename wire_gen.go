// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

// Injectors from app.go:

func WireApp(configPath string) (*App, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	mainTokensDB := NewTokensDB()
	accounts := NewAccounts(config)
	mainTelegram := NewTelegram(config)
	client, err := newEthClient(config)
	if err != nil {
		return nil, err
	}
	mainTokensManager := NewTokensManager(client, mainTokensDB)
	app := NewApp(config, mainTokensDB, accounts, mainTelegram, client, mainTokensManager)
	return app, nil
}

// app.go:

type App struct {
	config        *Config
	tokensDB      TokensDB
	accounts      Accounts
	telegram      Telegram
	client        *ethclient.Client
	tokensManager TokensManager
}

func NewApp(config *Config, tokensDB2 TokensDB, accounts Accounts, telegram2 Telegram, client *ethclient.Client, tokensManager2 TokensManager) *App {
	return &App{
		config:        config,
		tokensDB:      tokensDB2,
		accounts:      accounts,
		telegram:      telegram2,
		client:        client,
		tokensManager: tokensManager2,
	}
}

func newEthClient(config *Config) (*ethclient.Client, error) {
	return ethclient.Dial(config.EthUrl)
}