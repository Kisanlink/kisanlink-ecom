package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID generates a unique ID with a prefix.
func GenerateID(prefix string) string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%s_%d", prefix, timestamp)
	}
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s_%d_%s", prefix, timestamp, randomHex)
}

// GenerateUserID generates a unique user ID.
func GenerateUserID() string {
	return GenerateID("usr")
}

// GenerateProductID generates a unique product ID.
func GenerateProductID() string {
	return GenerateID("prd")
}

// GenerateOrderID generates a unique order ID.
func GenerateOrderID() string {
	return GenerateID("ord")
}
