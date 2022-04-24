package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func handleHeader(ctx context.Context, header *types.Header, transfersCh chan *Transfer, app *App) {
	block, err := app.client.BlockByNumber(ctx, header.Number)
	if err != nil {
		return
	}

	for _, tx := range block.Transactions() {
		msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), big.NewInt(1))
		if err != nil {
			continue
		}
		if tx.To() != nil {
			if _, has := app.accounts[*tx.To()]; has && tx.Value() != nil {
				transfersCh <- &Transfer{
					Direction: Received,
					From:      msg.From(),
					To:        *tx.To(),
					Value:     *tx.Value(),
					Token:     ETHToken,
				}
			}
		}
		if _, has := app.accounts[msg.From()]; has && msg.Value() != nil && msg.To() != nil {
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
	logs, err := app.client.FilterLogs(ctx, filterQuery)
	if err != nil {
		return
	}

	for _, logItem := range logs {
		if logItem.Topics[0] != LogTransferSigHash || len(logItem.Topics) != 3 {
			continue
		}

		token, err := app.tokensManager.GetToken(ctx, logItem.Address)
		if err != nil {
			log.Printf("Skipping log for non-ERC20 contract: %v", err)
			continue
		}

		type logTransfer struct {
			Value *big.Int
		}

		var transfer logTransfer
		if err := ERC20ABI.UnpackIntoInterface(&transfer, "Transfer", logItem.Data); err != nil || transfer.Value == nil {
			continue
		}
		from := common.HexToAddress(logItem.Topics[1].Hex())
		to := common.HexToAddress(logItem.Topics[2].Hex())
		if _, has := app.accounts[from]; has {
			transfersCh <- &Transfer{
				Direction: Sent,
				From:      from,
				To:        to,
				Value:     *transfer.Value,
				Token:     token,
			}
		}
		if _, has := app.accounts[to]; has {
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

func main() {
	log.Println("Starting eth-listener application...")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		sig := <-ch
		log.Printf("Shutting down due to %s", sig)
		cancel()
	}()

	app, err := WireApp(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	defer app.tokensDB.Close()

	log.Println("Watching for transactions...")

	headsCh := make(chan *types.Header)
	sub, err := app.client.SubscribeNewHead(ctx, headsCh)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	transfersCh := make(chan *Transfer, TransfersChBuffer)
	go HandleTransfersLoop(ctx, transfersCh, app)

mainLoop:
	for {
		select {
		case <-ctx.Done():
			break mainLoop
		case header := <-headsCh:
			handleHeader(ctx, header, transfersCh, app)
		}
	}

	app.telegram.Notify("Bot is shutting down...")
	log.Printf("Application stopped.")
}
