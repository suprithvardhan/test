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
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"libp2p_test/blockchain"
)

type Node struct {
	Host       host.Host
	PubSub     *pubsub.PubSub
	DHT        *dht.IpfsDHT
	Blockchain *blockchain.Blockchain
	Name       string
	Sub        *pubsub.Subscription
}

func NewNode(ctx context.Context, bootstrapPeers []peer.AddrInfo, privKey crypto.PrivKey, name string) (*Node, error) {
	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")

	cm, err := connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	peerSource := func(ctx context.Context, num int) <-chan peer.AddrInfo {
		ch := make(chan peer.AddrInfo)
		go func() {
			defer close(ch)
			for _, p := range bootstrapPeers {
				select {
				case ch <- p:
				case <-ctx.Done():
					return
				}
			}
		}()
		return ch
	}

	host, err := libp2p.New(
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(privKey),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelayWithPeerSource(peerSource),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ConnectionManager(cm),
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

	pubsubService, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub service: %w", err)
	}

	topic, err := pubsubService.Join("blockchain")
	if err != nil {
		return nil, fmt.Errorf("failed to join pubsub topic: %w", err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to pubsub topic: %w", err)
	}

	node := &Node{
		Host:       host,
		PubSub:     pubsubService,
		DHT:        kademliaDHT,
		Blockchain: blockchain.NewBlockchain(),
		Name:       name,
		Sub:        sub,
	}

	// Connect to bootstrap peers
	for _, peerInfo := range bootstrapPeers {
		if err := host.Connect(ctx, peerInfo); err != nil {
			log.Printf("Failed to connect to bootstrap peer %s: %v", peerInfo.ID, err)
		} else {
			log.Printf("Connected to bootstrap peer %s", peerInfo.ID)
		}
	}

	return node, nil
}

func (n *Node) Start(ctx context.Context) {
	// Handle incoming messages
	go func() {
		for {
			msg, err := n.Sub.Next(ctx)
			if err != nil {
				log.Fatalf("Failed to read message: %v", err)
			}

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
	}()

	// Simulate transaction generation
	go func() {
		for {
			time.Sleep(10 * time.Second)
			transaction := blockchain.Transaction{ID: fmt.Sprintf("tx%d", rand.Intn(1000)), Amount: rand.Intn(100)}
			data, _ := json.Marshal(transaction)
			if err := n.PubSub.Publish("blockchain", data); err != nil {
				log.Fatalf("Failed to publish transaction: %v", err)
			}
		}
	}()

	// Simulate mining
	go func() {
		for {
			time.Sleep(1 * time.Minute) // Simulate mining every 1 minute
			// Randomly decide if this node should mine
			if rand.Intn(2) == 0 {
				n.Blockchain.MineBlock(n.Name)
				block := n.Blockchain.Blocks[len(n.Blockchain.Blocks)-1]
				data, _ := json.Marshal(block)
				if err := n.PubSub.Publish("blockchain", data); err != nil {
					log.Fatalf("Failed to publish block: %v", err)
				}
			}
		}
	}()

	for _, addr := range n.Host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, n.Host.ID().String())
	}
}
