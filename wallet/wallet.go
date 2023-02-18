package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"os"

	"golang.org/x/crypto/ripemd160"

	"blockchain/pkg/util"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

func (w *Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := CheckSum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := util.Base58Encode(fullHash)

	return address
}

func NewKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	curve := elliptic.P256()

	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}

	return priv, priv.Public().(*ecdsa.PublicKey)
}

func NewWallet() *Wallet {
	priv, pub := NewKeyPair()

	encodedPriv, encodedPub := util.X509Encode(priv, pub)

	return &Wallet{encodedPriv, encodedPub}
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		panic(err)
	}

	return hasher.Sum(nil)
}

func CheckSum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

func (w *Wallet) saveToFile(address string) error {
	folder := walletDir + address
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err := os.MkdirAll(folder, 0755)
		if err != nil {
			return err
		}
		if err := os.WriteFile(folder+"/private.pem", w.PrivateKey, 0644); err != nil {
			return err
		}
		if err := os.WriteFile(folder+"/public.pem", w.PublicKey, 0644); err != nil {
			return err
		}
	}

	return nil
}

func loadWalletFromFile(dir string) (*Wallet, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	priv, err := os.ReadFile(dir + "/private.pem")
	if err != nil {
		return nil, err
	}
	pub, err := os.ReadFile(dir + "/public.pem")
	if err != nil {
		return nil, err
	}

	return &Wallet{priv, pub}, nil
}
