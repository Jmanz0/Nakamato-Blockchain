package server

import (
	"context"
	"fmt"
	"math/rand"
	"nakamoto-blockchain/internal/blockchain"
	"nakamoto-blockchain/logger"
	"time"
)

type BlockchainServer struct {
	Blockchain  *blockchain.Blockchain
	TxPool      *blockchain.TransactionPool
	cancelFunc  context.CancelFunc
	mining      bool
	Comms       OutgoingCommunicator
	PeerManager *PeerManager
	mode        int
}

func NewBlockchainServer(comms OutgoingCommunicator, peerManager *PeerManager, initialUTXOs []blockchain.UTXO, mode int) *BlockchainServer {
	return &BlockchainServer{
		Blockchain:  blockchain.NewBlockchain(initialUTXOs),
		TxPool:      blockchain.NewTransactionPool(),
		Comms:       comms,
		PeerManager: peerManager,
		mode:        mode,
	}
}

func (s *BlockchainServer) HandleTransactionSubmission(tx *blockchain.Transaction) (bool, error) {
	logger.DebugLogger.Printf("Transaction received: %s", tx.Hash)

	if !tx.Verify() {
		logger.DebugLogger.Printf("Invalid transaction: %s", tx.Hash)
		return false, fmt.Errorf("invalid transaction")
	}

	if !s.TxPool.HasTransaction(tx.Hash) && !s.Blockchain.HasTransaction(tx.Hash) {
		s.TxPool.AddTransaction(*tx)
		s.Comms.BroadcastTransaction(tx)
		logger.DebugLogger.Printf("Transaction processed: %s", tx.Hash)
		return true, nil
	}
	logger.DebugLogger.Printf("Duplicate transaction: %s", tx.Hash)
	return false, fmt.Errorf("transaction already in the pool or blockchain")
}

func (s *BlockchainServer) HandleBlockSubmission(block *blockchain.Block, hashes *[]string, peerAddr string) (bool, error) {
	// 1) If peer is blacklisted, reject immediately
	if s.PeerManager.IsBlacklisted(peerAddr) {
		return false, fmt.Errorf("peer %s is blacklisted", peerAddr)
	}

	logger.DebugLogger.Printf("Block received: %s from %s", block.Hash, peerAddr)

	if s.mode == 3 && (block.Header.Height >= 1 && block.Header.Height < 5) {
		return true, nil
	}

	// 2) If block is invalid, increment invalid count
	if !block.Verify() {
		s.PeerManager.IncrementInvalidCount(peerAddr)
		logger.InfoLogger.Printf("Invalid block: %s from %s", block.Hash, peerAddr)
		return false, fmt.Errorf("invalid block by peer %s", peerAddr)
	}

	lastBlock := s.Blockchain.GetLastBlock()
	if block.Header.PreviousHash == lastBlock.Hash {
		if err := s.Blockchain.AddBlock(block); err != nil {
			logger.ErrorLogger.Printf("[SubmitBlock] Failed to add block to main chain, hash: %s, Error: %v", block.Hash, err)
			return false, fmt.Errorf("failed to add block to the main chain: %v", err)
		}

		// Log time difference between blocks
		prevBlock := s.Blockchain.GetBlockByHash(block.Header.PreviousHash)
		if prevBlock != nil {
			timeDiff := float64(block.Header.Timestamp-prevBlock.Header.Timestamp) / 1000.0
			logger.InfoLogger.Printf("Block time difference: %.1fs (Height=%d)", timeDiff, block.Header.Height)
		}

		s.Comms.BroadcastBlock(block, *hashes)
		logger.InfoLogger.Printf("Block received and added: %s", block.Hash)
		return true, nil
	}

	if s.Blockchain.GetBlockByHash(block.Hash) != nil {
		logger.DebugLogger.Printf("Duplicate block: %s", block.Hash)
		return false, fmt.Errorf("block already in the blockchain")
	}

	err := s.Blockchain.HandleFork(*hashes, s.Comms.RequestBlockByHash)
	if err != nil {
		logger.ErrorLogger.Printf("[SubmitBlock] Fork handling error for block hash: %s, Error: %v", block.Hash, err)
		return false, fmt.Errorf("fork handling error: %v", err)
	}

	s.Comms.BroadcastBlock(block, *hashes)
	logger.DebugLogger.Printf("Fork resolved: %s", block.Hash)
	return true, nil
}

func (s *BlockchainServer) createBlockWithTransactions() (*blockchain.Block, error) {
	for {
		transactions := s.TxPool.GetUpToNTransactions(1, s.Blockchain.UTXOSet)
		if len(transactions) > 0 {
			block, err := s.Blockchain.CreateBlock(transactions)
			return block, err
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *BlockchainServer) MineBlocks() error {
	logger.InfoLogger.Println("Mining started")

	if s.mining {
		logger.DebugLogger.Println("Mining already in progress")
		return fmt.Errorf("mining is already in progress")
	}

	mineCtx, cancelFunc := context.WithCancel(context.Background())
	s.cancelFunc = cancelFunc
	s.mining = true

	go func() {
		defer func() {
			s.mining = false
			s.cancelFunc = nil
		}()

		for {
			select {
			case <-mineCtx.Done():
				logger.InfoLogger.Println("Mining stopped")
				return
			default:
				time.Sleep(1 * time.Second)
				block, err := s.createBlockWithTransactions()
				if err != nil {
					logger.ErrorLogger.Printf("[MineBlocks] Error creating block: %v", err)
					continue
				}

				logger.DebugLogger.Printf("Created new block: Height=%d, Transactions=%d", block.Header.Height, len(block.Content.Transactions))

				hash, _ := block.CalculateHash()

				// Initialize random number generator with current time
				rand.Seed(time.Now().UnixNano())

				for !block.VerifyHash(hash) {
					// Generate random nonce in a large range
					block.Header.Nonce = rand.Int63n(100000000000)

					if block.Header.Nonce%10000 == 0 {
						logger.DebugLogger.Printf("Mining progress: Height=%d, Nonce=%d", block.Header.Height, block.Header.Nonce)
						if block.Header.Height <= s.Blockchain.GetLastBlock().Header.Height {
							logger.DebugLogger.Printf("Chain advanced while mining: Height=%d", block.Header.Height)
							block, err = s.createBlockWithTransactions()
							if err != nil {
								logger.ErrorLogger.Printf("[MineBlocks] Error creating block: %v", err)
								continue
							}
						}
					}

					hash, _ = block.CalculateHash()
				}

				logger.DebugLogger.Printf("Valid hash found: Height=%d, Hash=%s", block.Header.Height, hash)

				block.Hash = hash

				if block.Header.Height <= s.Blockchain.GetLastBlock().Header.Height {
					logger.DebugLogger.Println("Chain advanced, restarting mining")
					continue
				}

				// Handle different modes
				switch s.mode {
				case 1:
					// Corrupt first transaction hash
					if len(block.Content.Transactions) > 0 {
						block.Content.Transactions[0].Hash = "0"
						logger.InfoLogger.Printf("Corrupted block: Height=%d, Hash=%s", block.Header.Height, block.Hash)
					}
				case 4:
					// Fall through to case 2 (without stopping on first iteration)
					fallthrough
				case 2:
					// Lie about hash and nonce
					block.Hash = "1"
					block.Header.Nonce = 1
					logger.InfoLogger.Printf("Lied about block: Height=%d, Hash=%s", block.Header.Height, block.Hash)
				case 3:
					// Fork mode - don't broadcast until 5th block
					if block.Header.Height > 1 && block.Header.Height < 5 {
						if err := s.Blockchain.AddBlock(block); err != nil {
							logger.ErrorLogger.Printf("[MineBlocks] Error adding block: %v", err)
							continue
						}
						logger.InfoLogger.Printf("Fork block mined (not broadcast): Height=%d, Hash=%s", block.Header.Height, block.Hash)
						continue
					}
				}

				if err := s.Blockchain.AddBlock(block); err != nil && s.mode == 0 {
					logger.ErrorLogger.Printf("[MineBlocks] Error adding block: %v", err)
					continue
				}

				// Log time difference between blocks
				prevBlock := s.Blockchain.GetBlockByHash(block.Header.PreviousHash)
				if prevBlock != nil {
					timeDiff := float64(block.Header.Timestamp-prevBlock.Header.Timestamp) / 1000.0
					logger.InfoLogger.Printf("Block time difference: %.1fs (Height=%d)", timeDiff, block.Header.Height)
				}

				logger.InfoLogger.Printf("Block mined: Height=%d, Hash=%s", block.Header.Height, block.Hash)

				// Don't broadcast in fork mode until 5th block
				if s.mode != 3 || block.Header.Height == 1 || block.Header.Height >= 5 {
					s.Comms.BroadcastBlock(block, s.Blockchain.GetLast100Hashes())
					logger.DebugLogger.Printf("Block broadcasted: Height=%d, Hash=%s", block.Header.Height, block.Hash)
				}

				// Stop after first block in modes 1 and 2
				if s.mode == 1 || s.mode == 2 || (s.mode == 3 && block.Header.Height == 10) {
					s.StopMining()
					return
				}
			}
		}
	}()

	return nil
}

func (s *BlockchainServer) StopMining() error {
	logger.InfoLogger.Println("Mining stop requested")

	if !s.mining {
		logger.DebugLogger.Println("Mining not active")
		return fmt.Errorf("mining is not active")
	}

	s.cancelFunc()
	s.mining = false
	logger.InfoLogger.Println("Mining stopped successfully")

	return nil
}
