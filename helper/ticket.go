package helper

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
)

const PREFIX_TICKET_NUMBER = "TKT"

func GenerateTicketNumber(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, RandomUpperAlphaNumeric(10))
}

// Returning 50 character
func GenerateTicketCode() (string, error) {
	const targetLength = 50
	const entropyBytes = 32 // start with 32 bytes

	for {
		raw := make([]byte, entropyBytes)
		_, err := rand.Read(raw)
		if err != nil {
			return "", err
		}

		code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
		if len(code) >= targetLength {
			return code[:targetLength], nil
		}

		// Jika terlalu pendek, tambahkan entropi dan ulang
		// (opsional: entropyBytes += 1)
	}
}
