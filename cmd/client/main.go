/*
Initialize a single client that will generate transactions and send them to a miner forever.
Takes in three arguments:
1. A file containing the initial UTXOs for each public key (what blockchain is initialized with).
2. A file containing the public and private keys of the clients.
3. A file containing the IP addresses of the miners.

The client will generate transactions between the public keys in the keys file and send them to the miners in a round-robin fashion.
The client will also update its UTXO set with the new transactions, able to continously make and send transactions forever.
*/

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"nakamoto-blockchain/internal/blockchain"
	"nakamoto-blockchain/internal/crypto"
	"nakamoto-blockchain/internal/server"
	"nakamoto-blockchain/logger"
	"nakamoto-blockchain/proto/gen"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	mu      sync.Mutex
	utxoSet = blockchain.NewUTXOSet()
)

func main() {
	logger.Init()

	if len(os.Args) < 4 {
		logger.ErrorLogger.Fatal("[Client] Usage: go run main.go <initialUXTOS_file> <keys_file> <miner_ip_file>")
	}

	initialUXTOSFile := os.Args[1]
	keysFile := os.Args[2]
	minerIPFile := os.Args[3]

	logger.InfoLogger.Println("[Client] Starting transaction generator...")

	keyMap, err := readKeyMap(keysFile)
	if err != nil {
		logger.ErrorLogger.Fatal("[Client] Failed to read keys:", err)
	}

	err = populateInitialUTXOSet(initialUXTOSFile)
	if err != nil {
		logger.ErrorLogger.Fatal("[Client] Failed to populate UTXOs:", err)
	}

	minerIPs, err := readMinerIPs(minerIPFile)
	if err != nil {
		logger.ErrorLogger.Fatal("[Client] Failed to read miner IPs:", err)
	}

	if len(minerIPs) == 0 {
		logger.ErrorLogger.Fatal("[Client] No miners available")
	}

	logger.InfoLogger.Println("[Client] Successfully initialized. Starting transaction loop...")

	for {
		createAndSendTransactions(keyMap, minerIPs)
		time.Sleep(2 * time.Second)
	}
}

func readMinerIPs(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var minerIPs []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		minerIPs = append(minerIPs, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return minerIPs, nil
}

func readKeyMap(filename string) (map[string]string, error) {
	logger.InfoLogger.Println("[Client] Reading keys from file:", filename)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keyPairs []struct {
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&keyPairs); err != nil {
		return nil, fmt.Errorf("invalid key file format: %w", err)
	}

	keyMap := make(map[string]string)
	for _, keyPair := range keyPairs {
		keyMap[keyPair.PublicKey] = keyPair.PrivateKey
	}

	logger.InfoLogger.Println("[Client] Successfully read keys.")
	return keyMap, nil
}

func populateInitialUTXOSet(filename string) error {
	logger.InfoLogger.Println("[Client] Populating UTXOs from file:", filename)

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open UTXO file: %w", err)
	}
	defer file.Close()

	var utxoEntries []struct {
		PublicKey string            `json:"public_key"`
		UTXOs     []blockchain.UTXO `json:"utxos"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&utxoEntries); err != nil {
		return fmt.Errorf("invalid UTXO format: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, entry := range utxoEntries {
		address := crypto.Key2Addr(entry.PublicKey)
		for _, utxo := range entry.UTXOs {
			utxo.Address = address
			utxoSet.AddUTXO(utxo)
		}
	}

	logger.InfoLogger.Printf("[Client] Successfully populated %d UTXO entries.", len(utxoEntries))
	utxoSet.PrintAllUTXOs()
	return nil
}

func createAndSendTransactions(keyMap map[string]string, minerIPs []string) {
	mu.Lock()
	defer mu.Unlock()

	logger.InfoLogger.Println("[Client] Starting transaction creation and sending process...")
	for senderPubKey := range keyMap {
		logger.InfoLogger.Printf("[Client] Processing sender: %s", senderPubKey)

		address := crypto.Key2Addr(senderPubKey)
		senderUTXOs := utxoSet.Get(address)
		if len(senderUTXOs) == 0 {
			logger.InfoLogger.Printf("[Client] No UTXOs available for sender: %s. Skipping.", senderPubKey)
			continue
		}
		logger.InfoLogger.Printf("[Client] Found %d UTXOs for sender: %s", len(senderUTXOs), senderPubKey)

		var recipientPubKey string
		for k := range keyMap {
			if k != senderPubKey {
				recipientPubKey = k
				break
			}
		}

		logger.InfoLogger.Printf("[Client] Selected recipient: %s for sender: %s", recipientPubKey, senderPubKey)

		tx, err := createTransaction(senderPubKey, keyMap[senderPubKey], recipientPubKey)
		if err != nil {
			logger.WarnLogger.Printf("[Client] Failed to create transaction from %s to %s: %v", senderPubKey, recipientPubKey, err)
			continue
		}
		logger.InfoLogger.Printf("[Client] Successfully created transaction from %s to %s.", senderPubKey, recipientPubKey)

		minerIP := minerIPs[rand.Intn(len(minerIPs))]
		logger.InfoLogger.Printf("[Client] Sending transaction to miner at IP: %s", minerIP)

		if err := sendTransactionToMiner(server.ConvertTransactionToGrpc(tx), minerIP+":50051"); err != nil {
			logger.WarnLogger.Printf("[Client] Failed to send transaction from %s to %s to miner %s: %v", senderPubKey, recipientPubKey, minerIP, err)
		} else {
			logger.InfoLogger.Printf("[Client] Transaction from %s to %s successfully sent to miner: %s", senderPubKey, recipientPubKey, minerIP)

			err = utxoSet.AddTransaction(*tx)
			if err != nil {
				logger.WarnLogger.Printf("[Client] Failed to update UTXOSet after transaction from %s to %s: %v", senderPubKey, recipientPubKey, err)
				continue
			}
			logger.InfoLogger.Printf("[Client] Updated UTXOSet after transaction from %s to %s", senderPubKey, recipientPubKey)
		}
	}
	logger.InfoLogger.Println("[Client] Completed transaction creation and sending process.")
}

func createTransaction(senderPubKey, privateKey, receiverPubKey string) (*blockchain.Transaction, error) {
	logger.InfoLogger.Printf("[Client] Creating transaction from sender: %s to recipient: %s", senderPubKey, receiverPubKey)

	senderAddress := crypto.Key2Addr(senderPubKey)
	receiverAddress := crypto.Key2Addr(receiverPubKey)

	amountToSend := rand.Int63n(5) + 1

	selectedUTXOs, err := utxoSet.GetUTXOs(senderAddress, amountToSend)
	if err != nil {
		return nil, fmt.Errorf("failed to select UTXOs: %w", err)
	}

	tx, err := blockchain.NewTransaction(selectedUTXOs, senderPubKey, receiverPubKey, amountToSend)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	if err := tx.Sign(privateKey); err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	logger.InfoLogger.Printf("[Client] Successfully created transaction from %s to %s for amount: %d", senderAddress, receiverAddress, amountToSend)

	return tx, nil
}

func sendTransactionToMiner(tx *gen.Transaction, minerIP string) error {
	logger.InfoLogger.Println("[Client] Sending transaction to miner:", minerIP)

	conn, err := grpc.Dial(minerIP, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to miner at %s: %w", minerIP, err)
	}
	defer conn.Close()

	client := gen.NewIncomingCommunicatorServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.SubmitTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	logger.InfoLogger.Println("[Client] Transaction successfully submitted to miner:", minerIP)
	return nil
}
