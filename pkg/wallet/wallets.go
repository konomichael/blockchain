package wallet

import (
	"bytes"
	"errors"
	"os"
	"sync"

	"blockchain/pkg/crypto"
	"blockchain/pkg/util"
)

const (
	walletDir = "./tmp/wallets/"
)

var once sync.Once

func init() {
	once.Do(func() {
		err := os.MkdirAll(walletDir, 0755)
		if errors.Is(err, os.ErrExist) {
			return
		}
		if err != nil {
			panic(err)
		}
	})
}

func CreateWallet() (string, error) {
	wallet := NewWallet()

	address := string(wallet.Address())
	if err := wallet.saveToFile(address); err != nil {
		return "", err
	}

	return address, nil
}

func GetWallet(address string) (*Wallet, error) {
	w := &Wallet{}
	if err := w.loadFromFile(address); err != nil {
		return nil, err
	}
	return w, nil
}

func PubKeyHashFromAddress(address string) ([]byte, error) {
	pubKeyHash := util.Base58Decode([]byte(address))
	if len(pubKeyHash) < checksumLength {
		return nil, errors.New("invalid address")
	}

	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]

	versionedPubKeyHash := pubKeyHash[:len(pubKeyHash)-checksumLength]
	targetChecksum := crypto.Checksum(versionedPubKeyHash, checksumLength)
	if bytes.Compare(actualChecksum, targetChecksum) != 0 {
		return nil, errors.New("invalid address")
	}

	return pubKeyHash[1 : len(pubKeyHash)-checksumLength], nil
}

func GetAllAddresses() []string {
	files, err := os.ReadDir(walletDir)
	if err != nil {
		return nil
	}

	var addresses []string

	for _, file := range files {
		if file.IsDir() {
			addresses = append(addresses, file.Name())
		}
	}

	return addresses
}
