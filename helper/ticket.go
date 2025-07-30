package helper

import (
	"fmt"
	"time"
)

const PREFIX_TICKET_NUMBER = "TKT"

func GenerateTicketNumber(prefix string) string {
	// Current timestamp as YYYY
	timestamp := time.Now().Format("20060102150405")

	// Generate a random 6-digit number
	randomID := rng.Intn(900000) + 100000

	return fmt.Sprintf("%s-%s-%d", prefix, timestamp, randomID)
}
