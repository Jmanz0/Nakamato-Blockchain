package blockchain

import "fmt"

type TransactionPool struct {
	Transactions map[string]Transaction
}

func NewTransactionPool() *TransactionPool {
	return &TransactionPool{Transactions: make(map[string]Transaction)}
}

func (tp *TransactionPool) GetAllTransactions() []Transaction {
	var transactions []Transaction
	for _, tx := range tp.Transactions {
		transactions = append(transactions, tx)
	}
	return transactions
}

func (tp *TransactionPool) GetUpToNTransactions(n int, utxoSet *UTXOSet) []Transaction {
	var transactions []Transaction
	count := 0
	for _, tx := range tp.Transactions {
		if count >= n {
			break
		}
		if utxoSet.CheckTransaction(tx) {
			transactions = append(transactions, tx)
			count++
		}
	}
	return transactions
}

func (tp *TransactionPool) AddTransaction(tx Transaction) error {
	if _, exists := tp.Transactions[tx.Hash]; exists {
		return fmt.Errorf("transaction with hash %s already exists in the pool", tx.Hash)
	}
	tp.Transactions[tx.Hash] = tx
	return nil
}

func (tp *TransactionPool) RemoveTransaction(tx Transaction) error {
	if _, exists := tp.Transactions[tx.Hash]; !exists {
		return fmt.Errorf("transaction with hash %s not found in the pool", tx.Hash)
	}
	delete(tp.Transactions, tx.Hash)
	return nil
}

func (tp *TransactionPool) HasTransaction(hash string) bool {
	_, exists := tp.Transactions[hash]
	return exists
}

func (tp *TransactionPool) Get(hash string) (*Transaction, error) {
	tx, exists := tp.Transactions[hash]
	if !exists {
		return nil, fmt.Errorf("transaction with hash %s not found in the pool", hash)
	}
	return &tx, nil
}

func (tp *TransactionPool) List() ([]Transaction, error) {
	if len(tp.Transactions) == 0 {
		return nil, fmt.Errorf("transaction pool is empty")
	}
	transactions := []Transaction{}
	for _, tx := range tp.Transactions {
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func (tp *TransactionPool) HandleStaleBlocks(staleBlocks []*Block) {
	for _, block := range staleBlocks {
		for _, tx := range block.Content.Transactions {
			if !tp.HasTransaction(tx.Hash) {
				tp.AddTransaction(tx)
			}
		}
	}
}
