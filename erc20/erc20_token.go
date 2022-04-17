package erc20

import (
	"fmt"
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
)

type ERC20Token struct {
	ContractAddress common.Address
	Name            string
	Symbol          string
	Decimals        uint8
	token           *ERC20
}

func FetchERC20Token(contractAddress common.Address, client *ethclient.Client) (*ERC20Token, error) {
	token, err := NewERC20(contractAddress, client)
	if err != nil {
		return nil, err
	}
	name, err := token.Name(&bind.CallOpts{})
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
	return &ERC20Token{
		ContractAddress: contractAddress,
		Name:            name,
		Symbol:          symbol,
		Decimals:        decimals,
		token:           token,
	}, nil
}

func (t ERC20Token) ConvertValue(value *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(math.Pow10(int(t.Decimals))))
}

func (t ERC20Token) FetchBalance(client *ethclient.Client, addr common.Address) (*big.Float, error) {
	balance, err := t.token.BalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		return nil, err
	}
	return t.ConvertValue(balance), nil
}

func (t ERC20Token) String() string {
	return fmt.Sprintf("%s (%s) at %s", t.Symbol, t.Name, t.ContractAddress.String())
}
