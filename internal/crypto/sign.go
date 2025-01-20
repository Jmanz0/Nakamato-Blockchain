package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"math/big"
)

// Sign generates a signature for the given hash using the provided private key in string format.
func Sign(hash string, privateKeyStr string) (string, error) {
	privateKey, err := parsePrivateKey(privateKeyStr)
	if err != nil {
		return "", err
	}

	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashBytes)
	if err != nil {
		return "", err
	}

	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)

	return hex.EncodeToString(signature), nil
}

// VerifySignature checks if the provided signature is valid for the given hash and public key in string format.
func VerifySignature(hash, signature string, publicKeyStr string) (bool, error) {
	publicKey, err := parsePublicKey(publicKeyStr)
	if err != nil {
		return false, err
	}

	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	rLen := len(sigBytes) / 2
	r := new(big.Int).SetBytes(sigBytes[:rLen])
	s := new(big.Int).SetBytes(sigBytes[rLen:])

	return ecdsa.Verify(publicKey, hashBytes, r, s), nil
}

// parsePrivateKey converts a base64 encoded private key string to an ecdsa.PrivateKey.
func parsePrivateKey(privateKeyStr string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, ErrInvalidKey
	}

	privateKey, err := x509.ParseECPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// parsePublicKey converts a base64 encoded public key string to an ecdsa.PublicKey.
func parsePublicKey(publicKeyStr string) (*ecdsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, ErrInvalidKey
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, err
	}

	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, ErrInvalidKey
	}

	return ecdsaPublicKey, nil
}

var ErrInvalidKey = x509.UnknownAuthorityError{}
