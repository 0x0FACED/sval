package sval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordRules(t *testing.T) {
	tests := []struct {
		name    string
		rules   PasswordRules
		value   interface{}
		wantErr bool
	}{
		// Basic validation tests
		{
			name:    "empty password when not required",
			rules:   PasswordRules{},
			value:   "",
			wantErr: false,
		},
		{
			name:    "empty password when required",
			rules:   PasswordRules{BaseRules: BaseRules{Required: true}},
			value:   "",
			wantErr: true,
		},
		{
			name:    "non-string value",
			rules:   PasswordRules{},
			value:   123,
			wantErr: true,
		},
		{
			name:    "nil value when not required",
			rules:   PasswordRules{},
			value:   nil,
			wantErr: false,
		},
		{
			name:    "nil value when required",
			rules:   PasswordRules{BaseRules: BaseRules{Required: true}},
			value:   nil,
			wantErr: true,
		},
		// Length validation tests
		{
			name:    "password shorter than minimum",
			rules:   PasswordRules{MinLen: 8},
			value:   "short",
			wantErr: true,
		},
		{
			name:    "password longer than maximum",
			rules:   PasswordRules{MaxLen: 10},
			value:   "thispasswordiswaytoolong",
			wantErr: true,
		},
		{
			name:    "password within length limits",
			rules:   PasswordRules{MinLen: 8, MaxLen: 16},
			value:   "goodpass123",
			wantErr: false,
		},
		// Character type tests
		{
			name:    "not enough uppercase letters",
			rules:   PasswordRules{MinUpper: 2},
			value:   "only1Upper",
			wantErr: true,
		},
		{
			name:    "sufficient uppercase letters",
			rules:   PasswordRules{MinUpper: 2},
			value:   "TwoUppers",
			wantErr: false,
		},
		{
			name:    "not enough lowercase letters",
			rules:   PasswordRules{MinLower: 3},
			value:   "NO",
			wantErr: true,
		},
		{
			name:    "sufficient lowercase letters",
			rules:   PasswordRules{MinLower: 3},
			value:   "hasLowercase",
			wantErr: false,
		},
		{
			name:    "not enough numbers",
			rules:   PasswordRules{MinNumbers: 2},
			value:   "only1number",
			wantErr: true,
		},
		{
			name:    "sufficient numbers",
			rules:   PasswordRules{MinNumbers: 2},
			value:   "has2numbers12",
			wantErr: false,
		},
		{
			name:    "not enough special characters",
			rules:   PasswordRules{MinSpecial: 2},
			value:   "only!one",
			wantErr: true,
		},
		{
			name:    "sufficient special characters",
			rules:   PasswordRules{MinSpecial: 2},
			value:   "has@two#special",
			wantErr: false,
		},
		// Special characters validation
		{
			name:    "using disallowed special character",
			rules:   PasswordRules{DisallowedChars: []rune{' ', '<', '>'}},
			value:   "contains space here",
			wantErr: true,
		},
		{
			name:    "using allowed special characters",
			rules:   PasswordRules{SpecialChars: []rune{'@', '#', '$'}},
			value:   "using@allowed#chars",
			wantErr: false,
		},
		{
			name:    "using non-allowed special characters",
			rules:   PasswordRules{SpecialChars: []rune{'@', '#', '$'}},
			value:   "using%notallowed^chars",
			wantErr: true,
		},
		// Pattern detection tests
		// {
		// 	name:    "contains linear pattern",
		// 	rules:   PasswordRules{DetectLinearPatterns: true},
		// 	value:   "pass123456",
		// 	wantErr: true,
		// },
		// {
		// 	name:    "no linear patterns",
		// 	rules:   PasswordRules{DetectLinearPatterns: true},
		// 	value:   "random135pass",
		// 	wantErr: false,
		// },
		// Repeating characters tests
		{
			name:    "too many repeating characters",
			rules:   PasswordRules{MaxRepeatRun: 2},
			value:   "passsword",
			wantErr: true,
		},
		{
			name:    "acceptable repeating characters",
			rules:   PasswordRules{MaxRepeatRun: 2},
			value:   "password",
			wantErr: false,
		},
		// Complex rules combination tests
		{
			name: "complex password rules - valid password",
			rules: PasswordRules{
				MinLen:               12,
				MaxLen:               20,
				MinUpper:             2,
				MinLower:             2,
				MinNumbers:           2,
				MinSpecial:           2,
				MaxRepeatRun:         2,
				DetectLinearPatterns: true,
				SpecialChars:         []rune{'!', '@', '#', '$', '%'},
			},
			value:   "SecureP@ss#12$",
			wantErr: false,
		},
		{
			name: "complex password rules - invalid password",
			rules: PasswordRules{
				MinLen:               12,
				MaxLen:               20,
				MinUpper:             2,
				MinLower:             2,
				MinNumbers:           2,
				MinSpecial:           2,
				MaxRepeatRun:         2,
				DetectLinearPatterns: true,
				SpecialChars:         []rune{'!', '@', '#', '$', '%'},
			},
			value:   "Weak1@",
			wantErr: true,
		},
		// Blacklist tests
		{
			name: "blacklisted password",
			rules: PasswordRules{
				Blacklist: []string{"password123", "admin123", "qwerty123"},
			},
			value:   "password123",
			wantErr: true,
		},
		{
			name: "not blacklisted password",
			rules: PasswordRules{
				Blacklist: []string{"password123", "admin123", "qwerty123"},
			},
			value:   "mySecurePass123!",
			wantErr: false,
		},
		// Entropy test
		{
			name: "low entropy password",
			rules: PasswordRules{
				MinEntropy: 3.0,
			},
			value:   "aaaaaaa",
			wantErr: true,
		},
		{
			name: "high entropy password",
			rules: PasswordRules{
				MinEntropy: 3.0,
			},
			value:   "Tr0ub4dour&3",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rules.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for %s with value %v", tt.name, tt.value)
			} else {
				assert.NoError(t, err, "Unexpected error for %s with value %v", tt.name, tt.value)
			}
		})
	}
}
