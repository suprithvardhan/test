package main

import (
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func main() {
	// Create a new libp2p Host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/4001", // Listen on all interfaces
		),
	)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}

	// Print the host's addresses
	for _, addr := range h.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, h.ID().String())
	}

	// Set up a simple protocol handler
	h.SetStreamHandler(protocol.ID("/chat/1.0.0"), func(s network.Stream) {
		log.Println("Got a new stream!")
	})

	// Keep the process running
	select {}
}
