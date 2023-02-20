package blockchain

import (
	"bytes"
	"encoding/hex"

	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix = []byte("utxo-")
	prefixLen  = len(utxoPrefix)
)

const collectSize = 100000

type UTXOSet struct {
	*BlockChain
}

func NewUTXOSet(chain *BlockChain) *UTXOSet {
	return &UTXOSet{chain}
}

func (u *UTXOSet) Reindex() error {
	if err := u.DeleteByPrefix(utxoPrefix); err != nil {
		return err
	}

	UTXOs := u.BlockChain.FindUTXOs()

	return u.database.Update(func(txn *badger.Txn) error {
		for txID, outs := range UTXOs {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			key = append(utxoPrefix, key...)
			val, err := outs.Serialize()
			if err != nil {
				return err
			}

			return txn.Set(key, val)
		}
		return nil
	})
}

func (u *UTXOSet) Update(block *Block) error {
	return u.database.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}

					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					if err != nil {
						return err
					}

					var outs TxOutputs
					err = item.Value(func(val []byte) error {
						return outs.Deserialize(val)
					})
					if err != nil {
						return err
					}

					for outIdx, out := range outs {
						if outIdx != in.Out {
							updatedOuts = append(updatedOuts, out)
						}
					}

					if len(updatedOuts) == 0 {
						if err = txn.Delete(inID); err != nil {
							return err
						}
						continue
					}

					encoded, err := updatedOuts.Serialize()
					if err != nil {
						return err
					}
					if err = txn.Set(inID, encoded); err != nil {
						return err
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs = append(newOutputs, out)
			}

			txID := append(utxoPrefix, tx.ID...)
			encoded, err := newOutputs.Serialize()
			if err != nil {
				return err
			}
			if err = txn.Set(txID, encoded); err != nil {
				return err
			}
		}
		return nil
	})
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) error {
	deleteKeys := func(keysForDelete [][]byte) error {
		return u.database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		})
	}

	return u.database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++

			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			return deleteKeys(keysForDelete)
		}

		return nil
	})
}

func (u *UTXOSet) CountTransactions() (int, error) {
	counter := 0

	err := u.database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	return counter, err
}

func (u *UTXOSet) FindUTXOs(pubKeyHash []byte) (*TxOutputs, error) {
	UTXOs := &TxOutputs{}

	err := u.database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			var outs TxOutputs
			err := item.Value(func(val []byte) error {
				return outs.Deserialize(val)
			})

			if err != nil {
				return err
			}

			for _, out := range outs {
				if out.IsLockedWithKey(pubKeyHash) {
					*UTXOs = append(*UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return UTXOs, nil
}

func (u *UTXOSet) FindSpendableUTXOs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOuts := make(map[string][]int) // txID -> []outIdx
	accumulated := 0

	err := u.database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			var outs TxOutputs
			err := item.Value(func(val []byte) error {
				return outs.Deserialize(val)
			})
			if err != nil {
				return err
			}

			txID := hex.EncodeToString(bytes.TrimPrefix(item.Key(), utxoPrefix))

			for outIdx, out := range outs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return accumulated, unspentOuts, nil
}
