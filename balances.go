package main

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pinebit/eth-listener/erc20"
	"go.uber.org/multierr"
)

type Balances interface {
	Update(ctx context.Context, client *ethclient.Client) error
	GetBalance(addr common.Address) *big.Float
}

type balances struct {
	mutex       *sync.RWMutex
	ethAddress  common.Address
	tokens      []*erc20.ERC20Token
	balancesMap map[common.Address]*big.Float
}

func NewBalances(ethAddress common.Address, tokens []*erc20.ERC20Token) Balances {
	return &balances{
		mutex:       &sync.RWMutex{},
		ethAddress:  ethAddress,
		tokens:      tokens,
		balancesMap: make(map[common.Address]*big.Float),
	}
}

func (b *balances) Update(ctx context.Context, client *ethclient.Client) error {
	var wg sync.WaitGroup
	wg.Add(len(b.tokens) + 1)
	var merr error

	go func() {
		defer wg.Done()
		balance, err := fetchEthBalance(ctx, b.ethAddress, client)

		b.mutex.Lock()
		defer b.mutex.Unlock()
		if err != nil {
			merr = multierr.Append(merr, err)
		} else {
			b.balancesMap[b.ethAddress] = balance
		}
	}()

	for _, token := range b.tokens {
		go func(token *erc20.ERC20Token) {
			defer wg.Done()
			balance, err := token.FetchBalance(client, b.ethAddress)

			b.mutex.Lock()
			defer b.mutex.Unlock()
			if err != nil {
				merr = multierr.Append(merr, err)
			} else {
				b.balancesMap[token.ContractAddress] = balance
			}
		}(token)
	}

	wg.Wait()
	return merr
}

func (b balances) GetBalance(addr common.Address) *big.Float {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.balancesMap[addr]
}

func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

func fetchEthBalance(ctx context.Context, addr common.Address, client *ethclient.Client) (*big.Float, error) {
	wei, err := client.BalanceAt(ctx, addr, big.NewInt(-1))
	if err != nil {
		return nil, err
	}
	return WeiToEther(wei), nil
}
