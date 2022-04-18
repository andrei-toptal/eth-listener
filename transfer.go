package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pinebit/eth-listener/erc20"
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
	Token     *erc20.Token
}

func HandleTransfersLoop(transfersCh <-chan *Transfer, client *ethclient.Client, tg Telegram, aliases Aliases) {
	for transfer := range transfersCh {
		value := transfer.Token.ConvertValue(&transfer.Value).String()
		if value == "0" {
			value = "~0"
		}

		var msg string
		switch transfer.Direction {
		case Sent:
			balance := transfer.Token.FetchBalance(client, transfer.From)
			msg = fmt.Sprintf("%s sent %s %s to %s, new balance: %s %s",
				aliases.lookup(transfer.From),
				value,
				transfer.Token.Symbol,
				aliases.lookup(transfer.To),
				balance,
				transfer.Token.Symbol)

		case Received:
			balance := transfer.Token.FetchBalance(client, transfer.To)
			msg = fmt.Sprintf("%s received %s %s from %s, new balance: %s %s",
				aliases.lookup(transfer.To),
				value,
				transfer.Token.Symbol,
				aliases.lookup(transfer.From),
				balance,
				transfer.Token.Symbol)
		}

		log.Println(msg)
		tg.Notify(msg)
	}
}

func (a Aliases) lookup(addr common.Address) string {
	alias, ok := a[addr]
	if !ok {
		return addr.String()
	}
	return alias
}
