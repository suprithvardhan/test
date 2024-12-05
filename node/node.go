package node

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	libp2pwebsocket "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
	"libp2p_test/blockchain"
)

type Node struct {
	Host       host.Host
	PubSub     *pubsub.PubSub
	DHT        *dht.IpfsDHT
	Blockchain *blockchain.Blockchain
	Name       string
}

func NewNode(ctx context.Context, bootstrapPeers []peer.AddrInfo, privKey crypto.PrivKey, name string) (*Node, error) {
	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")

	cm, err := connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	host, err := libp2p.New(
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(privKey),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.Transport(libp2pwebsocket.New),
		libp2p.ConnectionManager(cm),
		libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Connect to bootstrap peers
	for _, peerAddr := range bootstrapPeers {
		if err := host.Connect(ctx, peerAddr); err != nil {
			log.Printf("Failed to connect to bootstrap peer %s: %v", peerAddr, err)
		}
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create PubSub: %w", err)
	}

	bc := blockchain.NewBlockchain()

	return &Node{Host: host, PubSub: ps, DHT: kademliaDHT, Blockchain: bc, Name: name}, nil
}

func (n *Node) Start(ctx context.Context) {
	topic, err := n.PubSub.Join("blockchain-sync")
	if err != nil {
		log.Fatalf("Failed to join topic: %v", err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				log.Fatalf("Failed to get next message: %v", err)
			}
			if msg.GetFrom() != n.Host.ID() {
				var transaction blockchain.Transaction
				if err := json.Unmarshal(msg.Data, &transaction); err == nil {
					fmt.Printf("Received transaction from %s: %+v\n", msg.GetFrom(), transaction)
					n.Blockchain.AddTransaction(transaction)
					continue
				}

				var block blockchain.Block
				if err := json.Unmarshal(msg.Data, &block); err == nil {
					fmt.Printf("Received block from %s: %+v\n", msg.GetFrom(), block)
					// Validate and add the block to the blockchain
					if n.Blockchain.IsValid() {
						n.Blockchain.Blocks = append(n.Blockchain.Blocks, block)
					}
					continue
				}

				fmt.Printf("Received unknown message from %s: %s\n", msg.GetFrom(), string(msg.Data))
			}
		}
	}()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			transaction := blockchain.Transaction{ID: fmt.Sprintf("tx%d", rand.Intn(1000)), Amount: rand.Intn(100)}
			data, _ := json.Marshal(transaction)
			if err := topic.Publish(ctx, data); err != nil {
				log.Fatalf("Failed to publish transaction: %v", err)
			}
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Minute) // Simulate mining every 1 minute
			// Randomly decide if this node should mine
			if rand.Intn(2) == 0 {
				n.Blockchain.MineBlock(n.Name)
				block := n.Blockchain.Blocks[len(n.Blockchain.Blocks)-1]
				data, _ := json.Marshal(block)
				if err := topic.Publish(ctx, data); err != nil {
					log.Fatalf("Failed to publish block: %v", err)
				}
			}
		}
	}()

	for _, addr := range n.Host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, n.Host.ID().String())
	}
}
