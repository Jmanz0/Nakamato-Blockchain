package blockchain

import (
	"fmt"
	"math/big"
	"nakamoto-blockchain/logger"
	"strings"
)

type Blockchain struct {
	Blocks  []*Block
	UTXOSet *UTXOSet
}

func NewBlockchain(initialUTXOs []UTXO) *Blockchain {
	bc := &Blockchain{
		Blocks:  []*Block{},
		UTXOSet: NewUTXOSet(),
	}

	for _, utxo := range initialUTXOs {
		bc.UTXOSet.AddUTXO(utxo)
	}

	genesisBlock, err := NewBlock("", 0, bc.GetDifficulty(0), []Transaction{})
	if err != nil {
		panic(fmt.Sprintf("failed to create genesis block: %v", err))
	}

	bc.Blocks = append(bc.Blocks, genesisBlock)
	return bc
}

const (
	targetBlockTime  = 20 // seconds
	difficultyWindow = 10 // blocks
	dynamicStart     = 1000
)

func (bc *Blockchain) GetDifficulty(cur int) string {
	// Return initial difficulty for first difficultyWindow blocks
	if cur <= dynamicStart {
		return "fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	}

	// Get last difficultyWindow blocks
	start := cur - difficultyWindow
	if start < 2 {
		start = 2
	}

	// Calculate total time and sum of inverse difficulties
	totalTime := bc.Blocks[cur-1].Header.Timestamp - bc.Blocks[start-1].Header.Timestamp
	totalDifficulty := new(big.Rat)

	for i := start; i < cur; i++ {
		if i == 0 {
			continue // skip genesis block
		}

		// Sum inverse difficulties
		blockDiff := new(big.Int)
		blockDiff.SetString(bc.Blocks[i].Header.Difficulty, 16)
		if blockDiff.Sign() == 0 {
			continue // skip zero difficulty blocks
		}
		invDiff := new(big.Rat).SetFrac(big.NewInt(1), blockDiff)
		totalDifficulty.Add(totalDifficulty, invDiff)
	}

	// Convert totalTime to seconds
	totalTimeSec := new(big.Rat).SetInt64(totalTime / 1000)

	// Calculate difficulty ratio: totalDifficulty / totalTimeSec
	ratio := new(big.Rat).Quo(totalDifficulty, totalTimeSec)

	// Multiply by target block time
	adjustedRatio := new(big.Rat).Mul(ratio, new(big.Rat).SetInt64(int64(targetBlockTime)))

	// Take reciprocal for new difficulty
	newDiff := new(big.Rat).Inv(adjustedRatio)

	// Convert to big.Int and then to hex string
	newDiffInt := new(big.Int)
	newDiffInt.Div(newDiff.Num(), newDiff.Denom())

	newDiffStr := fmt.Sprintf("%x", newDiffInt)
	logger.InfoLogger.Printf("New difficulty for block %d: %s\n", cur, newDiffStr)

	return newDiffStr
}

func (bc *Blockchain) GetLastBlock() *Block {
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) Verify() bool {
	previousHash := ""
	for i, block := range bc.Blocks {
		if block.Header.Height != i {
			return false
		}

		if block.Header.PreviousHash != previousHash {
			return false
		}
		previousHash = block.Hash

		if block.Header.Difficulty != bc.GetDifficulty(i) {
			return false
		}

		if !block.Verify() {
			return false
		}

	}
	return true
}

func (bc *Blockchain) ValidateBlocks(blocks []*Block) error {
	previousHash := ""
	if len(bc.Blocks) > 0 {
		previousHash = bc.Blocks[len(bc.Blocks)-1].Hash
	}

	for i, block := range blocks {
		if !block.Verify() {
			return fmt.Errorf("block verification failed for block at index %d", i)
		}

		if i > 0 && block.Header.PreviousHash != previousHash {
			return fmt.Errorf("previous hash mismatch at block index %d", i)
		}

		previousHash = block.Hash
	}

	return nil
}

func (bc *Blockchain) HasTransaction(txHash string) bool {
	for _, block := range bc.Blocks {
		for _, tx := range block.Content.Transactions {
			if tx.Hash == txHash {
				return true
			}
		}
	}
	return false
}

func (bc *Blockchain) GetTransactionDepth(txHash string) int {
	for i := len(bc.Blocks) - 1; i >= 0; i-- {
		block := bc.Blocks[i]
		for _, tx := range block.Content.Transactions {
			if tx.Hash == txHash {
				return len(bc.Blocks) - i
			}
		}
	}
	return -1
}

func (bc *Blockchain) CreateBlock(transactions []Transaction) (*Block, error) {
	previousHash := ""
	if len(bc.Blocks) != 0 {
		previousHash = bc.Blocks[len(bc.Blocks)-1].Hash
	}
	height := len(bc.Blocks)
	return NewBlock(previousHash, height, bc.GetDifficulty(height), transactions)
}

func (bc *Blockchain) GetBlockByHeight(height int) *Block {
	if height < 0 || height >= len(bc.Blocks) {
		return nil
	}
	return bc.Blocks[height]
}

func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	for _, block := range bc.Blocks {
		if block.Hash == hash {
			return block
		}
	}
	return nil
}

func (bc *Blockchain) AddBlock(block *Block) error {
	if !block.Verify() {
		return fmt.Errorf("block verification failed")
	}

	if len(bc.Blocks) != 0 {
		previousBlock := bc.Blocks[len(bc.Blocks)-1]
		if block.Header.PreviousHash != previousBlock.Hash {
			return fmt.Errorf("previous hash mismatch")
		}
		if block.Header.Height != previousBlock.Header.Height+1 {
			return fmt.Errorf("height mismatch")
		}
	}

	err := bc.UTXOSet.AddBlock(block)
	if err != nil {
		return err
	}

	bc.Blocks = append(bc.Blocks, block)

	if block.Header.Height%10 == 0 {
		var hashes []string
		for _, blk := range bc.Blocks {
			hashes = append(hashes, blk.Hash)
		}
		logger.InfoLogger.Printf("Blockchain hashes at height %d: %s", block.Header.Height, strings.Join(hashes, ", "))
	}

	return nil
}

// func (bc *Blockchain) Fork(hash string) *Blockchain {
// 	index := -1
// 	for i, block := range bc.Blocks {
// 		if block.Hash == hash {
// 			index = i
// 		}
// 	}
// 	if index == -1 {
// 		return nil
// 	}

// 	newChain := NewBlockchain([]UTXO{})
// 	for i := 0; i <= index; i++ {
// 		newChain.AddBlock(bc.Blocks[i])
// 	}
// 	return newChain
// }

func (bc *Blockchain) RollbackToHash(toHash string) ([]*Block, error) {
	removedBlocks := []*Block{}

	for len(bc.Blocks) > 0 {
		lastBlock := bc.Blocks[len(bc.Blocks)-1]
		if lastBlock.Hash == toHash {
			break
		}

		removedBlocks = append(removedBlocks, lastBlock)
		bc.Blocks = bc.Blocks[:len(bc.Blocks)-1]
	}

	if len(bc.Blocks) == 0 || bc.Blocks[len(bc.Blocks)-1].Hash != toHash {
		return nil, fmt.Errorf("ancestor block with hash %s not found", toHash)
	}

	// return removed transactions back to mempool

	return removedBlocks, nil
}

// Dealing With Forks
/*
When we receive a new block from a peer, we need to check if it is part of the main chain or a fork.
There are a couple of scenarios to consider:
	Block Extends Main Chain				Add to main chain, process orphans.
	Block Extends Known Fork				Add to fork, validate fork, and switch if longer.
	Block References Unknown Parent			Add to orphan pool, request missing blocks.
	Invalid Block							Reject and log.
	Duplicate Block							Discard and log.
	Stale Block								Move transactions to mempool, discard block.
	Resolved Orphans						Add to chain, process dependent orphans recursively.
	Fork with Higher PoW					Rollback main chain, apply fork blocks.

Potential Improvement:
	1) Validate the block
	2) Check if the block extends the main chain; add to chain. Break.
	3) Create a new fork if block extends main chain.
	4) Check if the block extends a known fork; add to fork. Validate fork.
		a) If the fork is longer, switch to fork.
		b) If the block is a fork of a fork, add as a new fork with reference to parent.
	5) Check if the block references an unknown parent; add to orphan pool.

	6) If added to chain, process orphans.

Current Approach:
	1) Receive a new block (which contains the last 100 block head)
	2) Check if the block extends the main chain
	3) If block is longer, and dynamically ask for the missing blocks and update main if longer
*/

/*
	I still to do the following:
		1) Validate the incoming chain
*/

func (bc *Blockchain) GetLast100Hashes() []string {
	hashes := []string{}
	start := 0

	if len(bc.Blocks) > 100 {
		start = len(bc.Blocks) - 100
	}

	for i := start; i < len(bc.Blocks); i++ {
		hashes = append(hashes, bc.Blocks[i].Hash)
	}

	return hashes
}

// ? Here we are passing requestBlock down, alternatively we can lift the handle fork function to the blockchain node level
func (bc *Blockchain) HandleFork(incomingHashes []string, requestBlock func(hash string) *Block) error {
	logger.DebugLogger.Println("[HandleFork] Handling fork...")

	ancestor := bc.FindCommonAncestor(incomingHashes)
	if ancestor == nil {
		logger.ErrorLogger.Println("[HandleFork] Common ancestor not found for incoming hashes.")
		return fmt.Errorf("common ancestor not found for block")
	}
	logger.DebugLogger.Printf("[HandleFork] Found common ancestor: %s", ancestor.Hash)

	mainChainWork := bc.ComputeWork(bc.GetLast100Hashes(), ancestor.Hash)
	forkChainWork := bc.ComputeWork(incomingHashes, ancestor.Hash)

	logger.DebugLogger.Printf("[HandleFork] Main chain work: %d, Fork chain work: %d", mainChainWork, forkChainWork)

	if forkChainWork > mainChainWork {
		logger.InfoLogger.Println("[HandleFork] Fork chain has more work. Handling fork replacement...")

		// ? Here we are passing requestBlock down, alternatively we can lift the handle fork function to the blockchain node level
		missingBlocks := bc.RequestMissingBlocks(incomingHashes, ancestor.Hash, requestBlock)
		logger.InfoLogger.Printf("[HandleFork] Retrieved %d missing blocks from fork.", len(missingBlocks))

		if err := bc.ValidateBlocks(missingBlocks); err != nil {
			logger.InfoLogger.Printf("[HandleFork] Validation failed for blocks in fork: %v", err)
			return fmt.Errorf("invalid blocks in fork: %v", err)
		}
		logger.InfoLogger.Println("[HandleFork] Fork blocks validated successfully.")

		if err := bc.ReplaceWithFork(missingBlocks); err != nil {
			logger.ErrorLogger.Printf("[HandleFork] Failed to replace main chain with fork: %v", err)
			return fmt.Errorf("failed to replace chain with fork: %v", err)
		}
		logger.InfoLogger.Println("[HandleFork] Successfully replaced main chain with fork.")
		// Print last 100 hashes
		logger.InfoLogger.Printf("[HandleFork] Last 100 hashes: %v", bc.GetLast100Hashes())
	} else {
		logger.DebugLogger.Println("[HandleFork] Main chain has more work. No changes applied.")
	}

	return nil
}

func (bc *Blockchain) FindCommonAncestor(incomingHashes []string) *Block {
	bcHashToHeight := make(map[string]int)
	for i, block := range bc.Blocks {
		bcHashToHeight[block.Hash] = i
	}

	var best *Block
	bestHeight := -1

	for _, hash := range incomingHashes {
		if h, exists := bcHashToHeight[hash]; exists && h > bestHeight {
			bestHeight = h
			best = bc.Blocks[h]
		}
	}
	return best
}

// Does not work in dynamic difficulty
func (bc *Blockchain) ComputeWork(hashes []string, endHash string) int {
	return len(hashes)
}

// ? Here we are passing requestBlock down, alternatively we can lift the handle fork function to the blockchain node level
func (bc *Blockchain) RequestMissingBlocks(incomingHashes []string, startHash string, requestBlock func(hash string) *Block) []*Block {
	missingBlocksReversed := []*Block{}
	for i := len(incomingHashes) - 1; i >= 0; i-- {
		if incomingHashes[i] == startHash {
			break
		}
		missingBlocksReversed = append(missingBlocksReversed, requestBlock(incomingHashes[i]))
	}
	missingBlocks := []*Block{}
	for i := 0; i < len(missingBlocksReversed); i++ {
		missingBlocks = append(missingBlocks, missingBlocksReversed[len(missingBlocksReversed)-i-1])
	}
	return missingBlocks
}

func (bc *Blockchain) ReplaceWithFork(missingBlocks []*Block) error {
	if len(missingBlocks) == 0 {
		return fmt.Errorf("no blocks to replace with")
	}

	ancestorHash := missingBlocks[0].Header.PreviousHash

	_, err := bc.RollbackToHash(ancestorHash)
	if err != nil {
		return fmt.Errorf("failed to rollback to ancestor: %v", err)
	}

	for _, block := range missingBlocks {
		bc.Blocks = append(bc.Blocks, block)
	}

	return nil
}
