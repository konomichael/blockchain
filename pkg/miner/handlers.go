package miner

import (
	"bufio"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/p2p"
)

func HandlerTx(m *Miner) p2p.Handler {
	return func(data []byte, rw *bufio.ReadWriter) {
		tx := m.rawTxPool.Get().(*blockchain.Transaction)
		if err := tx.Deserialize(data); err != nil {
			m.rawTxPool.Put(tx)
			return
		}

		if !m.chain.VerifyTransaction(tx) {
			m.rawTxPool.Put(tx)
			return
		}

		m.pool.Add(tx)
	}
}
