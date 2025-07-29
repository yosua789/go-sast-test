package helper

import (
	"regexp"
	"strings"
)

func IsValidEmail(email string) bool {
	// Regular expression for email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func GetMailName(email string) string {
	emails := strings.Split(email, "@")
	return emails[0]
}
