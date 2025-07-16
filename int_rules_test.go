package sval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntRules(t *testing.T) {
	min := 0
	max := 100
	testCases := []struct {
		name     string
		rules    IntRules
		input    any
		wantErr  bool
		expected error
	}{
		{
			name: "valid int within range",
			rules: IntRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Min: &min,
				Max: &max,
			},
			input:    42,
			wantErr:  false,
			expected: nil,
		},
		{
			name: "nil input with required",
			rules: IntRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Min: nil,
				Max: nil,
			},
			input:   nil,
			wantErr: true,
			expected: func() error {
				err := NewValidationError()
				err.AddError(BaseRuleNameRequired, true, FieldIsRequired)
				return err
			}(),
		},
		{
			name: "value below minimum",
			rules: IntRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Min: &min,
				Max: nil,
			},
			input:   -1,
			wantErr: true,
			expected: func() error {
				err := NewValidationError()
				err.AddError(IntRuleNameMin, min, "value must be greater than or equal to min")
				return err
			}(),
		},
		{
			name: "value above maximum",
			rules: IntRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Min: nil,
				Max: &max,
			},
			input:   101,
			wantErr: true,
			expected: func() error {
				err := NewValidationError()
				err.AddError(IntRuleNameMax, max, "value must be less than or equal to max")
				return err
			}(),
		},
		{
			name: "invalid type",
			rules: IntRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Min: nil,
				Max: nil,
			},
			input:   "not an int",
			wantErr: true,
			expected: func() error {
				err := NewValidationError()
				err.AddError(BaseRuleNameType, TypeInt, "value must be int")
				return err
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.rules.Validate(tc.input)
			if tc.wantErr {
				assert.Error(t, err, "Expected error but got none")
				if tc.expected != nil {
					assert.IsType(t, tc.expected, err, "Error type does not match")
					assert.Equal(t, tc.expected, err, "Errors do not match")
				}
			} else {
				assert.NoError(t, err, "Got unexpected error")
			}
		})
	}
}
