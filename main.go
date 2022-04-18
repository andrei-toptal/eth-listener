package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/pinebit/eth-listener/config"
	"github.com/pinebit/eth-listener/erc20"
	"github.com/pinebit/eth-listener/telegram"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	tg := telegram.NewTelegram(cfg.Telegram)

	client, err := ethclient.Dial(cfg.EthUrl)
	if err != nil {
		log.Fatal(err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Working chain ID: %s", chainID.String())

	log.Println("Fetching tokens...")

	var tokens []*erc20.ERC20Token
	tokensMap := make(map[common.Address]*erc20.ERC20Token)
	filterQuery := ethereum.FilterQuery{}

	for _, ta := range cfg.Tokens {
		contractAddr := common.HexToAddress(ta)
		filterQuery.Addresses = append(filterQuery.Addresses, contractAddr)
		token, err := erc20.FetchERC20Token(contractAddr, client)
		if err != nil {
			log.Fatal(err)
		}
		tokens = append(tokens, token)
		tokensMap[contractAddr] = token
	}

	log.Println("Fetching balances...")

	addresses := make(map[common.Address]interface{})
	for _, hexAddr := range cfg.Addresses {
		ethAddr := common.HexToAddress(hexAddr)
		addresses[ethAddr] = struct{}{}
		balances := NewBalances(ethAddr, tokens)
		balances.Update(context.Background(), client)

		log.Printf("ADDRESS: %s", ethAddr.String())
		log.Printf("- ETH balance: %s", balances.GetBalance(ethAddr).String())

		for _, token := range tokens {
			log.Printf("- %s balance: %s", token.Symbol, balances.GetBalance(token.ContractAddress).String())
		}
	}

	log.Println("Watching for transactions...")

	abi, err := abi.JSON(strings.NewReader(string(erc20.ERC20ABI)))
	if err != nil {
		log.Fatal(err)
	}

	headsCh := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headsCh)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	for header := range headsCh {
		block, err := client.BlockByNumber(context.Background(), header.Number)
		if err != nil {
			continue
		}

		for _, tx := range block.Transactions() {
			if tx.To() != nil {
				if _, has := addresses[*tx.To()]; has && tx.Value() != nil {
					msg := fmt.Sprintf("You received: %s ETH", WeiToEther(tx.Value()).String())
					log.Println(msg)
					tg.Notify(msg)
				}
			}
			msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), big.NewInt(1))
			if err == nil {
				if _, has := addresses[msg.From()]; has && msg.Value() != nil && msg.To() != nil {
					msg := fmt.Sprintf("You sent %s ETH to %s", WeiToEther(msg.Value()).String(), msg.To().String())
					log.Println(msg)
					tg.Notify(msg)
				}
			}
		}

		filterQuery.FromBlock = block.Number()
		filterQuery.ToBlock = block.Number()
		logs, err := client.FilterLogs(context.Background(), filterQuery)
		if err != nil {
			continue
		}

		for _, logItem := range logs {
			if logItem.Topics[0] != erc20.LogTransferSigHash || len(logItem.Topics) != 3 {
				continue
			}

			token, has := tokensMap[logItem.Address]
			if !has {
				log.Printf("ERROR: missing token for %s", logItem.Address)
				continue
			}

			type logTransfer struct {
				Value *big.Int
			}

			var transfer logTransfer
			if err := abi.UnpackIntoInterface(&transfer, "Transfer", logItem.Data); err != nil || transfer.Value == nil {
				log.Printf("ERROR: failed to interpret Transfer event: %v", err)
				continue
			}
			from := common.HexToAddress(logItem.Topics[1].Hex())
			to := common.HexToAddress(logItem.Topics[2].Hex())
			if _, has := addresses[from]; has {
				log.Printf("You sent %s %s to %s", token.ConvertValue(transfer.Value).String(), token.Symbol, to)
			}
			if _, has := addresses[to]; has {
				log.Printf("You received %s %s from %s", token.ConvertValue(transfer.Value).String(), token.Symbol, from)
			}
		}
	}
}
