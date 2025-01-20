package server

import (
	"context"
	"fmt"
	"nakamoto-blockchain/logger"
	"nakamoto-blockchain/proto/gen"
	"google.golang.org/grpc/peer"
)

type IncomingCommunicator struct {
	gen.UnimplementedIncomingCommunicatorServiceServer
	Node *BlockchainServer
}

func (s *IncomingCommunicator) SubmitBlock(ctx context.Context, block *gen.BlockWithHashes) (*gen.BlockResponse, error) {
	// EXTRACT PEER IP ADDRESS
	peerAddr := "unknownPeer"
    if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
        peerAddr = p.Addr.String()
    }
	logger.DebugLogger.Println("[SubmitBlock] Called with block hash:", block.Block.Hash, "from:", peerAddr)

	if s.Node.PeerManager.IsBlacklisted(peerAddr) {
		errMsg := fmt.Sprintf("peer %s is blacklisted", peerAddr)
		logger.DebugLogger.Println("[SubmitBlock] Rejected block from blacklisted peer:", peerAddr)
		return &gen.BlockResponse{Accepted: false, Error: errMsg}, nil
	}

	blk := ConvertGrpcToBlock(block.Block)
	hashes := block.Last_100Hashes
	res, err := s.Node.HandleBlockSubmission(blk, &hashes, peerAddr)

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
		logger.DebugLogger.Printf("[SubmitBlock] Block %s failed: %v", block.Block.Hash, err)
	}
	return &gen.BlockResponse{Accepted: res, Error: errorMsg}, err
}

func (s *IncomingCommunicator) SubmitTransaction(ctx context.Context, tx *gen.Transaction) (*gen.TxResponse, error) {

	transaction := ConvertGrpcToTransaction(tx)
	res, err := s.Node.HandleTransactionSubmission(transaction)

	// Create response with error handling
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
		logger.DebugLogger.Printf("[SendTransaction] Transaction %s failed: %v", tx.Hash, err)
	}

	return &gen.TxResponse{Accepted: res, Error: errorMsg}, err
}

func (s *IncomingCommunicator) GetBlockByHash(ctx context.Context, req *gen.BlockRequest) (*gen.Block, error) {
	logger.InfoLogger.Println("[GetBlock] Called with hash:", req.Hash)
	block := s.Node.Blockchain.GetBlockByHash(req.Hash)

	if block == nil {
		logger.InfoLogger.Println("[GetBlock] Block not found for hash:", req.Hash)
		return nil, fmt.Errorf("block not found")
	}

	logger.InfoLogger.Println("[GetBlock] Block found for hash:", req.Hash)
	return ConvertBlockToGrpc(block), nil
}

func (s *IncomingCommunicator) GetTransactionStatus(ctx context.Context, req *gen.TransactionStatusRequest) (*gen.TransactionStatusResponse, error) {
    txHash := req.Hash
    k := req.K

	depth := int32(s.Node.Blockchain.GetTransactionDepth(txHash))
    isConfirmed := false

	if depth >= k {
		isConfirmed = true
	}

    if !isConfirmed {
        return &gen.TransactionStatusResponse{
            Confirmed: false,
            Error:     fmt.Sprintf("Transaction %s not confirmed with k=%d. Has depth of %d.", txHash, k, depth),
        }, nil
    }
    return &gen.TransactionStatusResponse{
        Confirmed: true,
        Error:     "",
    }, nil
}
