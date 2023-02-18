package blockchain

import (
	"blockchain/pkg/util"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxInput struct {
	Sig string
	ID  []byte
	Out int
}

type TxOutput struct {
	PubKey string
	Value  int
}

func NewTransaction(from, to string, amount int, chain *BlockChain) (*Transaction, error) {
	var (
		inputs  []TxInput
		outputs []TxOutput
	)

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)
	if acc < amount {
		return nil, errors.New("error: not enough funds")
	}

	for txId, outs := range validOutputs {
		txID, err := hex.DecodeString(txId)
		if err != nil {
			return nil, fmt.Errorf("error while decoding transaction ID, %w", err)
		}

		for _, out := range outs {
			inputs = append(inputs, TxInput{
				ID:  txID,
				Sig: from,
				Out: out,
			})
		}
	}

	outputs = append(outputs, TxOutput{
		Value:  amount,
		PubKey: to,
	})

	if acc > amount {
		outputs = append(outputs, TxOutput{
			Value:  acc - amount,
			PubKey: from,
		})
	}

	tx := &Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	if err := tx.SetID(); err != nil {
		return nil, err
	}

	return tx, nil
}

func CoinbaseTx(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txIn := TxInput{
		ID:  []byte{},
		Sig: data,
		Out: -1,
	}
	txOut := TxOutput{
		PubKey: to,
		Value:  100,
	}
	tx := &Transaction{
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}
	if err := tx.SetID(); err != nil {
		return nil, err
	}

	return tx, nil
}

func (tx *Transaction) SetID() error {
	encoded, err := util.GobEncode(tx)
	if err != nil {
		return fmt.Errorf("error while encoding transaction: %s", err)
	}

	hash := sha256.Sum256(encoded)
	tx.ID = hash[:]
	return nil
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
