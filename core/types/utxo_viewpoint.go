package types

import (
	"github.com/dominant-strategies/go-quai/common"
)

// UtxoEntry houses details about an individual transaction output in a utxo
// view such as whether or not it was contained in a coinbase tx, the height of
// the block that contains the tx, whether or not it is spent, its public key
// script, and how much it pays.
type UtxoEntry struct {
	Denomination uint8
	Address      []byte // The address of the output holder.
}

// Clone returns a shallow copy of the utxo entry.
func (entry *UtxoEntry) Clone() *UtxoEntry {
	if entry == nil {
		return nil
	}

	return &UtxoEntry{
		Denomination: entry.Denomination,
		Address:      entry.Address,
	}
}

// NewUtxoEntry returns a new UtxoEntry built from the arguments.
func NewUtxoEntry(txOut *TxOut) *UtxoEntry {
	return &UtxoEntry{
		Denomination: txOut.Denomination,
		Address:      txOut.Address,
	}
}

// ProtoEncode serializes the UtxoEntry into protobuf bytes.
func (entry *UtxoEntry) ProtoEncode() *ProtoUtxoEntry {
	pbEntry := &ProtoUtxoEntry{
		Denomination: uint32(entry.Denomination),
		Address:      entry.Address,
	}

	return pbEntry
}

func (entry *UtxoEntry) ProtoDecode(pbEntry *ProtoUtxoEntry) error {
	entry.Denomination = uint8(pbEntry.Denomination)
	entry.Address = pbEntry.Address
	return nil
}

type AddressUtxos struct {
	Address common.Address
	Utxos   []*UtxoEntry
}

type UtxosMap map[common.Address][]*UtxoEntry
