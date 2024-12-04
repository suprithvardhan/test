package main

import (
	"context"
	"libp2p_test/node"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ctx := context.Background()
	var bootstrapPeers []peer.AddrInfo

	if len(os.Args) > 1 {
		bootstrapAddr, err := multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			log.Fatalf("Failed to create bootstrap address: %v", err)
		}
		bootstrapPeer, err := peer.AddrInfoFromP2pAddr(bootstrapAddr)
		if err != nil {
			log.Fatalf("Failed to create bootstrap peer: %v", err)
		}
		bootstrapPeers = append(bootstrapPeers, *bootstrapPeer)
	}

	n, err := node.NewNode(ctx, bootstrapPeers)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}
	n.Start(ctx)

	select {}
}
