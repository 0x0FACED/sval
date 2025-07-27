package sval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockStringer struct {
	str string
}

func (m mockStringer) String() string {
	return m.str
}

func TestConvertToRuneArray(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []rune
		wantErr bool
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name:  "rune array input",
			input: []rune{'a', 'b', 'c'},
			want:  []rune{'a', 'b', 'c'},
		},
		{
			name:  "string array input",
			input: []string{"a", "b", "c"},
			want:  []rune{'a', 'b', 'c'},
		},
		{
			name:  "any array with runes",
			input: []any{'a', 'b', 'c'},
			want:  []rune{'a', 'b', 'c'},
		},
		{
			name:  "any array with strings",
			input: []any{"a", "b", "c"},
			want:  []rune{'a', 'b', 'c'},
		},
		{
			name:  "any array with mixed types",
			input: []any{'a', "b", 99},
			want:  []rune{'a', 'b', 'c'},
		},
		{
			name:  "string to runes",
			input: "abc",
			want:  []rune{'a', 'b', 'c'},
		},
		{
			name:  "int array input",
			input: []int{97, 98, 99},
			want:  []rune{'a', 'b', 'c'},
		},
		{
			name:    "string array with multi-char string",
			input:   []string{"ab"},
			wantErr: true,
		},
		{
			name:    "any array with invalid type",
			input:   []any{true},
			wantErr: true,
		},
		{
			name:    "int array with invalid rune",
			input:   []int{-1},
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   42,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToRuneArray(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestConvertToStringArray(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []string
		wantErr bool
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name:  "string array input",
			input: []string{"foo", "bar"},
			want:  []string{"foo", "bar"},
		},
		{
			name:  "rune array input",
			input: []rune{'a', 'b', 'c'},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "any array with strings",
			input: []any{"foo", "bar"},
			want:  []string{"foo", "bar"},
		},
		{
			name: "any array with stringers",
			input: []any{
				mockStringer{"foo"},
				mockStringer{"bar"},
			},
			want: []string{"foo", "bar"},
		},
		{
			name:    "any array with invalid type",
			input:   []any{42},
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   42,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToStringArray(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
