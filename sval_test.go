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
					Type: "number",
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
					Type: "number",
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
		name    string
		cfg     RuleConfig
		wantErr bool
	}{
		{
			name: "string rules",
			cfg: RuleConfig{
				Type: "string",
				Params: map[string]any{
					"required": true,
					"min_len":  5,
				},
			},
			wantErr: false,
		},
		{
			name: "email rules",
			cfg: RuleConfig{
				Type: "email",
				Params: map[string]any{
					"required": true,
					"rfc":      true,
				},
			},
			wantErr: false,
		},
		{
			name: "number rules",
			cfg: RuleConfig{
				Type: "number",
				Params: map[string]any{
					"required": true,
					"min":      float64(0),
					"max":      float64(100),
				},
			},
			wantErr: false,
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
			_, err := createRuleSet(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err, "Expected error, but got none")
			} else {
				assert.NoError(t, err, "Expected no error, but got one")
			}
		})
	}
}
