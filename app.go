//go:build wireinject

package main

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/wire"
)

type App struct {
	config        *Config
	tokensDB      TokensDB
	accounts      Accounts
	telegram      Telegram
	client        *ethclient.Client
	tokensManager TokensManager
}

func NewApp(config *Config, tokensDB TokensDB, accounts Accounts, telegram Telegram, client *ethclient.Client, tokensManager TokensManager) *App {
	return &App{
		config:        config,
		tokensDB:      tokensDB,
		accounts:      accounts,
		telegram:      telegram,
		client:        client,
		tokensManager: tokensManager,
	}
}

func newEthClient(config *Config) (*ethclient.Client, error) {
	return ethclient.Dial(config.EthUrl)
}

func WireApp(configPath string) (*App, error) {
	wire.Build(NewApp, LoadConfig, NewTokensDB, NewAccounts, NewTelegram, newEthClient, NewTokensManager)
	return nil, nil
}
