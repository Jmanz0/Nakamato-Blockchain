package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Hash calculates the SHA-256 hash of the provided data and returns it as a hexadecimal string.
func Hash(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}
