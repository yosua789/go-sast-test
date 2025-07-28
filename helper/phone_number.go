package helper

import "regexp"

func ValidatePhoneNumber(phone string) bool {
	// Regex allows + for country codes, digits, spaces, hyphens, and parentheses, but ensures max 15 digits
	re := regexp.MustCompile(`^\+?[0-9]{7,16}$`)
	return re.MatchString(phone)
}
