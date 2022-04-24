package main

import "github.com/ethereum/go-ethereum/common"

type Accounts map[common.Address]string

func NewAccounts(config *Config) Accounts {
	accounts := make(map[common.Address]string)

	for _, acc := range config.Accounts {
		accounts[common.HexToAddress(acc.Address)] = acc.Alias
	}

	return accounts
}

func (a Accounts) Lookup(addr common.Address) string {
	alias, ok := a[addr]
	if !ok {
		return addr.String()
	}
	return alias
}
