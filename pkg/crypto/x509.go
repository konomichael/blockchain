package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func X509EncodePrivate(privateKey *ecdsa.PrivateKey) []byte {
	x509EncodedPriv, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(fmt.Errorf("failed to marshal private key, err: %w", err))
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509EncodedPriv})
}

func X509DecodePrivate(pemEncodedPriv []byte) *ecdsa.PrivateKey {
	block, _ := pem.Decode(pemEncodedPriv)
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		panic(fmt.Errorf("failed to parse private key, err: %w", err))
	}

	return privateKey
}

func X509EncodePublic(publicKey *ecdsa.PublicKey) []byte {
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(fmt.Errorf("failed to marshal public key, err: %w", err))
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
}

func X509DecodePublic(pemEncodedPub []byte) *ecdsa.PublicKey {
	block, _ := pem.Decode(pemEncodedPub)
	x509Encoded := block.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509Encoded)
	if err != nil {
		panic(fmt.Errorf("failed to parse public key, err: %w", err))
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey
}
