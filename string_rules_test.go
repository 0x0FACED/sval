package sval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringRules(t *testing.T) {
	testStr := "test"
	tests := []struct {
		name    string
		rules   StringRules
		value   interface{}
		wantErr bool
	}{
		// Basic validation tests
		{
			name:    "empty string when not required",
			rules:   StringRules{},
			value:   "",
			wantErr: false,
		},
		{
			name:    "empty string when required",
			rules:   StringRules{BaseRules: BaseRules{Required: true}},
			value:   "",
			wantErr: true,
		},
		{
			name:    "non-string value",
			rules:   StringRules{},
			value:   123,
			wantErr: true,
		},
		{
			name:    "nil value when not required",
			rules:   StringRules{},
			value:   nil,
			wantErr: false,
		},
		{
			name:    "nil value when required",
			rules:   StringRules{BaseRules: BaseRules{Required: true}},
			value:   nil,
			wantErr: true,
		},
		{
			name:    "nil pointer when not required",
			rules:   StringRules{},
			value:   (*string)(nil),
			wantErr: false,
		},
		{
			name:    "nil pointer when required",
			rules:   StringRules{BaseRules: BaseRules{Required: true}},
			value:   (*string)(nil),
			wantErr: true,
		},
		{
			name:    "valid string pointer",
			rules:   StringRules{},
			value:   &testStr,
			wantErr: false,
		},

		// Length validation tests
		{
			name:    "string shorter than MinLen",
			rules:   StringRules{MinLen: 10},
			value:   "short",
			wantErr: true,
		},
		{
			name:    "string equal to MinLen",
			rules:   StringRules{MinLen: 4},
			value:   "test",
			wantErr: false,
		},
		{
			name:    "string longer than MaxLen",
			rules:   StringRules{MaxLen: 3},
			value:   "toolong",
			wantErr: true,
		},
		{
			name:    "string equal to MaxLen",
			rules:   StringRules{MaxLen: 4},
			value:   "test",
			wantErr: false,
		},
		{
			name:    "string within MinLen and MaxLen",
			rules:   StringRules{MinLen: 2, MaxLen: 6},
			value:   "test",
			wantErr: false,
		},
		{
			name:    "UTF-8 string length validation",
			rules:   StringRules{MinLen: 2, MaxLen: 4},
			value:   "привет",
			wantErr: true,
		},

		// Regex validation tests
		{
			name: "string matches regex",
			rules: StringRules{
				Regex: stringPtr(`^[a-z]+$`),
			},
			value:   "test",
			wantErr: false,
		},
		{
			name: "string does not match regex",
			rules: StringRules{
				Regex: stringPtr(`^[0-9]+$`),
			},
			value:   "test",
			wantErr: true,
		},

		// Only digits tests
		{
			name:    "only digits - valid",
			rules:   StringRules{OnlyDigits: true},
			value:   "12345",
			wantErr: false,
		},
		{
			name:    "only digits - invalid",
			rules:   StringRules{OnlyDigits: true},
			value:   "123abc",
			wantErr: true,
		},

		// Only letters tests
		{
			name:    "only letters - valid",
			rules:   StringRules{OnlyLetters: true},
			value:   "abcDEF",
			wantErr: false,
		},
		{
			name:    "only letters - invalid",
			rules:   StringRules{OnlyLetters: true},
			value:   "abc123",
			wantErr: true,
		},

		// No whitespace tests
		{
			name:    "no whitespace - valid",
			rules:   StringRules{NoWhitespace: true},
			value:   "nowhitespace",
			wantErr: false,
		},
		{
			name:    "no whitespace - invalid space",
			rules:   StringRules{NoWhitespace: true},
			value:   "has space",
			wantErr: true,
		},
		{
			name:    "no whitespace - invalid tab",
			rules:   StringRules{NoWhitespace: true},
			value:   "has\ttab",
			wantErr: true,
		},

		// Trim space tests
		{
			name:    "trim space - leading and trailing",
			rules:   StringRules{TrimSpace: true},
			value:   "  test  ",
			wantErr: false,
		},

		// Starts with tests
		{
			name: "starts with - valid",
			rules: StringRules{
				StartsWith: stringPtr("test"),
			},
			value:   "test_string",
			wantErr: false,
		},
		{
			name: "starts with - invalid",
			rules: StringRules{
				StartsWith: stringPtr("test"),
			},
			value:   "not_test",
			wantErr: true,
		},

		// Ends with tests
		{
			name: "ends with - valid",
			rules: StringRules{
				EndsWith: stringPtr("test"),
			},
			value:   "string_test",
			wantErr: false,
		},
		{
			name: "ends with - invalid",
			rules: StringRules{
				EndsWith: stringPtr("test"),
			},
			value:   "test_not",
			wantErr: true,
		},

		// Contains tests
		{
			name: "contains - all substrings present",
			rules: StringRules{
				Contains: []string{"test", "ing"},
			},
			value:   "testing_string",
			wantErr: false,
		},
		{
			name: "contains - missing substring",
			rules: StringRules{
				Contains: []string{"test", "foo"},
			},
			value:   "testing_string",
			wantErr: false,
		},

		// Not contains tests
		{
			name: "not contains - no forbidden substrings",
			rules: StringRules{
				NotContains: []string{"foo", "bar"},
			},
			value:   "testing_string",
			wantErr: false,
		},
		{
			name: "not contains - has forbidden substring",
			rules: StringRules{
				NotContains: []string{"test", "foo"},
			},
			value:   "testing_string",
			wantErr: true,
		},

		// One of tests
		{
			name: "one of - valid option",
			rules: StringRules{
				OneOf: []string{"option1", "option2", "option3"},
			},
			value:   "option2",
			wantErr: false,
		},
		{
			name: "one of - invalid option",
			rules: StringRules{
				OneOf: []string{"option1", "option2", "option3"},
			},
			value:   "option4",
			wantErr: true,
		},

		// Min entropy tests
		{
			name: "min entropy - high entropy string",
			rules: StringRules{
				MinEntropy: 2.0,
			},
			value:   "Abcd123!@#",
			wantErr: false,
		},
		{
			name: "min entropy - low entropy string",
			rules: StringRules{
				MinEntropy: 2.0,
			},
			value:   "aaaaaaa",
			wantErr: true,
		},

		// Combined rules tests
		{
			name: "multiple valid rules",
			rules: StringRules{
				MinLen:       4,
				MaxLen:       10,
				OnlyLetters:  true,
				NoWhitespace: true,
			},
			value:   "TestStr",
			wantErr: false,
		},
		{
			name: "multiple rules with one failure",
			rules: StringRules{
				MinLen:       4,
				MaxLen:       10,
				OnlyLetters:  true,
				NoWhitespace: true,
			},
			value:   "Test123",
			wantErr: true,
		},
		{
			name: "complex validation scenario",
			rules: StringRules{
				BaseRules:  BaseRules{Required: true},
				MinLen:     5,
				MaxLen:     20,
				StartsWith: stringPtr("test"),
				EndsWith:   stringPtr("end"),
				Contains:   []string{"mid"},
				TrimSpace:  true,
				MinEntropy: 1.5,
			},
			value:   "test_mid_end",
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

func stringPtr(s string) *string {
	return &s
}
