package types

import (
	"errors"
	"fmt"
	"math"

	"github.com/dominant-strategies/go-quai/common"
)

const (

	// UTXOVersion is the current latest supported transaction version.
	UTXOVersion = 1

	// MaxTxInSequenceNum is the maximum sequence number the sequence field
	// of a transaction input can be.
	MaxTxInSequenceNum uint32 = 0xffffffff

	// MaxPrevOutIndex is the maximum index the index field of a previous
	// outpoint can be.
	MaxPrevOutIndex uint32 = 0xffffffff

	// SatoshiPerBitcent is the number of satoshi in one bitcoin cent.
	SatoshiPerBitcent = 1e6

	// SatoshiPerBitcoin is the number of satoshi in one bitcoin (1 BTC).
	SatoshiPerBitcoin = 1e8

	// MaxSatoshi is the maximum transaction amount allowed in satoshi.
	MaxSatoshi = 21e6 * SatoshiPerBitcoin
)

// TxIn defines a Qi transaction input
type TxIn struct {
	PreviousOutPoint OutPoint
	PubKey           []byte
}

// OutPoint defines a Qi data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash  common.Hash
	Index uint32
}

// NewTxIn returns a new Qi transaction input
func NewTxIn(prevOut *OutPoint, pubkey []byte) *TxIn {
	return &TxIn{
		PreviousOutPoint: *prevOut,
		PubKey:           pubkey,
	}
}

// TxOut defines a bitcoin transaction output
type TxOut struct {
	Value   uint64
	Address []byte
}

// NewTxOut returns a new Qi transaction output
func NewTxOut(value uint64, address []byte) *TxOut {
	return &TxOut{
		Value:   value,
		Address: address,
	}
}
