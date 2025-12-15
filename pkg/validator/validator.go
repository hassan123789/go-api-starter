package validator

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(strings.TrimSpace(email))
}

// ValidatePassword validates a password
// Password must be at least 8 characters
func ValidatePassword(password string) bool {
	return len(password) >= 8
}

// ValidateTodoTitle validates a todo title
// Title must be non-empty and at most 255 characters
func ValidateTodoTitle(title string) bool {
	trimmed := strings.TrimSpace(title)
	return len(trimmed) > 0 && len(trimmed) <= 255
}
