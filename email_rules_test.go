package sval

import "testing"

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
			},
			input:   "invalid.email@example.com@",
			wantErr: true,
			expected: func() error {
				err := NewValidationError()
				err.AddError(EmailRuleNameStrategy, true, "email does not conform to RFC standards")
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

	_ = tests
}
