package main

import (
	"github.com/andrei-toptal/eth-listener/token"
	"github.com/ethereum/go-ethereum/ethclient"
)

type App struct {
	config        *Config
	tokensDB      token.TokensDB
	accounts      Accounts
	telegram      Telegram
	client        *ethclient.Client
	tokensManager token.TokensManager
}

func NewApp(config *Config, tokensDB token.TokensDB, accounts Accounts, telegram Telegram, client *ethclient.Client, tokensManager token.TokensManager) *App {
	return &App{
		config:        config,
		tokensDB:      tokensDB,
		accounts:      accounts,
		telegram:      telegram,
		client:        client,
		tokensManager: tokensManager,
	}
}
