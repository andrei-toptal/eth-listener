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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/pinebit/eth-listener/erc20"
)

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	accounts := make(map[common.Address]string)
	for _, acc := range cfg.Accounts {
		accounts[common.HexToAddress(acc.Address)] = acc.Alias
	}

	var tg Telegram
	if cfg.Telegram == nil {
		tg = NewNoopTelegram()
	} else {
		log.Println("Connecting Telegram bot...")
		tg = NewTelegram(cfg.Telegram)
	}

	log.Println("Connecting ETH node...")

	client, err := ethclient.Dial(cfg.EthUrl)
	if err != nil {
		log.Fatal(err)
	}

	tm := NewTokensManager(client)

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
	go HandleTransfersLoop(transfersCh, tm, tg, accounts)

	logTransferSigHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

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
				if _, has := accounts[*tx.To()]; has && tx.Value() != nil {
					transfersCh <- &Transfer{
						Direction: Received,
						From:      msg.From(),
						To:        *tx.To(),
						Value:     *tx.Value(),
						Token:     ETHToken,
					}
				}
			}
			if _, has := accounts[msg.From()]; has && msg.Value() != nil && msg.To() != nil {
				transfersCh <- &Transfer{
					Direction: Sent,
					From:      msg.From(),
					To:        *tx.To(),
					Value:     *tx.Value(),
					Token:     ETHToken,
				}
			}
		}

		filterQuery := ethereum.FilterQuery{}
		filterQuery.FromBlock = block.Number()
		filterQuery.ToBlock = block.Number()
		logs, err := client.FilterLogs(context.Background(), filterQuery)
		if err != nil {
			continue
		}

		for _, logItem := range logs {
			if logItem.Topics[0] != logTransferSigHash || len(logItem.Topics) != 3 {
				continue
			}

			token, err := tm.GetToken(context.Background(), logItem.Address)
			if err != nil {
				continue
			}

			type logTransfer struct {
				Value *big.Int
			}

			var transfer logTransfer
			if err := abi.UnpackIntoInterface(&transfer, "Transfer", logItem.Data); err != nil || transfer.Value == nil {
				continue
			}
			from := common.HexToAddress(logItem.Topics[1].Hex())
			to := common.HexToAddress(logItem.Topics[2].Hex())
			if _, has := accounts[from]; has {
				transfersCh <- &Transfer{
					Direction: Sent,
					From:      from,
					To:        to,
					Value:     *transfer.Value,
					Token:     token,
				}
			}
			if _, has := accounts[to]; has {
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

	tg.Notify("Bot is shutting down...")
	log.Fatalln("Application stopped receiving heads.")
}
