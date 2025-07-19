package sval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailRules(t *testing.T) {
	tests := []struct {
		name     string
		rules    *EmailRules
		input    any
		wantErr  bool
		expected error
	}{
		{
			name: "valid email",
			rules: &EmailRules{
				BaseRules: BaseRules{Required: true},
			},
			input:    "valid.email@example.com",
			wantErr:  false,
			expected: nil,
		},
		{
			name: "empty email when not required",
			rules: &EmailRules{
				BaseRules: BaseRules{Required: false},
			},
			input:    "",
			wantErr:  false,
			expected: nil,
		},
		{
			name: "empty email ptr when not required",
			rules: &EmailRules{
				BaseRules: BaseRules{Required: false},
			},
			input:    nil,
			wantErr:  false,
			expected: nil,
		},
		{
			name: "valid email with RFC 5322 strategy",
			rules: &EmailRules{
				BaseRules: BaseRules{Required: true},
			},
			input:    "valid.email.rfc@example.com",
			wantErr:  false,
			expected: nil,
		},
		{
			name: "valid email with allowed domain",
			rules: &EmailRules{
				BaseRules:      BaseRules{Required: true},
				AllowedDomains: []string{"example.com"},
			},
			input:    "valid.email.allowed.domain@example.com",
			wantErr:  false,
			expected: nil,
		},
		{
			name: "valid email with excluded domain",
			rules: &EmailRules{
				BaseRules:       BaseRules{Required: true},
				ExcludedDomains: []string{"excluded.com"},
			},
			input:    "valid.email.excluded.domain@example.com",
			wantErr:  false,
			expected: nil,
		},
		{
			name: "invalid email rfc 5322 - @ at the end",
			rules: &EmailRules{
				BaseRules: BaseRules{Required: true},
				Strategy:  string(RFC5322),
			},
			input:   "invalid.email@example.com@",
			wantErr: true,
			expected: func() error {
				err := NewValidationError()
				err.AddError(EmailRuleNameStrategy, string(RFC5322), "invalid.email@example.com@", "email does not conform to chosen strategy")
				return err
			}(),
		},
		{
			name: "invalid email rfc 5322 - len < 3",
			rules: &EmailRules{
				BaseRules: BaseRules{Required: true},
			},
			input:    "a@",
			wantErr:  false,
			expected: nil,
		},
		{
			name: "valid email",
			rules: &EmailRules{
				BaseRules: BaseRules{Required: true},
			},
			input:    "valid.email@example.com",
			wantErr:  false,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rules.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for %s with input %v", tt.name, tt.input)
				assert.Equal(t, tt.expected, err, "Unexpected error for %s with input %v", tt.name, tt.input)
			} else {
				assert.NoError(t, err, "Expected no error for %s with input %v", tt.name, tt.input)
			}
		})
	}
}
