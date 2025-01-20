package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func main() {
	numKeyPairs := flag.Int("count", 5, "Number of key pairs to generate")
	outputDir := flag.String("output", ".", "Output directory for the generated keys.json file")

	flag.Parse()

	keyPairs := generateMultipleKeyPairs(*numKeyPairs)

	outputFile := filepath.Join(*outputDir, "keys.json")
	err := saveKeyPairsToFile(outputFile, keyPairs)
	if err != nil {
		fmt.Println("Error saving keys to file:", err)
		return
	}

	fmt.Printf("%d key pairs have been generated and saved to %s\n", *numKeyPairs, outputFile)
}

func generateKeyPair() (KeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return KeyPair{}, err
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return KeyPair{}, err
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		PublicKey:  base64.StdEncoding.EncodeToString(publicKeyBytes),
		PrivateKey: base64.StdEncoding.EncodeToString(privateKeyBytes),
	}, nil
}

func generateMultipleKeyPairs(count int) []KeyPair {
	keyPairs := make([]KeyPair, 0, count)
	for i := 0; i < count; i++ {
		keyPair, err := generateKeyPair()
		if err != nil {
			fmt.Println("Error generating key pair:", err)
			continue
		}
		keyPairs = append(keyPairs, keyPair)
	}
	return keyPairs
}

func saveKeyPairsToFile(filename string, keyPairs []KeyPair) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(keyPairs)
}
