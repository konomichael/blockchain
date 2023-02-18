package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/gob"
	"encoding/pem"
	"fmt"

	"github.com/mr-tron/base58"
)

func Int64ToHex(num int64) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, num); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func MustInt64ToHex(num int64) []byte {
	buf, err := Int64ToHex(num)
	if err != nil {
		panic(fmt.Errorf("failed to convert int64 to hex, err: %w", err))
	}

	return buf
}

func GobEncode(data any) ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GobDecode(raw []byte, data any) error {
	decoder := gob.NewDecoder(bytes.NewReader(raw))

	return decoder.Decode(data)
}

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input))
	if err != nil {
		panic(fmt.Errorf("failed to decode base58, err: %w", err))
	}

	return decode
}

func X509Encode(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) ([]byte, []byte) {
	x509EncodedPriv, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(fmt.Errorf("failed to marshal private key, err: %w", err))
	}
	pemEncodedPriv := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509EncodedPriv})

	x509EncodedPub, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(fmt.Errorf("failed to marshal public key, err: %w", err))
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return pemEncodedPriv, pemEncodedPub
}

func X509Decode(pemEncodedPriv, pemEncodedPub []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	block, _ := pem.Decode(pemEncodedPriv)
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		panic(fmt.Errorf("failed to parse private key, err: %w", err))
	}

	block, _ = pem.Decode(pemEncodedPub)
	x509Encoded = block.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509Encoded)
	if err != nil {
		panic(fmt.Errorf("failed to parse public key, err: %w", err))
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return privateKey, publicKey
}
