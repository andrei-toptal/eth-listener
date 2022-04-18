package erc20

import (
	"context"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:generate abigen --abi erc20.abi --pkg erc20 --type ERC20 --out erc20.go

var (
	LogTransferSigHash = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	ETHToken = &Token{
		Symbol:   "ETH",
		Decimals: 18,
	}
)

type Token struct {
	ContractAddress common.Address
	Symbol          string
	Decimals        uint8
	token           *ERC20
}

func FetchERC20Token(contractAddress common.Address, client *ethclient.Client) (*Token, error) {
	token, err := NewERC20(contractAddress, client)
	if err != nil {
		return nil, err
	}
	decimals, err := token.Decimals(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}
	symbol, err := token.Symbol(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}
	return &Token{
		ContractAddress: contractAddress,
		Symbol:          symbol,
		Decimals:        decimals,
		token:           token,
	}, nil
}

func (t Token) ConvertValue(value *big.Int) *big.Float {
	bf := new(big.Float).SetMode(big.AwayFromZero)
	return bf.Quo(new(big.Float).SetInt(value), big.NewFloat(math.Pow10(int(t.Decimals))))
}

func (t Token) FetchBalance(client *ethclient.Client, addr common.Address) string {
	var val *big.Int
	var err error
	if t.token == nil {
		val, err = client.PendingBalanceAt(context.Background(), addr)
	} else {
		val, err = t.token.BalanceOf(&bind.CallOpts{}, addr)
	}
	if err != nil {
		return "N/A"
	}
	return t.ConvertValue(val).String()
}
