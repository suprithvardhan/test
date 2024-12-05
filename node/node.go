package node

import (
	"context"
	"fmt"
	"log"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	libp2pwebsocket "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	Host   host.Host
	PubSub *pubsub.PubSub
	DHT    *dht.IpfsDHT
}

func NewNode(ctx context.Context, bootstrapPeers []peer.AddrInfo) (*Node, error) {
	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")

	cm, err := connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	// Define your static relay addresses here
	staticRelays := []peer.AddrInfo{
		// Add your relay peer addresses here
	}

	host, err := libp2p.New(
		libp2p.ListenAddrs(listenAddr),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.Transport(libp2pwebsocket.New),
		libp2p.ConnectionManager(cm),
		libp2p.EnableAutoRelayWithStaticRelays(staticRelays),
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

	for _, peerAddr := range bootstrapPeers {
		if err := host.Connect(ctx, peerAddr); err != nil {
			log.Printf("Failed to connect to bootstrap peer %s: %v", peerAddr, err)
		}
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create PubSub: %w", err)
	}

	return &Node{Host: host, PubSub: ps, DHT: kademliaDHT}, nil
}

func (n *Node) Start(ctx context.Context) {
	topic, err := n.PubSub.Join("libp2p-test")
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
				fmt.Printf("Received message from %s: %s\n", msg.GetFrom(), string(msg.Data))
			}
		}
	}()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := topic.Publish(ctx, []byte("hi")); err != nil {
				log.Fatalf("Failed to publish message: %v", err)
			}
		}
	}()

	for _, addr := range n.Host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, n.Host.ID().String())
	}
}
