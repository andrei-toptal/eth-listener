package token

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Token struct {
	Address  common.Address
	Symbol   string
	Decimals uint8
}

// To unify transfers processing we mimic ETH to be a token.
// Address is set to 0 and not used for this "token".
var ETHToken = &Token{
	Address:  common.HexToAddress("0x0"),
	Symbol:   "ETH",
	Decimals: 18,
}

func (t Token) RenderValue(value *big.Int) string {
	val := new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(math.Pow10(int(t.Decimals))))
	return fmt.Sprintf("%s %s", val.String(), t.Symbol)
}
