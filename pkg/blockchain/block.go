package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"blockchain/pkg/util"
)

type Block struct {
	Transactions []*Transaction
	Hash         []byte
	PrevHash     []byte
	Nonce        int
}

func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Transactions: txs,
		PrevHash:     prevHash,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (b *Block) Serialize() []byte {
	encoded, err := util.GobEncode(b)
	if err != nil {
		panic(fmt.Errorf("error while serializing block: %w", err))
	}

	return encoded
}

func Deserialize(data []byte) *Block {
	var block Block
	if err := util.GobDecode(data, &block); err != nil {
		panic(fmt.Errorf("error while deserializing block: %w", err))
	}

	return &block
}
