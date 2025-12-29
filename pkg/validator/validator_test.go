package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	testCases := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "Valid email",
			email:    "user@example.com",
			expected: true,
		},
		{
			name:     "Valid email with subdomain",
			email:    "user@mail.example.com",
			expected: true,
		},
		{
			name:     "Valid email with plus sign",
			email:    "user+tag@example.com",
			expected: true,
		},
		{
			name:     "Valid email with numbers",
			email:    "user123@example123.com",
			expected: true,
		},
		{
			name:     "Email with leading space (trimmed)",
			email:    " user@example.com",
			expected: true,
		},
		{
			name:     "Email with trailing space (trimmed)",
			email:    "user@example.com ",
			expected: true,
		},
		{
			name:     "Invalid - no @ symbol",
			email:    "userexample.com",
			expected: false,
		},
		{
			name:     "Invalid - no domain",
			email:    "user@",
			expected: false,
		},
		{
			name:     "Invalid - no TLD",
			email:    "user@example",
			expected: false,
		},
		{
			name:     "Invalid - empty",
			email:    "",
			expected: false,
		},
		{
			name:     "Invalid - just whitespace",
			email:    "   ",
			expected: false,
		},
		{
			name:     "Invalid - double @",
			email:    "user@@example.com",
			expected: false,
		},
		{
			name:     "Invalid - spaces in middle",
			email:    "user @example.com",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateEmail(tc.email)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "Valid password - 8 characters",
			password: "12345678",
			expected: true,
		},
		{
			name:     "Valid password - more than 8 characters",
			password: "verysecurepassword123",
			expected: true,
		},
		{
			name:     "Valid password - exactly 8 characters with symbols",
			password: "p@ssw0rd",
			expected: true,
		},
		{
			name:     "Invalid - 7 characters",
			password: "1234567",
			expected: false,
		},
		{
			name:     "Invalid - empty",
			password: "",
			expected: false,
		},
		{
			name:     "Invalid - 1 character",
			password: "a",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidatePassword(tc.password)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateTodoTitle(t *testing.T) {
	testCases := []struct {
		name     string
		title    string
		expected bool
	}{
		{
			name:     "Valid title",
			title:    "Buy groceries",
			expected: true,
		},
		{
			name:     "Valid title - single character",
			title:    "A",
			expected: true,
		},
		{
			name:     "Valid title - 255 characters",
			title:    strings.Repeat("a", 255),
			expected: true,
		},
		{
			name:     "Title with leading/trailing spaces (trimmed)",
			title:    "  Buy groceries  ",
			expected: true,
		},
		{
			name:     "Invalid - empty",
			title:    "",
			expected: false,
		},
		{
			name:     "Invalid - just whitespace",
			title:    "   ",
			expected: false,
		},
		{
			name:     "Invalid - 256 characters",
			title:    strings.Repeat("a", 256),
			expected: false,
		},
		{
			name:     "Invalid - way too long",
			title:    strings.Repeat("a", 1000),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateTodoTitle(tc.title)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Test edge cases
func TestValidateEmailEdgeCases(t *testing.T) {
	// International domains
	assert.True(t, ValidateEmail("user@example.co.uk"))
	assert.True(t, ValidateEmail("user@example.io"))

	// Various TLDs
	assert.True(t, ValidateEmail("user@example.org"))
	assert.True(t, ValidateEmail("user@example.net"))
}

func TestValidatePasswordBoundary(t *testing.T) {
	// Exact boundary
	assert.False(t, ValidatePassword("1234567"))  // 7 chars
	assert.True(t, ValidatePassword("12345678"))  // 8 chars
	assert.True(t, ValidatePassword("123456789")) // 9 chars
}

func TestValidateTodoTitleBoundary(t *testing.T) {
	// Exact boundary at 255
	assert.True(t, ValidateTodoTitle(strings.Repeat("a", 254)))  // 254 chars
	assert.True(t, ValidateTodoTitle(strings.Repeat("a", 255)))  // 255 chars
	assert.False(t, ValidateTodoTitle(strings.Repeat("a", 256))) // 256 chars
}
