package miner

import (
	"os"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/miner"
)

func main() {
	chain, err := blockchain.ContinueBlockChain()
	if err != nil {
		panic(err)
	}

	walletAddr := os.Getenv("WALLET_ADDR")
	nodeAddr := os.Getenv("NODE_ADDR")
	fullNodeAddr := os.Getenv("FULL_NODE_ADDR")

	m := miner.NewMiner(walletAddr, nodeAddr, fullNodeAddr, chain)
	m.Start()
}
