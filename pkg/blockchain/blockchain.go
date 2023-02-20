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

		encodedGenesis, err1 := genesis.Serialize()
		if err1 != nil {
			return fmt.Errorf("error while serializing genesis block: %w", err1)
		}
		if err1 := txn.Set(genesis.Hash, encodedGenesis); err1 != nil {
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

	return &BlockChain{database: db, lastHash: lastHash}, nil
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

	return &BlockChain{database: db, lastHash: lastHash}, nil
}

func (bc *BlockChain) AddBlock(block *Block) error {
	return bc.database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		item, err := txn.Get(bc.lastHash)
		if err != nil {
			return fmt.Errorf("error while getting last block: %w", err)
		}

		lastBlock := &Block{}
		if err := item.Value(func(val []byte) error {
			return lastBlock.Deserialize(val)
		}); err != nil {
			return fmt.Errorf("error while deserializing last block: %w", err)
		}

		if block.Height != lastBlock.Height+1 {
			return fmt.Errorf("error while adding block: %w, lastBlock's height: %d, block's height: %d", ErrorBlkHeightInvalid, lastBlock.Height, block.Height)
		}

		if block.PrevHash == nil || !bytes.Equal(block.PrevHash, lastBlock.Hash) {
			return fmt.Errorf("error while adding block: %w, lastBlock's hash: %x, block's prevHash: %x", ErrorBlkPrevHashInvalid, lastBlock.Hash, block.PrevHash)
		}

		blockData, err := block.Serialize()
		if err != nil {
			return fmt.Errorf("error while serializing block: %w", err)
		}

		if err := txn.Set(block.Hash, blockData); err != nil {
			return fmt.Errorf("error while setting block: %w", err)
		}

		if err := txn.Set([]byte("lh"), block.Hash); err != nil {
			return fmt.Errorf("error while setting last hash: %w", err)
		}

		bc.lastHash = block.Hash

		return nil
	})
}

func (bc *BlockChain) MineBlock(transactions []*Transaction) (*Block, error) {
	var lastHeight int
	lastHash := bc.lastHash

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			return nil, fmt.Errorf("while verifying transaction %s: %w", hex.EncodeToString(tx.ID), ErrorTxInvalid)
		}
	}

	err := bc.database.View(func(txn *badger.Txn) error {
		item, err1 := txn.Get(lastHash)
		if err1 != nil {
			return fmt.Errorf("error while getting last block: %w", err1)
		}

		if err1 = item.Value(func(val []byte) error {
			lastBlock := &Block{}
			if err2 := lastBlock.Deserialize(val); err2 != nil {
				return fmt.Errorf("error while deserializing last block: %w", err2)
			}

			lastHeight = lastBlock.Height
			return nil
		}); err1 != nil {
			return fmt.Errorf("error while getting last block's height: %w", err1)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while getting last hash: %w", err)
	}

	block := CreateBlock(transactions, lastHash, lastHeight+1)

	err = bc.AddBlock(block)
	if err != nil {
		return nil, fmt.Errorf("error while adding block: %w", err)
	}

	return block, nil
}

func (bc *BlockChain) FindUTXOs() map[string]*TxOutputs {
	UTXOs := make(map[string]*TxOutputs)
	spentTXOs := make(map[string]struct{})

	iter := bc.Iterator()
	for iter.HasNext() {
		block := iter.Next()
		txs := block.Transactions
		for _, tx := range txs {
			txID := hex.EncodeToString(tx.ID)
		Output:
			for outIdx, out := range tx.Outputs {
				k := fmt.Sprintf("%s-%d", txID, outIdx)
				if _, ok := spentTXOs[k]; ok {
					continue Output
				}
				outs := UTXOs[txID]
				if outs == nil {
					outs = &TxOutputs{}
					UTXOs[txID] = outs
				}
				*outs = append(*outs, out)
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					k := fmt.Sprintf("%s-%d", hex.EncodeToString(in.ID), in.Out)
					spentTXOs[k] = struct{}{}
				}
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
			if bytes.Equal(tx.ID, ID) {
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
	if tx.IsCoinbase() {
		return true
	}

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

func (bc *BlockChain) GetBaseHeight() (int, error) {
	var lastBlock *Block

	err := bc.database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bc.lastHash)
		if err != nil {
			return fmt.Errorf("error while getting last block: %w", err)
		}

		err = item.Value(func(val []byte) error {
			return lastBlock.Deserialize(val)
		})
		if err != nil {
			return fmt.Errorf("error while getting last block: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("error while getting last block: %w", err)
	}

	return lastBlock.Height, nil
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
