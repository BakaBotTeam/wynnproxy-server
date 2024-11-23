package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256Hash(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
