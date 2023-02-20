package p2p

import (
	"bufio"
	"context"
	"crypto/ecdsa"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

func MakeHost(nodeAddr string, privKey *ecdsa.PrivateKey) (host.Host, error) {
	sourceMultiAddr, err := multiaddr.NewMultiaddr(nodeAddr)
	if err != nil {
		return nil, err
	}

	key, _, err := crypto.ECDSAKeyPairFromKey(privKey)
	if err != nil {
		return nil, err
	}

	return libp2p.New(libp2p.ListenAddrs(sourceMultiAddr), libp2p.Identity(key))
}

func StartPeerAndConnect(ctx context.Context, h host.Host, dest string, pid protocol.ID) (*bufio.ReadWriter, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		maddr, err := multiaddr.NewMultiaddr(dest)
		if err != nil {
			return nil, err
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, err
		}
		h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
		// Create a new stream, this stream will be handled by handleStream on the other side.
		s, err := h.NewStream(ctx, info.ID, pid)
		if err != nil {
			return nil, err
		}

		// Create a buffer stream for non blocking read and write.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		return rw, nil
	}
}
