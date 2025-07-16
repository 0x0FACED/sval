package sval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmailRFC5322(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		// valid
		{
			name:  "simple valid email",
			email: "test@example.com",
			want:  true,
		},
		{
			name:  "email with dots in local part",
			email: "user.name@example.com",
			want:  true,
		},
		{
			name:  "email with special chars in local part",
			email: "user+test@example.com",
			want:  true,
		},
		{
			name:  "email with quoted local part (spaces)",
			email: "\"John Doe\"@example.com",
			want:  true,
		},
		{
			name:  "email with escaped quotes",
			email: "\"quoted\\\"quotes\"@example.com",
			want:  true,
		},
		{
			name:  "email with escaped backslash",
			email: "\"back\\\\slash\"@example.com",
			want:  true,
		},
		{
			name:  "email with quoted local part and dots",
			email: "\"very.(),:;<>[]\\\".VERY.\\\"very@\\\\ \\\"very\\\".unusual\"@strange.example.com",
			want:  true,
		},
		{
			name:  "email with subdomain",
			email: "user@subdomain.example.com",
			want:  true,
		},

		// invalid
		{
			name:  "empty string",
			email: "",
			want:  false,
		},
		{
			name:  "missing @",
			email: "userexample.com",
			want:  false,
		},
		{
			name:  "multiple @",
			email: "user@domain@example.com",
			want:  false,
		},
		{
			name:  "email with quoted special chars",
			email: `\"(),:;<>@[\\]\"@example.com`,
			want:  false,
		},
		{
			name:  "email with all special chars",
			email: `\"very.unusual.@.unusual.com\"@example.com`,
			want:  false,
		},
		{
			name:  "local part too long",
			email: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn@example.com",
			want:  false,
		},
		{
			name:  "domain part missing dot",
			email: "user@examplecom",
			want:  false,
		},
		{
			name:  "domain label starts with hyphen",
			email: "user@-example.com",
			want:  false,
		},
		{
			name:  "domain label ends with hyphen",
			email: "user@example-.com",
			want:  false,
		},
		{
			name:  "consecutive dots in local part",
			email: "user..name@example.com",
			want:  false,
		},
		{
			name:  "local part starts with dot",
			email: ".username@example.com",
			want:  false,
		},
		{
			name:  "local part ends with dot",
			email: "username.@example.com",
			want:  false,
		},
		{
			name:  "quoted string with unescaped quote",
			email: `"test"test"@example.com`,
			want:  false,
		},
		{
			name:  "quoted string with unescaped backslash",
			email: `"test\test"@example.com`,
			want:  false,
		},
		{
			name:  "unclosed quotes",
			email: `"test@example.com`,
			want:  false,
		},
		{
			name:  "quotes in middle of local part",
			email: `before"quoted"after@example.com`,
			want:  false,
		},
		{
			name:  "nested quotes",
			email: `"outer"inner"outer"@example.com`,
			want:  false,
		},
		{
			name:  "quote after escaped quote",
			email: `"quote\""more"@example.com`,
			want:  false,
		},
		{
			name:  "backslash at end of quoted string",
			email: `"ends with\"@example.com`,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateEmailRFC5322(tt.email)
			assert.Equal(t, tt.want, got, "validateEmailRFC5322() for %s", tt.name)
		})
	}
}
