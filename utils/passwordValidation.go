package utils

import (
	"errors"
	"regexp"
	"unicode"
)

// ValidatePassword valide la force d'un mot de passe
func ValidatePassword(password string) error {
	var errs []string

	// Longueur minimale
	if len(password) < 8 {
		errs = append(errs, "le mot de passe doit contenir au moins 8 caractères")
	}

	// Au moins une lettre majuscule
	hasUpper := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		errs = append(errs, "le mot de passe doit contenir au moins une lettre majuscule")
	}

	// Au moins une lettre minuscule
	hasLower := false
	for _, char := range password {
		if unicode.IsLower(char) {
			hasLower = true
			break
		}
	}
	if !hasLower {
		errs = append(errs, "le mot de passe doit contenir au moins une lettre minuscule")
	}

	// Au moins un chiffre
	hasDigit := false
	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		errs = append(errs, "le mot de passe doit contenir au moins un chiffre")
	}

	// Au moins un caractère spécial
	specialChars := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>?]`)
	if !specialChars.MatchString(password) {
		errs = append(errs, "le mot de passe doit contenir au moins un caractère spécial")
	}

	if len(errs) > 0 {
		return errors.New("mot de passe faible: " + joinStrings(errs, ", "))
	}

	return nil
}

// joinStrings joint des chaînes avec un séparateur
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}

