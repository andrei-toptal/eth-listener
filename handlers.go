package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/andrei-toptal/eth-listener/token"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func handleTransfer(transfer *Transfer, app *App, ctx context.Context) {
	value := transfer.Token.RenderValue(&transfer.Value)

	getBalanceStr := func(addr common.Address) string {
		balanceStr := "N/A"
		balance, err := app.tokensManager.FetchBalance(ctx, transfer.Token, addr)
		if err == nil {
			balanceStr = transfer.Token.RenderValue(balance)
		} else {
			log.Printf("Failed to fetch balance for %s: %v", transfer.Token.Symbol, err)
		}
		return balanceStr
	}

	var msg string
	switch transfer.Direction {
	case Sent:
		msg = fmt.Sprintf("%s sent %s to %s, new balance: %s",
			app.accounts.Lookup(transfer.From),
			value,
			app.accounts.Lookup(transfer.To),
			getBalanceStr(transfer.From))

	case Received:
		msg = fmt.Sprintf("%s received %s from %s, new balance: %s",
			app.accounts.Lookup(transfer.To),
			value,
			app.accounts.Lookup(transfer.From),
			getBalanceStr(transfer.To))
	}

	log.Println(msg)
	app.telegram.Notify(msg)
}

func handleHeader(ctx context.Context, header *types.Header, transfersCh chan<- *Transfer, app *App) {
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
					Token:     token.ETHToken,
				}
			}
		}
		if _, has := app.accounts[msg.From()]; has && msg.Value() != nil && msg.To() != nil {
			transfersCh <- &Transfer{
				Direction: Sent,
				From:      msg.From(),
				To:        *tx.To(),
				Value:     *tx.Value(),
				Token:     token.ETHToken,
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
		if len(logItem.Topics) != 3 || logItem.Topics[0] != LogTransferSigHash {
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
