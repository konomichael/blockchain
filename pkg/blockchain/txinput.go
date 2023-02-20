package blockchain

import (
	"bytes"

	"blockchain/pkg/crypto"
)

// TxInput represents a transaction input
type TxInput struct {
	// ID of the transaction that contains the output we're referencing
	ID []byte
	// Index of the output we're referencing
	Out       int
	Signature []byte
	PubKey    []byte
}

func NewTxInput(id []byte, out int, pubKey []byte) *TxInput {
	return &TxInput{
		ID:     id,
		Out:    out,
		PubKey: pubKey,
	}
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := crypto.HashPublicKey(in.PubKey)

	return bytes.Equal(lockingHash, pubKeyHash)
}
