package blockchain

import (
	"errors"
	"math/big"
	"nakamoto-blockchain/internal/crypto"
	"strings"
	"time"
)

type BlockHeader struct {
	Timestamp    int64
	PreviousHash string
	ContentHash  string
	Height       int
	Difficulty   string
	Nonce        int64
}

type BlockContent struct {
	Transactions []Transaction
}

type Block struct {
	Header  BlockHeader
	Content BlockContent
	Hash    string
}

func (b *Block) CalculateContentHash() (string, error) {
	// ? I think mining empty blocks is fine, technically increases security and also makes demo easier
	// if len(b.Content.Transactions) == 0 {
	// 	return "", errors.New("no transactions in the block")
	// }

	var transactionHashes []string
	for _, tx := range b.Content.Transactions {
		if tx.Hash == "" {
			return "", errors.New("transaction hash is empty")
		}
		if !tx.Verify() {
			return "", errors.New("transaction verification failed")
		}
		transactionHashes = append(transactionHashes, tx.Hash)
	}
	contentHashString := strings.Join(transactionHashes, "")
	contentHash, err := crypto.Hash(contentHashString)
	if err != nil {
		return "", err
	}
	return contentHash, nil
}

func (b *Block) CalculateHash() (string, error) {
	blockHash, err := crypto.Hash(b.Header)
	if err != nil {
		return "", err
	}
	return blockHash, nil
}

func (b *Block) VerifyHash(hash string) bool {
	hashBigInt := new(big.Int)
	hashBigInt.SetString(hash, 16)

	difficultyBigInt := new(big.Int)
	difficultyBigInt.SetString(b.Header.Difficulty, 16)

	return hashBigInt.Cmp(difficultyBigInt) < 0
}

func (b *Block) Verify() bool {
	contentHash, err := b.CalculateContentHash()

	if err != nil {
		return false
	}

	if contentHash != b.Header.ContentHash {
		return false
	}

	hash, err := b.CalculateHash()
	if err != nil {
		return false
	}

	if hash != b.Hash {
		return false
	}

	return b.VerifyHash(hash)
}

func NewBlock(previousHash string, height int, difficulty string, transactions []Transaction) (*Block, error) {
	block := &Block{
		Header: BlockHeader{
			Timestamp:    time.Now().UnixMilli(),
			PreviousHash: previousHash,
			Height:       height,
			Difficulty:   difficulty,
			Nonce:        0,
		},
		Content: BlockContent{
			Transactions: transactions,
		},
	}
	contentHash, err := block.CalculateContentHash()
	if err != nil {
		return nil, err
	}
	block.Header.ContentHash = contentHash

	return block, nil
}
