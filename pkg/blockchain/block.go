package blockchain

import (
	"time"

	"blockchain/pkg/util"
)

type Block struct {
	Timestamp    int64
	Transactions []*Transaction
	Hash         []byte
	PrevHash     []byte
	Nonce        int
	Height       int
}

func CreateBlock(txs []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{
		Timestamp:    time.Now().Unix(),
		Transactions: txs,
		PrevHash:     prevHash,
		Height:       height,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	tree := NewMerkleTree(txHashes)

	return tree.RootNode.Data
}

func (b *Block) Serialize() ([]byte, error) {
	return util.GobEncode(b)
}

func (b *Block) Deserialize(data []byte) error {
	return util.GobDecode(data, b)
}
