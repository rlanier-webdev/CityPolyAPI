package utils

import (
	"errors"
	"strings"
	"unicode"
)

func ValidatePassword(password string) error {
	var errs []string

	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if !hasUpper {
		errs = append(errs, "must contain at least one uppercase letter")
	}
	if !hasLower {
		errs = append(errs, "must contain at least one lowercase letter")
	}
	if !hasNumber {
		errs = append(errs, "must contain at least one number")
	}
	if !hasSymbol {
		errs = append(errs, "must contain at least one symbol (e.g. !@#$%)")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
