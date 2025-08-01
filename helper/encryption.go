package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func Hash256Key(privkey string) string {
	sha256Hasher := sha256.New()
	sha256Hasher.Write([]byte(privkey))
	hashedPassword := sha256Hasher.Sum(nil)

	// Convert the SHA-256 hash to a string for bcrypt
	hashedPasswordStr := hex.EncodeToString(hashedPassword)

	return string(hashedPasswordStr)
}

func HashBcryptKey(privkey string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(privkey), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash), err
}

func ValidateAPIKey(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetPrivateKey(pathKey string) (privateKey string) {
	file, err := os.Open(pathKey)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	privateKeyBytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	privateKey = string(privateKeyBytes)
	privateKey = strings.TrimRightFunc(privateKey, unicode.IsSpace)

	return
}
