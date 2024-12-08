package main

import (
	"context"
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ctx := context.Background()

	// Create a new libp2p Host
	h, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}

	// Define the bootstrap node's multiaddress
	bootstrapAddr, err := multiaddr.NewMultiaddr("/ip4/192.168.71.57/tcp/4001/p2p/12D3KooWAkY8S99HCQ3mUVuVdG65reMeWKH1rbq3rD7bpqzyhV1n")
	if err != nil {
		log.Fatalf("Failed to parse multiaddr: %v", err)
	}

	// Extract the peer ID from the multiaddress
	peerInfo, err := peer.AddrInfoFromP2pAddr(bootstrapAddr)
	if err != nil {
		log.Fatalf("Failed to extract peer info: %v", err)
	}

	// Add the bootstrap node to the peerstore
	h.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)

	// Connect to the bootstrap node
	if err := h.Connect(ctx, *peerInfo); err != nil {
		log.Fatalf("Failed to connect to bootstrap node: %v", err)
	}

	fmt.Println("Connected to bootstrap node!")

	// Set up a simple protocol handler
	h.SetStreamHandler(protocol.ID("/chat/1.0.0"), func(s network.Stream) {
		log.Println("Got a new stream!")
	})

	// Keep the process running
	select {}
} 