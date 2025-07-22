package sval

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMACRules(t *testing.T) {
	tests := []struct {
		name    string
		rules   MACRules
		value   any
		wantErr bool
	}{
		// Basic validation tests
		{
			name:    "empty string when not required",
			rules:   MACRules{},
			value:   "",
			wantErr: false,
		},
		{
			name:    "empty string when required",
			rules:   MACRules{BaseRules: BaseRules{Required: true}},
			value:   "",
			wantErr: true,
		},
		{
			name:    "non-string value",
			rules:   MACRules{},
			value:   123,
			wantErr: true,
		},
		{
			name:    "nil value when not required",
			rules:   MACRules{},
			value:   nil,
			wantErr: false,
		},
		{
			name:    "nil value when required",
			rules:   MACRules{BaseRules: BaseRules{Required: true}},
			value:   nil,
			wantErr: true,
		},

		// Format validation tests
		{
			name: "valid colon format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatColon},
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "invalid colon format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatColon},
			},
			value:   "00-11-22-33-44-55",
			wantErr: true,
		},
		{
			name: "valid hyphen format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatHyphen},
			},
			value:   "00-11-22-33-44-55",
			wantErr: false,
		},
		{
			name: "invalid hyphen format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatHyphen},
			},
			value:   "00.11.22.33.44.55",
			wantErr: true,
		},
		{
			name: "valid dot format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatDot},
			},
			value:   "0011.2233.4455",
			wantErr: false,
		},
		{
			name: "invalid dot format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatDot},
			},
			value:   "00.11.22.33.44.55",
			wantErr: true,
		},
		{
			name: "valid raw format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatRaw},
			},
			value:   "001122334455",
			wantErr: false,
		},
		{
			name: "invalid raw format",
			rules: MACRules{
				Formats: []MACFormat{MACFormatRaw},
			},
			value:   "00:11:22:33:44:55",
			wantErr: true,
		},
		{
			name: "format any accepts all formats",
			rules: MACRules{
				Formats: []MACFormat{MACFormatAny},
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},

		// Case validation tests
		{
			name: "valid lower case",
			rules: MACRules{
				Cases: []MACCase{MACCaseLower},
			},
			value:   "00:11:22:aa:bb:cc",
			wantErr: false,
		},
		{
			name: "invalid lower case",
			rules: MACRules{
				Cases: []MACCase{MACCaseLower},
			},
			value:   "00:11:22:AA:BB:CC",
			wantErr: true,
		},
		{
			name: "valid upper case",
			rules: MACRules{
				Cases: []MACCase{MACCaseUpper},
			},
			value:   "00:11:22:AA:BB:CC",
			wantErr: false,
		},
		{
			name: "invalid upper case",
			rules: MACRules{
				Cases: []MACCase{MACCaseUpper},
			},
			value:   "00:11:22:aA:BB:CC",
			wantErr: true,
		},
		{
			name: "valid camel case",
			rules: MACRules{
				// format must be any or dot
				Formats: []MACFormat{MACFormatDot},
				Cases:   []MACCase{MACCaseCamel},
			},
			value:   "0011.22AA.BBCC",
			wantErr: false,
		},

		// Type validation tests
		{
			name: "valid unicast only",
			rules: MACRules{
				Types: []MACAddressType{MACTypeUnicast},
			},
			value:   "02:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "valid multicast only",
			rules: MACRules{
				Types:          []MACAddressType{MACTypeMulticast},
				AllowMulticast: ptr(true),
			},
			value:   "01:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "valid universal address",
			rules: MACRules{
				Types: []MACAddressType{MACTypeUniversal},
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "valid local address",
			rules: MACRules{
				Types: []MACAddressType{MACTypeLocal},
			},
			value:   "02:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "valid non-zero",
			rules: MACRules{
				AllowZero: ptr(false),
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "valid non-broadcast",
			rules: MACRules{
				AllowBroadcast: ptr(false),
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},

		// OUI validation tests
		{
			name: "valid OUI",
			rules: MACRules{
				OUIWhitelist: []string{"001122"},
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "invalid OUI",
			rules: MACRules{
				OUIWhitelist: []string{"AABBCC"},
			},
			value:   "00:11:22:33:44:55",
			wantErr: true,
		},
		{
			name: "multiple valid OUIs",
			rules: MACRules{
				OUIWhitelist: []string{"001122", "AABBCC"},
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},

		// Blacklist validation tests
		{
			name: "blacklisted MAC",
			rules: MACRules{
				Blacklist: []string{"001122"},
			},
			value:   "00:11:22:33:44:55",
			wantErr: true,
		},
		{
			name: "not blacklisted MAC",
			rules: MACRules{
				Blacklist: []string{"AABBCC"},
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},

		// Special addresses tests
		{
			name: "zero MAC not allowed",
			rules: MACRules{
				AllowZero: ptr(false),
			},
			value:   "00:00:00:00:00:00",
			wantErr: true,
		},
		{
			name: "zero MAC allowed",
			rules: MACRules{
				AllowZero: ptr(true),
			},
			value:   "00:00:00:00:00:00",
			wantErr: false,
		},
		{
			name: "broadcast MAC not allowed",
			rules: MACRules{
				AllowBroadcast: ptr(false),
			},
			value:   "FF:FF:FF:FF:FF:FF",
			wantErr: true,
		},
		{
			name: "broadcast MAC allowed",
			rules: MACRules{
				AllowBroadcast: ptr(true),
			},
			value:   "FF:FF:FF:FF:FF:FF",
			wantErr: false,
		},
		{
			name: "multicast MAC not allowed",
			rules: MACRules{
				AllowMulticast: ptr(false),
			},
			value:   "01:00:5E:00:00:00",
			wantErr: true,
		},
		{
			name: "multicast MAC allowed",
			rules: MACRules{
				AllowMulticast: ptr(true),
			},
			value:   "01:00:5E:00:00:00",
			wantErr: false,
		},

		// MaxOctets validation tests
		{
			name: "valid 6 octets",
			rules: MACRules{
				MaxOctets: ptr(6),
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "invalid 8 octets when max is 6",
			rules: MACRules{
				MaxOctets: ptr(6),
			},
			value:   "00:11:22:33:44:55:66:77",
			wantErr: true,
		},

		// net.HardwareAddr tests
		{
			name:    "valid hardware address",
			rules:   MACRules{},
			value:   net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			wantErr: false,
		},
		{
			name:    "nil hardware address when not required",
			rules:   MACRules{},
			value:   net.HardwareAddr(nil),
			wantErr: false,
		},
		{
			name:    "nil hardware address when required",
			rules:   MACRules{BaseRules: BaseRules{Required: true}},
			value:   net.HardwareAddr(nil),
			wantErr: true,
		},

		// Combined rules tests
		{
			name: "multiple valid rules",
			rules: MACRules{
				Formats:      []MACFormat{MACFormatColon},
				Cases:        []MACCase{MACCaseUpper},
				Types:        []MACAddressType{MACTypeUnicast, MACTypeUniversal},
				OUIWhitelist: []string{"001122"},
				MaxOctets:    ptr(6),
			},
			value:   "00:11:22:33:44:55",
			wantErr: false,
		},
		{
			name: "multiple rules with one failure",
			rules: MACRules{
				Formats:      []MACFormat{MACFormatColon},
				Cases:        []MACCase{MACCaseUpper},
				Types:        []MACAddressType{MACTypeMulticast}, // This should fail for a unicast address
				OUIWhitelist: []string{"001122"},
				MaxOctets:    ptr(6),
			},
			value:   "00:11:22:33:44:55",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rules.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for %s with value %v", tt.name, tt.value)
			} else {
				assert.NoError(t, err, "Unexpected error for %s with value %v: %v", tt.name, tt.value, err)
			}
		})
	}
}
