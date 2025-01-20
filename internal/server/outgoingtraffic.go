package server

import (
	"context"
	"nakamoto-blockchain/internal/blockchain"
	"nakamoto-blockchain/logger"
	"nakamoto-blockchain/proto/gen"
	"strings"
)

type OutgoingCommunicator struct {
	PeerManager *PeerManager
}

func (s *OutgoingCommunicator) BroadcastTransaction(tx *blockchain.Transaction) {
	logger.DebugLogger.Println("[BroadcastTransaction] Called with transaction hash:", tx.Hash)

	for _, client := range s.PeerManager.ListPeerClients() {
		_, err := client.SubmitTransaction(context.Background(), ConvertTransactionToGrpc(tx))
		if err != nil {
			if strings.Contains(err.Error(), "already") {
				logger.DebugLogger.Printf("[BroadcastTransaction] Error broadcasting transaction: %v", err)
			} else {
				logger.WarnLogger.Printf("[BroadcastTransaction] Error broadcasting transaction: %v", err)
			}
		}
	}

	logger.DebugLogger.Println("[BroadcastTransaction] Completed for hash:", tx.Hash)
}

func (s *OutgoingCommunicator) BroadcastBlock(block *blockchain.Block, hashes []string) {
	logger.DebugLogger.Println("[BroadcastBlock] Called with block hash:", block.Hash)

	for _, client := range s.PeerManager.ListPeerClients() {
		_, err := client.SubmitBlock(context.Background(), &gen.BlockWithHashes{
			Block:          ConvertBlockToGrpc(block),
			Last_100Hashes: hashes,
		})
		if err != nil {
			logger.DebugLogger.Printf("[BroadcastBlock] Error broadcasting block: %v", err)
		}
	}

	logger.DebugLogger.Println("[BroadcastBlock] Completed for hash:", block.Hash)
}

func (s *OutgoingCommunicator) RequestBlockByHash(hash string) *blockchain.Block {
	logger.InfoLogger.Println("[RequestBlockByHash] Called with hash:", hash)

	for _, client := range s.PeerManager.ListPeerClients() {
		blockResponse, err := client.GetBlockByHash(context.Background(), &gen.BlockRequest{Hash: hash})
		if err != nil {
			logger.ErrorLogger.Printf("[RequestBlockByHash] Error requesting block by hash: %v", err)
			continue
		}
		return ConvertGrpcToBlock(blockResponse)
	}

	logger.ErrorLogger.Println("[RequestBlockByHash] Block not found for hash:", hash)
	return nil
}
