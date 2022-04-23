package main

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pinebit/eth-listener/erc20"
)

type Token struct {
	Address  common.Address
	Symbol   string
	Decimals uint8
}

var zeroAddress = common.HexToAddress("0x0")

// To unify transfers processing we mimic ETH to be a token.
// Address is set to 0 and not used for this "token".
var ETHToken = &Token{
	Address:  zeroAddress,
	Symbol:   "ETH",
	Decimals: 18,
}

type TokensManager interface {
	// Returns token for the given contract address or error.
	// For unknown tokens this fetches token details from the contract.
	// NOT THREAD SAFE
	GetToken(ctx context.Context, contractAddress common.Address) (*Token, error)
	// Returns token's balance for the given address or error.
	FetchBalance(ctx context.Context, token *Token, addr common.Address) (*big.Int, error)
}

type tokensManager struct {
	client *ethclient.Client
	tokens map[common.Address]*Token
}

func NewTokensManager(client *ethclient.Client) TokensManager {
	tokens := make(map[common.Address]*Token)
	tokens[ETHToken.Address] = ETHToken

	return &tokensManager{
		client: client,
		tokens: tokens,
	}
}

func (tm *tokensManager) GetToken(ctx context.Context, contractAddress common.Address) (*Token, error) {
	if t, has := tm.tokens[contractAddress]; has {
		return t, nil
	}

	token, err := erc20.NewERC20(contractAddress, tm.client)
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

	t := &Token{
		Address:  contractAddress,
		Symbol:   symbol,
		Decimals: decimals,
	}

	tm.tokens[contractAddress] = t

	return t, nil
}

func (tm *tokensManager) FetchBalance(ctx context.Context, token *Token, addr common.Address) (*big.Int, error) {
	if token == ETHToken {
		return tm.client.PendingBalanceAt(ctx, addr)
	}

	erc20t, err := erc20.NewERC20(token.Address, tm.client)
	if err != nil {
		return nil, err
	}
	return erc20t.BalanceOf(&bind.CallOpts{}, addr)
}

func (t Token) RenderValue(value *big.Int) string {
	val := new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(math.Pow10(int(t.Decimals))))
	return fmt.Sprintf("%s %s", val.String(), t.Symbol)
}
