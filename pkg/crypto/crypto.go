package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

var curve elliptic.Curve

func init() {
	curve = elliptic.P256()
}

func NewKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}

	return priv, priv.Public().(*ecdsa.PublicKey)
}

func Sign(priv *ecdsa.PrivateKey, data []byte) []byte {
	r, s, err := ecdsa.Sign(rand.Reader, priv, data)
	if err != nil {
		panic(err)
	}

	signature := append(r.Bytes(), s.Bytes()...)

	return signature
}

func Verify(pub, data, signature []byte) bool {
	var (
		xInt, yInt big.Int
		rInt, sInt big.Int
	)
	xInt.SetBytes(pub[:len(pub)/2])
	yInt.SetBytes(pub[len(pub)/2:])

	pubKey := &ecdsa.PublicKey{Curve: curve, X: &xInt, Y: &yInt}

	rInt.SetBytes(signature[:len(signature)/2])
	sInt.SetBytes(signature[len(signature)/2:])

	return ecdsa.Verify(pubKey, data, &rInt, &sInt)
}

func HashPublicKey(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		panic(err)
	}

	return hasher.Sum(nil)
}

func Checksum(payload []byte, len int) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:len]
}
