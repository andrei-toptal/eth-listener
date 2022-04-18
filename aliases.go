package main

import (
	"github.com/ethereum/go-ethereum/common"
)

type Aliases map[common.Address]string

func NewAliases(aliases []AliasConfig) Aliases {
	am := make(map[common.Address]string)
	for _, alias := range aliases {
		am[common.HexToAddress(alias.Address)] = alias.Alias
	}
	return am
}
