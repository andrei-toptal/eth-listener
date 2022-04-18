package main

import (
	"context"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/pinebit/eth-listener/erc20"
)

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	aliases := NewAliases(cfg.Aliases)

	log.Println("Connecting Telegram bot...")

	tg := NewTelegram(cfg.Telegram)

	log.Println("Connecting ETH node...")

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

	tokensMap := make(map[common.Address]*erc20.Token)
	filterQuery := ethereum.FilterQuery{}

	for _, ta := range cfg.Tokens {
		contractAddr := common.HexToAddress(ta)
		filterQuery.Addresses = append(filterQuery.Addresses, contractAddr)
		token, err := erc20.FetchERC20Token(contractAddr, client)
		if err != nil {
			log.Fatal(err)
		}
		tokensMap[contractAddr] = token
	}

	addresses := make(map[common.Address]interface{})
	for _, hexAddr := range cfg.Addresses {
		ethAddr := common.HexToAddress(hexAddr)
		addresses[ethAddr] = struct{}{}
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

	transfersCh := make(chan *Transfer, 32)
	go HandleTransfersLoop(transfersCh, client, tg, aliases)

	for header := range headsCh {
		block, err := client.BlockByNumber(context.Background(), header.Number)
		if err != nil {
			continue
		}

		for _, tx := range block.Transactions() {
			msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), big.NewInt(1))
			if err != nil {
				continue
			}
			if tx.To() != nil {
				if _, has := addresses[*tx.To()]; has && tx.Value() != nil {
					transfersCh <- &Transfer{
						Direction: Received,
						From:      msg.From(),
						To:        *tx.To(),
						Value:     *tx.Value(),
						Token:     erc20.ETHToken,
					}
				}
			}
			if _, has := addresses[msg.From()]; has && msg.Value() != nil && msg.To() != nil {
				transfersCh <- &Transfer{
					Direction: Sent,
					From:      msg.From(),
					To:        *tx.To(),
					Value:     *tx.Value(),
					Token:     erc20.ETHToken,
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
				transfersCh <- &Transfer{
					Direction: Sent,
					From:      from,
					To:        to,
					Value:     *transfer.Value,
					Token:     token,
				}
			}
			if _, has := addresses[to]; has {
				transfersCh <- &Transfer{
					Direction: Received,
					From:      from,
					To:        to,
					Value:     *transfer.Value,
					Token:     token,
				}
			}
		}
	}

	log.Fatalln("Application stopped receiving heads.")
}
