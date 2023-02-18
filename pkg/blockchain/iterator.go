package blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

type Iterator struct {
	Database    *badger.DB
	CurrentHash []byte
}

func (iter *Iterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		if err != nil {
			return fmt.Errorf("error while getting last hash: %w", err)
		}

		return item.Value(func(val []byte) error {
			block = Deserialize(val)

			return nil
		})
	})
	if err != nil {
		log.Panic(err)
	}

	iter.CurrentHash = block.PrevHash

	return block
}

func (iter *Iterator) HasNext() bool {
	// Genesis' hash is empty
	return len(iter.CurrentHash) > 0
}
