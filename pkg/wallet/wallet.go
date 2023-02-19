package wallet

import (
	"crypto/ecdsa"
	"os"

	"blockchain/pkg/crypto"
	"blockchain/pkg/util"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

func NewWallet() *Wallet {
	priv, pub := crypto.NewKeyPair()

	return &Wallet{priv, pub}
}

func (w *Wallet) Address() []byte {
	pubHashBytes := w.PublicKeyBytes()
	pubHash := crypto.HashPublicKey(pubHashBytes)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := crypto.Checksum(versionedHash, checksumLength)

	fullHash := append(versionedHash, checksum...)
	address := util.Base58Encode(fullHash)

	return address
}

func (w *Wallet) PublicKeyBytes() []byte {
	return append(w.PublicKey.X.Bytes(), w.PublicKey.Y.Bytes()...)
}

func (w *Wallet) saveToFile(address string) error {
	folder := walletDir + address
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err := os.MkdirAll(folder, 0755)
		if err != nil {
			return err
		}
		encodedPrivKey := crypto.X509EncodePrivate(w.PrivateKey)

		if err := os.WriteFile(folder+"/private.pem", encodedPrivKey, 0644); err != nil {
			return err
		}

		encodedPubKey := crypto.X509EncodePublic(w.PublicKey)
		if err := os.WriteFile(folder+"/public.pem", encodedPubKey, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (w *Wallet) loadFromFile(address string) error {
	folder := walletDir + address
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return err
	}

	priv, err := os.ReadFile(folder + "/private.pem")
	privKey := crypto.X509DecodePrivate(priv)
	if err != nil {
		return err
	}

	pub, err := os.ReadFile(folder + "/public.pem")
	pubKey := crypto.X509DecodePublic(pub)
	if err != nil {
		return err
	}

	w.PrivateKey, w.PublicKey = privKey, pubKey

	return nil
}
