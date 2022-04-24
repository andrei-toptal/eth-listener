package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Direction int

const (
	Sent Direction = iota
	Received
)

type Transfer struct {
	Direction Direction
	From      common.Address
	To        common.Address
	Value     big.Int
	Token     *Token
}

func HandleTransfersLoop(ctx context.Context, transfersCh <-chan *Transfer, app *App) {
	for {
		select {
		case <-ctx.Done():
			return
		case transfer := <-transfersCh:
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
	}
}
