package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	database *badger.DB
	lastHash []byte
}

func InitBlockChain(address string) (*BlockChain, error) {
	if dbExists() {
		return nil, ErrorBCExists
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("error while opening database: %w", err)
	}

	var lastHash []byte

	err = db.Update(func(txn *badger.Txn) error {
		cbtx, err1 := CoinbaseTx(address, genesisData)
		if err1 != nil {
			return fmt.Errorf("error while create coinbase transaction: %w", err1)
		}

		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")

		if err1 := txn.Set(genesis.Hash, genesis.Serialize()); err1 != nil {
			return fmt.Errorf("error while setting genesis block: %w", err1)
		}

		if err1 := txn.Set([]byte("lh"), genesis.Hash); err1 != nil {
			return fmt.Errorf("error while setting last hash: %w", err1)
		}

		lastHash = genesis.Hash

		return nil
	})
	if err != nil {
		_ = db.Close()
		_ = os.RemoveAll(dbPath)
		return nil, fmt.Errorf("error while updating blockchain: %w", err)
	}

	return &BlockChain{db, lastHash}, nil
}

func ContinueBlockChain() (*BlockChain, error) {
	if !dbExists() {
		return nil, errors.New("no existing blockchain found. Create one first")
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("error while opening database: %w", err)
	}

	var lastHash []byte
	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return fmt.Errorf("error while getting last hash: %w", err)
		}

		return item.Value(func(val []byte) error {
			lastHash = append(lastHash, val...)

			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error while getting last hash: %w", err)
	}

	return &BlockChain{db, lastHash}, nil
}

func (bc *BlockChain) AddBlock(transactions []*Transaction) error {
	var lastHash []byte

	err := bc.database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return fmt.Errorf("error while getting last hash: %w", err)
		}

		return item.Value(func(val []byte) error {
			lastHash = append(lastHash, val...)

			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("error while getting last hash: %w", err)
	}

	block := CreateBlock(transactions, lastHash)

	err = bc.database.Update(func(txn *badger.Txn) error {
		if err1 := txn.Set(block.Hash, block.Serialize()); err1 != nil {
			return fmt.Errorf("error while setting new block: %w", err1)
		}

		if err1 := txn.Set([]byte("lh"), block.Hash); err1 != nil {
			return fmt.Errorf("error while setting last hash: %w", err1)
		}

		bc.lastHash = block.Hash

		return nil
	})

	if err != nil {
		return fmt.Errorf("error while updating blockchain: %w", err)
	}

	return nil
}

func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []*Transaction {
	var unspentTxs []*Transaction
	spentTXOs := make(map[string][]int)

	iter := bc.Iterator()
	for iter.HasNext() {
		block := iter.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if txos, ok := spentTXOs[txID]; ok {
					for _, spentOut := range txos {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}
	}
	return unspentTxs
}

func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func (bc *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *BlockChain) FindTransaction(ID []byte) (*Transaction, error) {
	iter := bc.Iterator()
	for iter.HasNext() {
		block := iter.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return tx, nil
			}
		}
	}
	return nil, ErrorTxNotFound
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey *ecdsa.PrivateKey) error {
	prevTXs := make(map[string]*Transaction)

	for _, in := range tx.Inputs {
		if prevTx, err := bc.FindTransaction(in.ID); err != nil {
			return err
		} else {
			prevTXs[hex.EncodeToString(prevTx.ID)] = prevTx
		}
	}
	return tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]*Transaction)

	for _, in := range tx.Inputs {
		if prevTx, err := bc.FindTransaction(in.ID); err != nil {
			return false
		} else {
			prevTXs[hex.EncodeToString(prevTx.ID)] = prevTx
		}
	}
	return tx.Verify(prevTXs)
}

func (bc *BlockChain) Close() {
	_ = bc.database.Close()
}

func (bc *BlockChain) Iterator() *Iterator {
	return &Iterator{
		Database:    bc.database,
		CurrentHash: bc.lastHash,
	}
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
