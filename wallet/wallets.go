package wallet

import (
	"errors"
	"os"
	"sync"
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
