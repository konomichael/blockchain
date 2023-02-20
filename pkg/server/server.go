package server

import (
	"net"

	"blockchain/pkg/blockchain"
)

type Server struct {
	minerAddr string
	nodeAddr  string

	protocol string

	peers   []string
	clients []string

	chain *blockchain.BlockChain
}

func NewServer(minerAddr, nodeAddr string) *Server {
	return &Server{
		minerAddr: minerAddr,
		nodeAddr:  nodeAddr,
		protocol:  "tcp",
		peers:     []string{"localhost:3000"},
	}
}

func (s *Server) Start() {
	ln, err := net.Listen(s.protocol, s.nodeAddr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	chain, err := blockchain.ContinueBlockChain()
	if err != nil {
		panic(err)
	}
	defer chain.Close()
	s.chain = chain

	go s.gracefulShutdown()

	if s.nodeAddr != s.peers[0] {
		s.syncVersion()
	}

}
