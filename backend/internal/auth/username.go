package auth

import "strings"

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func IsValidUsername(username string) bool {
	if len(username) < 4 || len(username) > 20 {
		return false
	}

	if username[0] < 'a' || username[0] > 'z' {
		return false
	}

	previousUnderscore := false
	for i := 1; i < len(username); i++ {
		character := username[i]
		isLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		isUnderscore := character == '_'

		if !isLetter && !isDigit && !isUnderscore {
			return false
		}
		if isUnderscore && previousUnderscore {
			return false
		}

		previousUnderscore = isUnderscore
	}

	return !previousUnderscore
}
