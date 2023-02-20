package miner

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/vrecan/death/v3"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/p2p"
	"blockchain/pkg/wallet"
)

var handlers map[string]p2p.Handler

type Miner struct {
	walletAddr string // btc address
	chain      *blockchain.BlockChain
	utxoSet    *blockchain.UTXOSet

	nodeAddr string

	fullNodeAddr string
	peers        map[string]*bufio.ReadWriter // peers[addr] = rw
	host         host.Host

	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey

	pool      *TxPool
	rawTxPool sync.Pool

	ctx    context.Context
	cancel context.CancelFunc
}

func init() {
	handlers = make(map[string]p2p.Handler)
}

func NewMiner(walletAddr, fullNodeAddr string, port int, chain *blockchain.BlockChain) *Miner {
	wa, err := wallet.GetWallet(walletAddr)
	if err != nil {
		panic(err)
	}

	nodeAddr := fmt.Sprintf("/ip4/localhost/tcp/%d", port)
	_host, err := p2p.MakeHost(nodeAddr, wa.PrivateKey)
	if err != nil {
		panic(err)
	}
	_host.SetStreamHandler("/miner/1.0.0", p2p.MakeStreamHandler(handlers))

	ctx, cancel := context.WithCancel(context.Background())

	return &Miner{
		walletAddr:   walletAddr,
		chain:        chain,
		utxoSet:      blockchain.NewUTXOSet(chain),
		fullNodeAddr: fullNodeAddr,
		host:         _host,
		nodeAddr:     nodeAddr,
		privateKey:   wa.PrivateKey,
		publicKey:    wa.PublicKey,
		pool:         NewTxPool(),
		rawTxPool: sync.Pool{
			New: func() interface{} {
				return &blockchain.Transaction{}
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (m *Miner) Start() {
	rw, err := p2p.StartPeerAndConnect(m.ctx, m.host, m.fullNodeAddr, "/miner/1.0.0")
	if err != nil {
		m.chain.Close()
		panic(err)
	}

	go m.gracefulShutdown()

	go m.mine()

	go m.heartbeat(rw)

	m.start()
}

func (m *Miner) mine() {
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			txs := m.pool.GetPack()
			if len(txs) == 0 {
				continue
			}

			cbTx, err := blockchain.CoinbaseTx(m.walletAddr, "")
			if err != nil {
				m.cancel()
				return
			}
			txs = append(txs, cbTx)

			block, err := m.chain.MineBlock(txs)
			if err != nil {
				m.cancel()
				return
			}

			_ = m.utxoSet.Reindex()
			blkData, _ := block.Serialize()
			m.broadcastCommand("block", blkData)

			m.putBackTxs(txs[:len(txs)-1])
		}
	}
}

func (m *Miner) gracefulShutdown() {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		m.shutdown()
	})

}

func (m *Miner) start() {
	select {
	case <-m.ctx.Done():
		return
	}
}

func (m *Miner) shutdown() {
	if m.chain != nil {
		m.chain.Close()
	}
	m.cancel()
}

func (m *Miner) heartbeat(rw *bufio.ReadWriter) {
	var count int
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-time.After(5 * time.Second):
			if err := p2p.SendCommand(rw, "heartbeat", []byte(m.nodeAddr)); err != nil {
				count++
				if count > 30 {
					m.cancel()
					return
				}
			} else {
				count = 0
			}
		}
	}
}

func (m *Miner) broadcastCommand(cmd string, data []byte) {
	for addr, rw := range m.peers {
		if addr != m.nodeAddr {
			go p2p.SendCommand(rw, cmd, data)
		}
	}
}

func (m *Miner) putBackTxs(txs []*blockchain.Transaction) {
	for _, tx := range txs {
		m.rawTxPool.Put(tx)
	}
}
