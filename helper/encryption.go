package helper

import (
	"crypto/sha256"
	"encoding/hex"
)

func Hash256Key(privkey string) string {
	sha256Hasher := sha256.New()
	sha256Hasher.Write([]byte(privkey))
	hashedPassword := sha256Hasher.Sum(nil)

	// Convert the SHA-256 hash to a string for bcrypt
	hashedPasswordStr := hex.EncodeToString(hashedPassword)

	return string(hashedPasswordStr)
}
