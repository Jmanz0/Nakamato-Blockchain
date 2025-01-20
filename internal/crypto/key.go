package crypto

import (
	"crypto/sha256"
	"encoding/base64"
)

func Key2Addr(pubKey string) string {
	// Decode Base64 public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return "Invalid public key format"
	}

	// Step 1: SHA256 hash
	hasher := sha256.New()
	hasher.Write(pubKeyBytes)
	hash := hasher.Sum(nil)

	// Step 2: Take first 20 bytes
	shortHash := hash[:20]

	// Step 3: Add version byte (0x00 for mainnet)
	versionedHash := append([]byte{0x00}, shortHash...)

	// Step 4: Encode in Base64 for simplicity
	address := base64.StdEncoding.EncodeToString(versionedHash)

	return address
}
