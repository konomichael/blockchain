package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"blockchain/pkg/crypto"
	"blockchain/pkg/util"
	"blockchain/pkg/wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func NewTransaction(from, to string, amount int, chain *BlockChain) (*Transaction, error) {
	var (
		inputs  []TxInput
		outputs []TxOutput
	)

	w, err := wallet.GetWallet(from)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorTxCreateFailed, err)
	}

	pubKeyBytes := w.PublicKeyBytes()
	pubKeyHash := crypto.HashPublicKey(pubKeyBytes)

	acc, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		return nil, fmt.Errorf("%w: not enough funds", ErrorTxCreateFailed)
	}

	for txId, outs := range validOutputs {
		txID, err := hex.DecodeString(txId)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrorTxCreateFailed, err)
		}

		for _, out := range outs {
			in := NewTxInput(txID, out, pubKeyBytes)
			inputs = append(inputs, *in)
		}
	}

	out := NewTXOutput(amount, to)
	outputs = append(outputs, *out)

	if acc > amount {
		outputs = append(outputs, TxOutput{
			Value:      acc - amount,
			PubKeyHash: pubKeyHash,
		})
	}

	tx := &Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	if err := tx.SetID(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorTxCreateFailed, err)
	}

	if err := chain.SignTransaction(tx, w.PrivateKey); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorTxCreateFailed, err)
	}

	return tx, nil
}

func CoinbaseTx(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txIn := NewTxInput(nil, -1, nil)
	txOut := NewTXOutput(100, to)
	tx := &Transaction{
		Inputs:  []TxInput{*txIn},
		Outputs: []TxOutput{*txOut},
	}
	if err := tx.SetID(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorTxCreateFailed, err)
	}

	return tx, nil
}

func (tx *Transaction) SetID() error {
	tx.ID = nil

	encode, err := util.GobEncode(tx)
	if err != nil {
		return fmt.Errorf("error while encoding transaction, %w", err)
	}

	hash := sha256.Sum256(encode)
	tx.ID = hash[:]

	return nil
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]*Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			return fmt.Errorf("%w: previous transaction is not correct", ErrorTxSignFailed)
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].PubKey = prevTX.Outputs[in.Out].PubKeyHash

		if err := txCopy.SetID(); err != nil {
			return fmt.Errorf("%w: %s", ErrorTxSignFailed, err)
		}

		txCopy.Inputs[inID].PubKey = nil

		tx.Inputs[inID].Signature = crypto.Sign(privKey, txCopy.ID)
	}

	return nil
}

// Verify checks if the transaction is valid
func (tx *Transaction) Verify(prevTXs map[string]*Transaction) bool {
	// Coinbase type transactions don't have inputs
	if tx.IsCoinbase() {
		return true
	}

	// Check if the transaction inputs are valid
	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			return false
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, in := range tx.Inputs {
		prevTXs := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].PubKey = prevTXs.Outputs[in.Out].PubKeyHash

		if err := txCopy.SetID(); err != nil {
			return false
		}

		txCopy.Inputs[inID].PubKey = nil

		if !crypto.Verify(in.PubKey, txCopy.ID, in.Signature) {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() *Transaction {
	var (
		inputs  []TxInput
		outputs []TxOutput
	)

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{
			ID:  in.ID,
			Out: in.Out,
		})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      out.Value,
			PubKeyHash: out.PubKeyHash,
		})
	}

	return &Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, in := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("Input %d:", i))
		lines = append(lines, fmt.Sprintf("  ID:      %x", in.ID))
		lines = append(lines, fmt.Sprintf("  Out:     %d", in.Out))
		lines = append(lines, fmt.Sprintf("  Signature: %x", in.Signature))
		lines = append(lines, fmt.Sprintf("  PubKey: %x", in.PubKey))
	}

	for i, out := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("Output %d:", i))
		lines = append(lines, fmt.Sprintf("  Value:  %d", out.Value))
		lines = append(lines, fmt.Sprintf("  PubKeyHash: %x", out.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
