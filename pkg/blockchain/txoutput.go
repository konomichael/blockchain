package blockchain

import (
	"bytes"
	"fmt"

	"blockchain/pkg/wallet"
)

// TxOutput represents a transaction output
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

func NewTXOutput(value int, address string) *TxOutput {
	out := &TxOutput{
		Value: value,
	}

	out.Lock(address)

	return out
}

func (out *TxOutput) Lock(address string) {
	pubKeyHash, err := wallet.PubKeyHashFromAddress(address)
	if err != nil {
		panic(fmt.Sprintf("failed to get public key hash from address: %s", err))
	}
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
