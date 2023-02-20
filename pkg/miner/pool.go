package miner

import (
	"context"
	"sync"
	"time"

	"blockchain/pkg/blockchain"
)

const (
	// DefaultPackSize is the default number of transactions to pack into a block
	DefaultPackSize = 10
	// DefaultPackTickSec is the default number of seconds to wait before packing transactions
	DefaultPackTickSec = 30
)

// TxPool is a pool of unconfirmed transactions
type TxPool struct {
	unconfirmedTxs []*blockchain.Transaction

	mu sync.Mutex

	packSize int
	packTick time.Duration

	packSignal chan struct{}
}

type TxPoolOpt func(*TxPool)

func NewTxPool(opts ...TxPoolOpt) *TxPool {
	p := &TxPool{
		packSignal: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithPackSize(size int) TxPoolOpt {
	if size <= 0 {
		size = DefaultPackSize
	}
	return func(p *TxPool) {
		p.packSize = size
	}
}

func WithPackTick(tick time.Duration) TxPoolOpt {
	if tick <= 0 {
		tick = DefaultPackTickSec * time.Second
	}
	return func(p *TxPool) {
		p.packTick = tick
	}
}

func (p *TxPool) Add(tx *blockchain.Transaction) {
	p.mu.Lock()
	p.unconfirmedTxs = append(p.unconfirmedTxs, tx)
	p.mu.Unlock()

	if len(p.unconfirmedTxs) >= p.packSize {
		p.packSignal <- struct{}{}
	}
}

func (p *TxPool) GetPack() []*blockchain.Transaction {
	ctx, cancel := context.WithTimeout(context.Background(), p.packTick)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			p.mu.Lock()
			txs := p.unconfirmedTxs
			p.unconfirmedTxs = nil
			return txs
		case <-p.packSignal:
			p.mu.Lock()
			if len(p.unconfirmedTxs) < p.packSize {
				p.mu.Unlock()
				continue
			}
			txs := p.unconfirmedTxs[:p.packSize]
			p.unconfirmedTxs = p.unconfirmedTxs[p.packSize:]
			p.mu.Unlock()
			return txs
		}
	}
}
