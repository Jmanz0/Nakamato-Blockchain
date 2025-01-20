package server

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"nakamoto-blockchain/logger"
	"nakamoto-blockchain/proto/gen"

	"google.golang.org/grpc"
)

const DefaultGRPCPort = 50051

type PeerManager struct {
	mu           sync.Mutex
	peerClients  map[string]*grpc.ClientConn
	blacklisted  map[string]bool
	invalidCount map[string]int
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peerClients:  make(map[string]*grpc.ClientConn),
		blacklisted:  make(map[string]bool),
		invalidCount: make(map[string]int),
	}
}

func (pm *PeerManager) AddPeer(address string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	originalAddress := address
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%s:%d", address, DefaultGRPCPort)
		logger.InfoLogger.Printf("Appended default port to peer address: %s -> %s", originalAddress, address)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil || host == "" || port == "" {
		logger.ErrorLogger.Printf("Invalid peer address format after appending port: %s", address)
		return fmt.Errorf("invalid peer address format: %s", address)
	}

	if _, exists := pm.peerClients[address]; exists {
		logger.DebugLogger.Printf("Peer already exists: %s", address)
		return nil
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		logger.ErrorLogger.Printf("Failed to connect to peer %s: %v", address, err)
		return fmt.Errorf("failed to connect to peer %s: %v", address, err)
	}

	pm.peerClients[address] = conn
	logger.InfoLogger.Printf("Peer added and connected: %s", address)
	return nil
}

func (pm *PeerManager) AddPeers(addresses []string) {
	for _, address := range addresses {
		splitAddresses := strings.Split(address, ",")
		for _, addr := range splitAddresses {
			addr = strings.TrimSpace(addr)
			if addr == "" {
				continue
			}
			if err := pm.AddPeer(addr); err != nil {
				logger.ErrorLogger.Printf("Failed to add peer %s: %v", addr, err)
			}
		}
	}
}

func (pm *PeerManager) RemovePeer(address string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	conn, exists := pm.peerClients[address]
	if !exists {
		logger.DebugLogger.Printf("Peer not found: %s", address)
		return
	}

	if err := conn.Close(); err != nil {
		logger.ErrorLogger.Printf("Error closing connection for peer %s: %v", address, err)
	} else {
		logger.InfoLogger.Printf("Peer removed and connection closed: %s", address)
	}
	delete(pm.peerClients, address)
}

func (pm *PeerManager) ListPeers() []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	peers := make([]string, 0, len(pm.peerClients))
	for address := range pm.peerClients {
		peers = append(peers, address)
	}
	return peers
}

func (pm *PeerManager) ListPeerClients() []gen.IncomingCommunicatorServiceClient {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	clients := make([]gen.IncomingCommunicatorServiceClient, 0, len(pm.peerClients))
	for _, conn := range pm.peerClients {
		clients = append(clients, gen.NewIncomingCommunicatorServiceClient(conn))
	}
	return clients
}

func (pm *PeerManager) IsBlacklisted(address string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.blacklisted[address]
}

func (pm *PeerManager) IncrementInvalidCount(address string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.invalidCount[address]++
	if pm.invalidCount[address] >= 3 {
		pm.blacklisted[address] = true
		logger.WarnLogger.Printf("[PeerManager] Blacklisted peer: %s", address)
		// pm.RemovePeer(address) // Potentially add
	}
}
