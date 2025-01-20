package main

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"

	"google.golang.org/grpc"

	"nakamoto-blockchain/internal/blockchain"
	"nakamoto-blockchain/internal/server"
	"nakamoto-blockchain/logger"
	"nakamoto-blockchain/proto/gen"
)

func main() {
	logger.Init()
	if len(os.Args) < 5 {
		logger.ErrorLogger.Fatal("[Server] Usage: go run main.go <initial_UTXOs> <httpServer_port> <grpc_port> <mode> <peer1> <peer2> ...")
	}

	utxoFile := os.Args[1]
	httpPort := os.Args[2]
	grpcPort := os.Args[3]
	mode, _ := strconv.Atoi(os.Args[4])
	peerAddresses := os.Args[5:]

	logger.InfoLogger.Printf("[Server] Starting miner with gRPC port: %s and peers: %v", grpcPort, peerAddresses)

	utxos := loadUTXOs(utxoFile)
	peerManager := server.NewPeerManager()
	peerManager.AddPeers(peerAddresses)

	outgoingComms := server.OutgoingCommunicator{PeerManager: peerManager}
	blockchainServer := server.NewBlockchainServer(outgoingComms, peerManager, utxos, mode)
	incomingComms := &server.IncomingCommunicator{Node: blockchainServer}

	// Start mining immediately
	logger.InfoLogger.Println("[Server] Starting mining immediately")
	if err := blockchainServer.MineBlocks(); err != nil {
		logger.ErrorLogger.Fatalf("[Server] Failed to start mining: %v", err)
	}

	go startHTTPServer(httpPort, blockchainServer)
	go startGRPCServer(grpcPort, incomingComms)

	logger.InfoLogger.Println("[Server] Server is running and mining...")
	select {} // Block forever instead of using wait group
}

func loadUTXOs(utxoFile string) []blockchain.UTXO {
	logger.InfoLogger.Printf("[Server] Loading UTXOs from file: %s", utxoFile)

	file, err := os.Open(utxoFile)
	if err != nil {
		logger.ErrorLogger.Fatalf("[Server] Failed to open UTXO file: %v", err)
	}
	defer file.Close()

	var entries []struct {
		PublicKey string            `json:"public_key"`
		UTXOs     []blockchain.UTXO `json:"utxos"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		logger.ErrorLogger.Fatalf("[Server] Failed to decode UTXO file: %v", err)
	}

	var utxos []blockchain.UTXO
	for _, entry := range entries {
		utxos = append(utxos, entry.UTXOs...)
	}

	logger.InfoLogger.Printf("[Server] Successfully loaded %d UTXOs.", len(utxos))
	return utxos
}

func startHTTPServer(httpPort string, blockchainServer *server.BlockchainServer) {
	http.HandleFunc("/addpeers", handleAddPeers(blockchainServer))
	http.HandleFunc("/mineblocks", handleMineBlocks(blockchainServer))
	http.HandleFunc("/stopmining", handleStopMining(blockchainServer))

	logger.InfoLogger.Printf("[HTTP Server] Listening on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
		logger.ErrorLogger.Fatalf("[HTTP Server] Failed to start: %v", err)
	}
}

func handleAddPeers(blockchainServer *server.BlockchainServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		type PeersRequest struct {
			Addresses []string `json:"addresses"`
		}

		var req PeersRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		blockchainServer.PeerManager.AddPeers(req.Addresses)

		logger.InfoLogger.Printf("[HTTP Server] Added peers: %v", req.Addresses)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"peers added"}`))
	}
}

func handleMineBlocks(blockchainServer *server.BlockchainServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		logger.InfoLogger.Println("[HTTP Server] MineBlocks endpoint called.")
		err := blockchainServer.MineBlocks()
		if err != nil {
			http.Error(w, "Failed to start mining: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{"status": "mining started"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func handleStopMining(blockchainServer *server.BlockchainServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		logger.InfoLogger.Println("[HTTP Server] StopMining endpoint called.")
		err := blockchainServer.StopMining()
		if err != nil {
			http.Error(w, "Failed to stop mining: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{"status": "mining stopped"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func startGRPCServer(grpcPort string, incomingComms *server.IncomingCommunicator) {
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.ErrorLogger.Fatalf("[GRPC Server] Failed to listen on %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	gen.RegisterIncomingCommunicatorServiceServer(grpcServer, incomingComms)

	logger.InfoLogger.Printf("[GRPC Server] Listening on port %s", grpcPort)
	if err := grpcServer.Serve(listener); err != nil {
		logger.ErrorLogger.Fatalf("[GRPC Server] Failed to start: %v", err)
	}
}
