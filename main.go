package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"libp2p_test/node"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const keyFilePath = "node_key"

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go <node-name>")
	}
	nodeName := os.Args[1]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load or generate a private key
	privKey, err := loadOrCreatePrivateKey(keyFilePath)
	if err != nil {
		log.Fatalf("Failed to load or create private key: %v", err)
	}

	// Define the bootstrap peer address
	bootstrapPeerAddr, err := multiaddr.NewMultiaddr("/ip4/49.204.109.180/tcp/4001/p2p/12D3KooWAcngjjMCPKketoTETkEB6w9fbdm3fjUn4gkeMNuTiskp")
	if err != nil {
		log.Fatalf("Failed to create multiaddr: %v", err)
	}

	bootstrapPeerInfo, err := peer.AddrInfoFromP2pAddr(bootstrapPeerAddr)
	if err != nil {
		log.Fatalf("Failed to create peer info: %v", err)
	}

	// Pass the bootstrap peer to the NewNode function
	bootstrapPeers := []peer.AddrInfo{*bootstrapPeerInfo}

	n, err := node.NewNode(ctx, bootstrapPeers, privKey, nodeName)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	n.Start(ctx)

	// Log successful connection to the bootstrap node
	log.Println("Successfully connected to the bootstrap node")

	// Check for connected peers
	go func() {
		for {
			time.Sleep(10 * time.Second)
			peers := n.Host.Network().Peers()
			log.Printf("Connected peers: %v\n", peers)
		}
	}()

	// Print the peer ID and addresses
	fmt.Printf("Node started with ID: %s\n", n.Host.ID().String())
	for _, addr := range n.Host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, n.Host.ID().String())
	}

	// Wait for a signal to gracefully shut down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Shutting down...")
}

func loadOrCreatePrivateKey(filePath string) (crypto.PrivKey, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Generate a new private key
		privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key: %w", err)
		}

		// Save the private key to a file
		keyBytes, err := crypto.MarshalPrivateKey(privKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key: %w", err)
		}

		if err := os.WriteFile(filePath, keyBytes, 0600); err != nil {
			return nil, fmt.Errorf("failed to write private key to file: %w", err)
		}

		return privKey, nil
	}

	// Load the private key from the file
	keyBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from file: %w", err)
	}

	privKey, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	return privKey, nil
}
