package main

import (
	"crypto/rand"
	"io"
	mrand "math/rand"

	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
)

// func(...) error later :>
func addPeer(h Host, dest string) error {

	addr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		return err
	}

	peer, err := peer.AddrInfoFromP2pAddr(addr)

	if err != nil {
		return err
	}

	h.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
	return nil
}

// returns the host multiaddr
func getHostInfo(p Host) (string, error) {
	peer2 := peer.AddrInfo{
		ID:    p.ID(),
		Addrs: p.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peer2)
	return addrs[0].String(), err
}

// builds a new Host
func makeHost(seed int64, port int) (Host, error) {
	var r io.Reader

	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA /*ECDSA*/, 2048, r)

	if err != nil {
		return nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))

	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}