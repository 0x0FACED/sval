package sval

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() error
		cleanup func()
		wantErr bool
	}{
		{
			name: "yaml config",
			setup: func() error {
				return os.WriteFile("sval.yaml",
					[]byte(`
rules:
  test.field:
    type: string
    params:
      required: true
`), 0644)
			},
			cleanup: func() {
				os.Remove("sval.yaml")
			},
			wantErr: false,
		},
		{
			name: "json config",
			setup: func() error {
				return os.WriteFile("sval.json",
					[]byte(`{
"rules": {
	"test.field": {
		"type": "string",
			"params": {
				"required": true
			}
	}
}
}`), 0644)
			},
			cleanup: func() {
				os.Remove("sval.json")
			},
			wantErr: false,
		},
		{
			name: "invalid yaml",
			setup: func() error {
				return os.WriteFile("sval.yaml", []byte(`invalid yaml`), 0644)
			},
			cleanup: func() {
				os.Remove("sval.yaml")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			if tt.setup != nil {
				err := tt.setup()
				assert.NoError(t, err, "setup must not be failed")
			}

			loader := DefaultConfigLoader()
			assert.NotNil(t, loader, "Loader must not be nil")

			config, err := loader.Load()
			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}

			if err == nil {
				assert.NotNil(t, config.Rules, "Rules must not be nil")
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	type TestStruct struct {
		Name    string `sval:"name"`
		Age     *int   `sval:"age"`
		Email   string `sval:"email"`
		Details struct {
			Address string `sval:"address"`
		} `sval:"details"`
	}

	type TestSlice struct {
		Items []struct {
			Value int `sval:"value"`
		} `sval:"items"`
	}

	tests := []struct {
		name      string
		rules     map[string]RuleConfig
		data      any
		wantError bool
	}{
		{
			name: "valid struct",
			rules: map[string]RuleConfig{
				"name": {
					Type: "string",
					Params: map[string]any{
						"required": true,
						"min_len":  3,
					},
				},
			},
			data: TestStruct{
				Name: "John",
			},
			wantError: false,
		},
		{
			name: "required field missing",
			rules: map[string]RuleConfig{
				"name": {
					Type: "string",
					Params: map[string]any{
						"required": true,
					},
				},
			},
			data: TestStruct{
				Name: "",
			},
			wantError: true,
		},
		{
			name: "nil pointer with required",
			rules: map[string]RuleConfig{
				"age": {
					Type: "int",
					Params: map[string]any{
						"required": true,
					},
				},
			},
			data: TestStruct{
				Age: nil,
			},
			wantError: true,
		},
		{
			name: "nested struct validation",
			rules: map[string]RuleConfig{
				"details.address": {
					Type: "string",
					Params: map[string]any{
						"required": true,
					},
				},
			},
			data: TestStruct{
				Details: struct {
					Address string `sval:"address"`
				}{
					Address: "",
				},
			},
			wantError: true,
		},
		{
			name: "slice validation",
			rules: map[string]RuleConfig{
				"items[].value": {
					Type: "int",
					Params: map[string]any{
						"min": 0,
						"max": 100,
					},
				},
			},
			data: TestSlice{
				Items: []struct {
					Value int `sval:"value"`
				}{
					{Value: -1},  // Invalid: < min
					{Value: 0},   // Valid
					{Value: 101}, // Invalid: > max
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := NewValidatorFromConfig(ValidatorConfig{
				Rules: tt.rules,
			})
			assert.NoError(t, err)

			err = v.Validate(tt.data)
			if tt.wantError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one: %v", err)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{
			path: "users[0].name",
			want: "users[].name",
		},
		{
			path: "users[123].addresses[456].street",
			want: "users[].addresses[].street",
		},
		{
			path: "simple.path",
			want: "simple.path",
		},
		{
			path: "",
			want: "",
		},
		{
			path: "array[0]",
			want: "array[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := normalizePath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateRuleSet(t *testing.T) {
	tests := []struct {
		name      string
		cfg       RuleConfig
		wantErr   bool
		wantRules RuleSet
	}{
		{
			name: "string rules",
			cfg: RuleConfig{
				Type: "string",
				Params: map[string]any{
					"required":      true,
					"min_len":       5,
					"max_len":       15,
					"regex":         "^[a-zA-Z0-9]+$",
					"only_digits":   true,
					"only_letters":  false,
					"no_whitespace": true,
					"trim_space":    true,
					"starts_with":   "test",
					"ends_with":     "end",
					"contains":      []any{"example", "test"},
					"not_contains":  []any{"invalid"},
					"one_of":        []any{"option1", "option2"},
					"min_entropy":   2.0,
				},
			},
			wantErr: false,
			wantRules: &StringRules{
				BaseRules: BaseRules{
					Required: true,
				},
				MinLen:       5,
				MaxLen:       15,
				OnlyDigits:   true,
				OnlyLetters:  false,
				NoWhitespace: true,
				TrimSpace:    true,
				StartsWith:   ptr("test"),
				EndsWith:     ptr("end"),
				Regex:        ptr("^[a-zA-Z0-9]+$"),
				Contains:     []string{"example", "test"},
				NotContains:  []string{"invalid"},
				OneOf:        []string{"option1", "option2"},
				MinEntropy:   2.0,
			},
		},
		{
			name: "email rules",
			cfg: RuleConfig{
				Type: "email",
				Params: map[string]any{
					"required":         true,
					"strategy":         "rfc5322",
					"min_domain_len":   3,
					"excluded_domains": []string{"example.com", "test.com"},
					"allowed_domains":  []string{"allowed.com", "example.org"},
					"regex":            "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
				},
			},
			wantErr: false,
			wantRules: &EmailRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Strategy:        "rfc5322",
				MinDomainLen:    3,
				ExcludedDomains: []string{"example.com", "test.com"},
				AllowedDomains:  []string{"allowed.com", "example.org"},
				Regex:           ptr("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"),
			},
		},
		{
			name: "int rules",
			cfg: RuleConfig{
				Type: "int",
				Params: map[string]any{
					"required": true,
					"min":      0,
					"max":      100,
				},
			},
			wantErr: false,
			wantRules: &IntRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Min: ptr(0),
				Max: ptr(100),
			},
		},
		{
			name: "float rules",
			cfg: RuleConfig{
				Type: "float",
				Params: map[string]any{
					"required": true,
					"min":      0.0,
					"max":      100.0,
				},
			},
			wantErr: false,
			wantRules: &FloatRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Min: ptr(0.0),
				Max: ptr(100.0),
			},
		},
		{
			name: "ip rules",
			cfg: RuleConfig{
				Type: "ip",
				Params: map[string]any{
					"required":         true,
					"version":          4,
					"allow_private":    true,
					"allowed_subnets":  []string{"192.168.0.0/16"},
					"excluded_subnets": []string{"172.18.0.0/24"},
				},
			},
			wantErr: false,
			wantRules: &IPRules{
				BaseRules: BaseRules{
					Required: true,
				},
				Version:         4,
				AllowPrivate:    true,
				AllowedSubnets:  []string{"192.168.0.0/16"},
				ExcludedSubnets: []string{"172.18.0.0/24"},
			},
		},
		{
			name: "password rules",
			cfg: RuleConfig{
				Type: "password",
				Params: map[string]any{
					"required":               true,
					"min_len":                8,
					"max_len":                64,
					"min_upper":              2,
					"min_lower":              2,
					"min_digits":             2,
					"min_special":            2,
					"special_chars":          "!@#$%^&*()",
					"allowed_chars":          "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()",
					"disallowed_chars":       " ",
					"max_repeat_run":         3,
					"detect_linear_patterns": true,
					"blacklist":              []string{"password", "123456", "qwerty"},
					"min_entropy":            3.0,
				},
			},
			wantErr: false,
			wantRules: &PasswordRules{
				BaseRules: BaseRules{
					Required: true,
				},
				MinLen:       8,
				MaxLen:       64,
				MinUpper:     2,
				MinLower:     2,
				MinDigits:    2,
				MinSpecial:   2,
				SpecialChars: []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')'},
				AllowedChars: []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
					'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
					'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')'},
				DisallowedChars:      []rune{' '},
				MaxRepeatRun:         3,
				DetectLinearPatterns: true,
				Blacklist:            []string{"password", "123456", "qwerty"},
				MinEntropy:           3.0,
			},
		},
		{
			name: "unknown type",
			cfg: RuleConfig{
				Type: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := createRuleSet(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err, "Expected error, but got none")
			} else {
				assert.NoError(t, err, "Expected no error, but got one")
				assert.NotNil(t, rules, "Expected rules to be created, but got nil")
				if tt.wantRules != nil {

					assert.IsType(t, tt.wantRules, rules, "Expected rules type to match")
					assert.Equal(t, tt.wantRules, rules, "Expected rules to match")
				}
			}
		})
	}
}
