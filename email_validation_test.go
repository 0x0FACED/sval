package sval

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		strategy EmailValidationStrategy
		want     bool
	}{
		// RFC 5321 (SMTP) tests
		{
			name:     "RFC5321: simple valid email",
			email:    "user@example.com",
			strategy: RFC5321,
			want:     true,
		},
		{
			name:     "RFC5321: valid with allowed special chars",
			email:    "user.name+tag@example.com",
			strategy: RFC5321,
			want:     true,
		},
		{
			name:     "RFC5321: invalid quoted string",
			email:    "\"user name\"@example.com",
			strategy: RFC5321,
			want:     false,
		},
		{
			name:     "RFC5321: invalid domain without dot",
			email:    "user@localhost",
			strategy: RFC5321,
			want:     false,
		},
		{
			name:     "RFC5321: invalid TLD with number",
			email:    "user@example.c0m",
			strategy: RFC5321,
			want:     false,
		},

		// HTML5 tests
		{
			name:     "HTML: simple valid email",
			email:    "user@example.com",
			strategy: HTMLInput,
			want:     true,
		},
		{
			name:     "HTML: valid with numbers in domain",
			email:    "user@sub2.example123.com",
			strategy: HTMLInput,
			want:     true,
		},
		{
			name:     "HTML: invalid spaces",
			email:    "user name@example.com",
			strategy: HTMLInput,
			want:     false,
		},
		{
			name:     "HTML: valid with dots",
			email:    "user.name@example.com",
			strategy: HTMLInput,
			want:     true,
		},
		{
			name:     "HTML: valid with plus",
			email:    "user+tag@example.com",
			strategy: HTMLInput,
			want:     true,
		},

		// common invalid cases for all strategies
		{
			name:     "Common: empty string",
			email:    "",
			strategy: RFC5321,
			want:     false,
		},
		{
			name:     "Common: missing @",
			email:    "userexample.com",
			strategy: RFC5322,
			want:     false,
		},
		{
			name:     "Common: multiple @",
			email:    "user@domain@example.com",
			strategy: HTMLInput,
			want:     false,
		},
		{
			name:     "Common: too long email",
			email:    "user@" + strings.Repeat("a", 250) + ".com",
			strategy: RFC5321,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateEmail(tt.email, tt.strategy)
			assert.Equal(t, tt.want, got, "ValidateEmail() for strategy %s", tt.strategy)
		})
	}
}
