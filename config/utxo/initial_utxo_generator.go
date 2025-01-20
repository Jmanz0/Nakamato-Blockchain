package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"nakamoto-blockchain/internal/crypto"
)

type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type UTXO struct {
	TxID    string `json:"TxID"`
	Index   int    `json:"Index"`
	Amount  int64  `json:"Amount"`
	Address string `json:"Address"`
}

type UTXOEntry struct {
	PublicKey string `json:"public_key"`
	UTXOs     []UTXO `json:"utxos"`
}

func main() {
	keysFile := flag.String("keys", "keys.json", "Path to the keys JSON file")
	outputDir := flag.String("output", ".", "Path to the output UTXOs Dir file")
	numUTXOsPerKey := flag.Int("utxos", 3, "Number of UTXOs to generate per key")
	maxAmount := flag.Int64("max", 10000, "Maximum amount per UTXO")
	minAmount := flag.Int64("min", 1000, "Minimum amount per UTXO")

	flag.Parse()

	outputFile := filepath.Join(*outputDir, "initial_utxos.json")

	keyPairs, err := readKeys(*keysFile)
	if err != nil {
		fmt.Println("Error reading keys:", err)
		return
	}

	initialUTXOs := generateInitialUTXOs(keyPairs, *numUTXOsPerKey, *maxAmount, *minAmount)

	err = saveUTXOs(outputFile, initialUTXOs)
	if err != nil {
		fmt.Println("Error saving UTXOs:", err)
		return
	}

	fmt.Printf("Initial UTXOs have been saved to %s\n", outputFile)
}

func readKeys(filename string) ([]KeyPair, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keyPairs []KeyPair
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&keyPairs)
	if err != nil {
		return nil, err
	}
	return keyPairs, nil
}

func generateInitialUTXOs(keyPairs []KeyPair, numUTXOsPerKey int, maxAmount int64, minAmount int64) []UTXOEntry {
	rand.Seed(time.Now().UnixNano())

	var initialUTXOs []UTXOEntry
	for _, keyPair := range keyPairs {
		address := crypto.Key2Addr(keyPair.PublicKey)
		var utxos []UTXO
		for i := 0; i < numUTXOsPerKey; i++ {
			amount := rand.Int63n(maxAmount-minAmount+1) + minAmount
			utxo := UTXO{
				TxID:    generateTxID(),
				Index:   i,
				Amount:  amount,
				Address: address,
			}
			utxos = append(utxos, utxo)
		}
		initialUTXOs = append(initialUTXOs, UTXOEntry{
			PublicKey: keyPair.PublicKey,
			UTXOs:     utxos,
		})
	}
	return initialUTXOs
}

func generateTxID() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}

func saveUTXOs(filename string, utxoEntries []UTXOEntry) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(utxoEntries)
}
