package main

import (
	"math/big"

	"github.com/andrei-toptal/eth-listener/token"
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
	Token     *token.Token
}
