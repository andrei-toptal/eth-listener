package token

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
)

type TokensDB interface {
	AddToken(token *Token) error
	GetToken(addr common.Address) (*Token, error)
	Close()
}

type tokensDB struct {
	db *leveldb.DB
}

func NewTokensDB(dbPath string) TokensDB {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	tdbPath := path.Join(homeDir, dbPath)
	db, err := leveldb.OpenFile(tdbPath, nil)
	if err != nil {
		log.Panicf("Failed to open TokensDB: %v", err)
	}
	return &tokensDB{
		db: db,
	}
}

func (tdb *tokensDB) Close() {
	if err := tdb.db.Close(); err != nil {
		log.Printf("Failed to close DB: %v", err)
	}
	tdb.db = nil
}

type tokenDTO struct {
	Symbol   string
	Decimals uint8
}

func (tdb tokensDB) AddToken(token *Token) error {
	if tdb.db == nil {
		log.Panicln("Failed to AddToken on closed TokensDB")
	}

	dto := tokenDTO{
		Symbol:   token.Symbol,
		Decimals: token.Decimals,
	}

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(dto); err != nil {
		return err
	}

	return tdb.db.Put(token.Address.Bytes(), buf.Bytes(), nil)
}

func (tdb tokensDB) GetToken(addr common.Address) (*Token, error) {
	if tdb.db == nil {
		log.Panicln("Failed to GetToken on closed TokensDB")
	}

	value, err := tdb.db.Get(addr.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(value)

	var dto tokenDTO
	if err := gob.NewDecoder(buf).Decode(&dto); err != nil {
		return nil, err
	}

	return &Token{
		Address:  addr,
		Symbol:   dto.Symbol,
		Decimals: dto.Decimals,
	}, nil
}
