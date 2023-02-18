package blockchain

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"

	"blockchain/pkg/util"
)

const Difficulty = 12

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevHash,
			pow.block.HashTransactions(),
			util.MustInt64ToHex(int64(nonce)),
			util.MustInt64ToHex(int64(Difficulty)),
		},
		[]byte{},
	)

	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var (
		intHash big.Int
		hash    [32]byte
		nonce   int
	)

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.target) == -1
}
