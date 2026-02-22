package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func BearerToken() (rawToken string, tokenHash string, err error) {
	prefix := "bearer_"

	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Hex encode
	encoded := hex.EncodeToString(bytes)

	// Add prefix
	token := prefix + encoded

	// Create a new hash instance
	hasher := sha256.New()

	// Write the token bytes to the hasher
	hasher.Write([]byte(token))

	// Get the finalized hash result as a byte slice
	tokenHash = hex.EncodeToString(hasher.Sum(nil))

	// Encode the byte slice into a human-readable hexadecimal string
	return token, tokenHash, nil
}