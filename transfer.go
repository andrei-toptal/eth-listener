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

func HandleTransfersLoop(transfersCh <-chan *Transfer, tm TokensManager, tg Telegram, accounts map[common.Address]string) {
	for transfer := range transfersCh {
		value := transfer.Token.RenderValue(&transfer.Value)

		var msg string
		switch transfer.Direction {
		case Sent:
			balanceStr := "N/A"
			balance, err := tm.FetchBalance(context.Background(), transfer.Token, transfer.From)
			if err == nil {
				balanceStr = transfer.Token.RenderValue(balance)
			}
			msg = fmt.Sprintf("%s sent %s to %s, new balance: %s",
				lookup(accounts, transfer.From),
				value,
				lookup(accounts, transfer.To),
				balanceStr)

		case Received:
			balanceStr := "N/A"
			balance, err := tm.FetchBalance(context.Background(), transfer.Token, transfer.To)
			if err == nil {
				balanceStr = transfer.Token.RenderValue(balance)
			}
			msg = fmt.Sprintf("%s received %s from %s, new balance: %s",
				lookup(accounts, transfer.To),
				value,
				lookup(accounts, transfer.From),
				balanceStr)
		}

		log.Println(msg)
		tg.Notify(msg)
	}
}

func lookup(accounts map[common.Address]string, addr common.Address) string {
	alias, ok := accounts[addr]
	if !ok {
		return addr.String()
	}
	return alias
}
