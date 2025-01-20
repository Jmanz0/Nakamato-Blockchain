package blockchain

import (
	"errors"
	"nakamoto-blockchain/internal/crypto"
	"nakamoto-blockchain/logger"
	"time"
)

type TransactionContent struct {
	InputUTXOs   []UTXO
	OutputUTXOs  []UTXO
	SenderPubKey string
	Timestamp    int64
}

type Transaction struct {
	Content   TransactionContent
	Signature string
	Hash      string
}

func NewTransaction(InputUTXOs []UTXO, SenderPubKey, ReceiverPublicKey string, Amount int64) (*Transaction, error) {
	var totalInput int64
	for _, utxo := range InputUTXOs {
		totalInput += utxo.Amount
	}

	var outputUTXOs []UTXO
	outputUTXOs = append(outputUTXOs, UTXO{
		TxID:    "",
		Index:   0,
		Amount:  Amount,
		Address: crypto.Key2Addr(ReceiverPublicKey),
	})

	leftoverAmount := totalInput - Amount
	if leftoverAmount > 0 {
		outputUTXOs = append(outputUTXOs, UTXO{
			TxID:    "",
			Index:   1,
			Amount:  leftoverAmount,
			Address: crypto.Key2Addr(SenderPubKey),
		})
	}

	txContent := TransactionContent{
		InputUTXOs:   InputUTXOs,
		OutputUTXOs:  outputUTXOs,
		SenderPubKey: SenderPubKey,
		Timestamp:    time.Now().UnixMilli(),
	}

	tx := &Transaction{
		Content: txContent,
	}

	if !tx.VerifyContent() {
		return nil, errors.New("transaction content verification failed")
	}

	return tx, nil
}

func (tx *Transaction) GetUTXO(Index int) (UTXO, error) {
	if Index < 0 || Index >= len(tx.Content.OutputUTXOs) {
		return UTXO{}, errors.New("invalid UTXO index")
	}

	OutUTXO := tx.Content.OutputUTXOs[Index]
	OutUTXO.TxID = tx.Hash
	return OutUTXO, nil
}

func (tx *Transaction) Sign(privateKey string) error {
	txHash, err := crypto.Hash(tx)
	if err != nil {
		return err
	}

	tx.Signature, err = crypto.Sign(txHash, privateKey)
	if err != nil {
		return err
	}

	tx.Hash = txHash
	return nil
}

func (tx *Transaction) VerifySignature() bool {
	valid, err := crypto.VerifySignature(tx.Hash, tx.Signature, tx.Content.SenderPubKey)
	return err == nil && valid
}

func (tx *Transaction) VerifyContent() bool {
	var totalInput, totalOutput int64

	if len(tx.Content.OutputUTXOs) > 2 {
		logger.WarnLogger.Println("Transaction has more than 2 output UTXOs")
		return false
	}

	for _, utxo := range tx.Content.InputUTXOs {
		if utxo.Amount <= 0 || utxo.TxID == "" {
			logger.WarnLogger.Println("Invalid input UTXO")
			return false
		}
		totalInput += utxo.Amount
	}

	for i, utxo := range tx.Content.OutputUTXOs {
		if utxo.Amount <= 0 || utxo.TxID != "" || utxo.Index != i {
			logger.WarnLogger.Println("Invalid output UTXO")
			return false
		}
		totalOutput += utxo.Amount
	}

	if totalInput != totalOutput {
		logger.WarnLogger.Println("Total input amount does not match total output amount")
		return false
	}

	senderAddress := crypto.Key2Addr(tx.Content.SenderPubKey)
	for i, utxo := range tx.Content.InputUTXOs {
		if utxo.Address != senderAddress {
			logger.WarnLogger.Printf("Input UTXO address mismatch in transaction %s (UTXO index %d): received %s, expected %s", 
				tx.Hash, i, utxo.Address, senderAddress)
			return false
		}
	}

	if len(tx.Content.OutputUTXOs) > 1 && tx.Content.OutputUTXOs[1].Address != senderAddress {
		logger.WarnLogger.Println("Leftover UTXO address does not match sender address")
		return false
	}

	return true
}

func (tx *Transaction) Verify() bool {
	return tx.VerifyContent() && tx.VerifySignature()
}
