package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Define the bootstrap peer address
	bootstrapPeerAddr, err := multiaddr.NewMultiaddr("/ip4/49.204.109.180/tcp/4001/p2p/12D3KooWM8QPTxjVzEic7C74UpzpRmeUXufvjDdqqhAhdBnDfMBF")
	if err != nil {
		log.Fatalf("Failed to create multiaddr: %v", err)
	}

	bootstrapPeerInfo, err := peer.AddrInfoFromP2pAddr(bootstrapPeerAddr)
	if err != nil {
		log.Fatalf("Failed to create peer info: %v", err)
	}

	// Pass the bootstrap peer to the NewNode function
	bootstrapPeers := []peer.AddrInfo{*bootstrapPeerInfo}

	n, err := node.NewNode(ctx, bootstrapPeers)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	n.Start(ctx)

	// Wait for a signal to gracefully shut down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Shutting down...")
}
