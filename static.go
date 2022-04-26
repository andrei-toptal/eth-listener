package main

import (
	"strings"

	"github.com/andrei-toptal/eth-listener/token/erc20"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ConfigPath        = "config.yaml"
	TokensDBPath      = ".tokens-db"
	TransfersChBuffer = 32
)

var (
	LogTransferSigHash = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ERC20ABI           abi.ABI
)

func init() {
	abi, err := abi.JSON(strings.NewReader(string(erc20.ERC20MetaData.ABI)))
	if err != nil {
		panic(err)
	}
	ERC20ABI = abi
}
