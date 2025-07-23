package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func InitCustomValidator(validate *validator.Validate) {
	validate.RegisterValidation("not_blank", notBlank)
	validate.RegisterValidation("custom_email", validateEmail)
	validate.RegisterValidation("custom_phone_number", validatePhoneNumber)
	validate.RegisterValidation("alphaunicodespaces", validateAlphaUnicodeWithSpace)
}

func notBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func validateEmail(fl validator.FieldLevel) bool {
	// Regular expression for email validation
	var re = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	var email = fl.Field().String()

	if !re.MatchString(email) {
		return false
	}

	if strings.Contains(email, "..") {
		return false
	}

	return true
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
	// Regex allows + for country codes, digits, spaces, hyphens, and parentheses, but ensures max 15 digits
	re := regexp.MustCompile(`^\+?[0-9]{7,16}$`)
	return re.MatchString(fl.Field().String())
}

var alphaUnicodeWithSpaceRegex = regexp.MustCompile(`^[\p{L} ]+$`)

func validateAlphaUnicodeWithSpace(fl validator.FieldLevel) bool {
	return alphaUnicodeWithSpaceRegex.MatchString(fl.Field().String())
}
