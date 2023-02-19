package util

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
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
