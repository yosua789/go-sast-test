package helper

import "regexp"

func IsValidUsername(username string) bool {
	// Define the regular expression pattern
	pattern := `^[a-zA-Z][a-zA-Z\s'.]*$`
	// Compile the regular expression
	re := regexp.MustCompile(pattern)
	// Check if the input string matches the pattern
	return re.MatchString(username)
}
