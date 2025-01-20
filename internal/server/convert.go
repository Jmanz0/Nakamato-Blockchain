package server

import (
	"nakamoto-blockchain/internal/blockchain"
	"nakamoto-blockchain/proto/gen"
)

// ConvertUTXOToGrpc Converts a UTXO to a grpc UTXO.
func ConvertUTXOToGrpc(utxo *blockchain.UTXO) *gen.UTXO {
	return &gen.UTXO{
		TxHash:  utxo.TxID,
		Index:   int32(utxo.Index),
		Amount:  utxo.Amount,
		Address: utxo.Address,
	}
}

// ConvertTransactionToGrpc Converts a transaction to a grpc transaction.
func ConvertTransactionToGrpc(tx *blockchain.Transaction) *gen.Transaction {
	grpcInputs := make([]*gen.UTXO, len(tx.Content.InputUTXOs))
	for i, input := range tx.Content.InputUTXOs {
		grpcInputs[i] = ConvertUTXOToGrpc(&input)
	}

	grpcOutputs := make([]*gen.UTXO, len(tx.Content.OutputUTXOs))
	for i, output := range tx.Content.OutputUTXOs {
		grpcOutputs[i] = ConvertUTXOToGrpc(&output)
	}

	return &gen.Transaction{
		Inputs:     grpcInputs,
		Outputs:    grpcOutputs,
		Timestamp:  tx.Content.Timestamp,
		Signature:  tx.Signature,
		Hash:       tx.Hash,
		Senderpubkey: tx.Content.SenderPubKey,
	}
}

// ConvertBlockToGrpc Converts a block to a grpc block.
func ConvertBlockToGrpc(block *blockchain.Block) *gen.Block {
	grpcBlockHeader := &gen.BlockHeader{
		Timestamp:    block.Header.Timestamp,
		PreviousHash: block.Header.PreviousHash,
		ContentHash:  block.Header.ContentHash,
		Height:       int32(block.Header.Height),
		Difficulty:   block.Header.Difficulty,
		Nonce:        block.Header.Nonce,
	}

	grpcTransactions := make([]*gen.Transaction, len(block.Content.Transactions))
	for i, tx := range block.Content.Transactions {
		grpcTransactions[i] = ConvertTransactionToGrpc(&tx)
	}

	grpcBlockContent := &gen.BlockContent{
		Transactions: grpcTransactions,
	}

	grpcBlock := &gen.Block{
		Header:  grpcBlockHeader,
		Content: grpcBlockContent,
		Hash:    block.Hash,
	}

	return grpcBlock
}

// ConvertBlockHeadersToGrpc Converts a slice of block headers to a slice of grpc block headers.
func ConvertBlockHeadersToGrpc(blockHeaders []blockchain.BlockHeader) []*gen.BlockHeader {
	grpcBlockHeaders := make([]*gen.BlockHeader, len(blockHeaders))
	for i, header := range blockHeaders {
		grpcBlockHeaders[i] = &gen.BlockHeader{
			Timestamp:    header.Timestamp,
			PreviousHash: header.PreviousHash,
			ContentHash:  header.ContentHash,
			Height:       int32(header.Height),
			Difficulty:   header.Difficulty,
			Nonce:        header.Nonce,
		}
	}
	return grpcBlockHeaders
}

// ConvertGrpcToUTXO Converts a grpc UTXO to a UTXO.
func ConvertGrpcToUTXO(grpcUTXO *gen.UTXO) *blockchain.UTXO {
	return &blockchain.UTXO{
		TxID:    grpcUTXO.TxHash,
		Index:   int(grpcUTXO.Index),
		Amount:  grpcUTXO.Amount,
		Address: grpcUTXO.Address,
	}
}

func ConvertGrpcToUTXOs(grpcUTXOs []*gen.UTXO) []blockchain.UTXO {
	utxos := make([]blockchain.UTXO, len(grpcUTXOs))
	for i, utxo := range grpcUTXOs {
		utxos[i] = *ConvertGrpcToUTXO(utxo)
	}
	return utxos
}

// ConvertGrpcToTransaction Converts a grpc transaction to a transaction.
func ConvertGrpcToTransaction(grpcTx *gen.Transaction) *blockchain.Transaction {
	inputUTXOs := make([]blockchain.UTXO, len(grpcTx.Inputs))
	for i, input := range grpcTx.Inputs {
		inputUTXOs[i] = *ConvertGrpcToUTXO(input)
	}

	outputUTXOs := make([]blockchain.UTXO, len(grpcTx.Outputs))
	for i, output := range grpcTx.Outputs {
		outputUTXOs[i] = *ConvertGrpcToUTXO(output)
	}

	return &blockchain.Transaction{
		Content: blockchain.TransactionContent{
			InputUTXOs:  inputUTXOs,
			OutputUTXOs: outputUTXOs,
			Timestamp:   grpcTx.Timestamp,
			SenderPubKey: grpcTx.Senderpubkey,
		},
		Signature: grpcTx.Signature,
		Hash:      grpcTx.Hash,
	}
}

// ConvertGrpcToBlock Converts a grpc block to a block.
func ConvertGrpcToBlock(grpcBlock *gen.Block) *blockchain.Block {
	blockHeader := blockchain.BlockHeader{
		Timestamp:    grpcBlock.Header.Timestamp,
		PreviousHash: grpcBlock.Header.PreviousHash,
		ContentHash:  grpcBlock.Header.ContentHash,
		Height:       int(grpcBlock.Header.Height),
		Difficulty:   grpcBlock.Header.Difficulty,
		Nonce:        grpcBlock.Header.Nonce,
	}

	transactions := make([]blockchain.Transaction, len(grpcBlock.Content.Transactions))
	for i, tx := range grpcBlock.Content.Transactions {
		transactions[i] = *ConvertGrpcToTransaction(tx)
	}

	blockContent := blockchain.BlockContent{
		Transactions: transactions,
	}

	return &blockchain.Block{
		Header:  blockHeader,
		Content: blockContent,
		Hash:    grpcBlock.Hash,
	}
}

// ConvertGrpcHeadersToBlockHeaders Converts a slice of grpc block headers to a slice of block headers.
func ConvertGrpcHeadersToBlockHeaders(grpcHeaders []*gen.BlockHeader) []blockchain.BlockHeader {
	headers := make([]blockchain.BlockHeader, len(grpcHeaders))
	for i, header := range grpcHeaders {
		headers[i] = blockchain.BlockHeader{
			Timestamp:    header.Timestamp,
			PreviousHash: header.PreviousHash,
			ContentHash:  header.ContentHash,
			Height:       int(header.Height),
			Difficulty:   header.Difficulty,
			Nonce:        header.Nonce,
		}
	}
	return headers
}
