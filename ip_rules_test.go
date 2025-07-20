package sval

import (
	"net"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPRules(t *testing.T) {
	tests := []struct {
		name    string
		rules   IPRules
		value   any
		wantErr bool
	}{
		// Basic validation tests
		{
			name:    "empty string when not required",
			rules:   IPRules{BaseRules: BaseRules{Required: false}},
			value:   "",
			wantErr: false,
		},
		{
			name:    "ip in net.IP format",
			rules:   IPRules{BaseRules: BaseRules{Required: false}, AllowPrivate: true},
			value:   net.IPv4(192, 168, 0, 1),
			wantErr: false,
		},
		{
			name:    "ip in netip.Addr format",
			rules:   IPRules{BaseRules: BaseRules{Required: false}, AllowPrivate: true},
			value:   netip.AddrFrom4([4]byte{192, 168, 0, 1}),
			wantErr: false,
		},
		{
			name:  "ip in *net.IP format",
			rules: IPRules{BaseRules: BaseRules{Required: false}, AllowPrivate: true},
			value: func() *net.IP {
				ip := net.IPv4(192, 168, 0, 1)
				return &ip
			},
			wantErr: false,
		},
		{
			name:  "ip in *netip.Addr format",
			rules: IPRules{BaseRules: BaseRules{Required: false}, AllowPrivate: true},
			value: func() *netip.Addr {
				ip := netip.AddrFrom4([4]byte{192, 168, 0, 1})
				return &ip
			},
			wantErr: false,
		},
		{
			name:    "empty string when not required",
			rules:   IPRules{BaseRules: BaseRules{Required: false}},
			value:   "",
			wantErr: false,
		},
		{
			name:    "empty string when required",
			rules:   IPRules{BaseRules: BaseRules{Required: true}},
			value:   "",
			wantErr: true,
		},
		{
			name:    "empty string ptr when required",
			rules:   IPRules{BaseRules: BaseRules{Required: true}},
			value:   nil,
			wantErr: true,
		},
		{
			name:    "empty string ptr when not required",
			rules:   IPRules{BaseRules: BaseRules{Required: false}},
			value:   nil,
			wantErr: false,
		},
		{
			name:    "non-string value",
			rules:   IPRules{},
			value:   123,
			wantErr: true,
		},

		// IPv4 validation tests
		{
			name:    "valid IPv4",
			rules:   IPRules{Version: 4, AllowPrivate: true},
			value:   "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "invalid IPv4 format",
			rules:   IPRules{Version: 4},
			value:   "256.1.2.3",
			wantErr: true,
		},
		{
			name:    "IPv4 with wrong segments count",
			rules:   IPRules{Version: 4},
			value:   "192.168.1",
			wantErr: true,
		},
		{
			name:    "IPv4 with letters",
			rules:   IPRules{Version: 4},
			value:   "192.168.1.abc",
			wantErr: true,
		},

		// IPv6 validation tests
		{
			name:    "valid IPv6",
			rules:   IPRules{Version: 6},
			value:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantErr: false,
		},
		{
			name:    "valid IPv6 compressed",
			rules:   IPRules{Version: 6},
			value:   "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "invalid IPv6 format",
			rules:   IPRules{Version: 6},
			value:   "2001:db8::::::1",
			wantErr: true,
		},
		{
			name:    "IPv6 with invalid characters",
			rules:   IPRules{Version: 6},
			value:   "2001:db8::xyz",
			wantErr: true,
		},

		// Mixed version tests
		{
			name:    "valid IPv4 when both versions allowed",
			rules:   IPRules{Version: 0, AllowPrivate: true},
			value:   "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "Valid IPv6 when both versions allowed",
			rules:   IPRules{Version: 0},
			value:   "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "IPv6 when only IPv4 allowed",
			rules:   IPRules{Version: 4},
			value:   "2001:db8::1",
			wantErr: true,
		},
		{
			name:    "IPv4 when only IPv6 allowed",
			rules:   IPRules{Version: 6},
			value:   "192.168.1.1",
			wantErr: true,
		},

		// Special cases
		{
			name:    "IPv4 loopback",
			rules:   IPRules{Version: 4},
			value:   "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "IPv4 zeros",
			rules:   IPRules{Version: 4},
			value:   "0.0.0.0",
			wantErr: false,
		},
		{
			name:    "IPv4 broadcast",
			rules:   IPRules{Version: 4},
			value:   "255.255.255.255",
			wantErr: false,
		},
		{
			name:    "IPv6 loopback",
			rules:   IPRules{Version: 6},
			value:   "::1",
			wantErr: false,
		},
		{
			name:    "IPv6 zeros",
			rules:   IPRules{Version: 6},
			value:   "::",
			wantErr: false,
		},

		// allowed and not allowed private IPs
		{
			name:    "not allowed private IPv4",
			rules:   IPRules{Version: 4, AllowPrivate: false},
			value:   "192.168.0.1",
			wantErr: true,
		},
		{
			name:    "not allowed private IPv6",
			rules:   IPRules{Version: 6, AllowPrivate: false},
			value:   "fe80::1",
			wantErr: true,
		},
		{
			name:    "allowed private IPv4",
			rules:   IPRules{Version: 4, AllowPrivate: true},
			value:   "0.0.0.0",
			wantErr: false,
		},
		{
			name:    "allowed private IPv6",
			rules:   IPRules{Version: 6, AllowPrivate: true},
			value:   "fe80::1",
			wantErr: false,
		},

		// allowed subnets tests
		{
			name: "IPv4 in allowed subnet",
			rules: IPRules{
				Version:        4,
				AllowedSubnets: []string{"192.168.0.0/16"},
				AllowPrivate:   true,
			},
			value:   "192.168.1.1",
			wantErr: false,
		},
		{
			name: "IPv4 not in allowed subnet",
			rules: IPRules{
				Version:        4,
				AllowedSubnets: []string{"192.168.0.0/16"},
			},
			value:   "172.16.1.1",
			wantErr: true,
		},
		{
			name: "IPv4 in one of multiple allowed subnets",
			rules: IPRules{
				Version:        4,
				AllowedSubnets: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
				AllowPrivate:   true,
			},
			value:   "172.16.1.1",
			wantErr: false,
		},
		{
			name: "IPv6 in allowed subnet",
			rules: IPRules{
				Version:        6,
				AllowedSubnets: []string{"2001:db8::/32"},
				AllowPrivate:   true,
			},
			value:   "2001:db8::1",
			wantErr: false,
		},
		{
			name: "IPv6 not in allowed subnet",
			rules: IPRules{
				Version:        6,
				AllowedSubnets: []string{"2001:db8::/32"},
			},
			value:   "2002:db8::1",
			wantErr: true,
		},
		{
			name: "Invalid CIDR in allowed subnets",
			rules: IPRules{
				Version:        4,
				AllowedSubnets: []string{"invalid-cidr"},
			},
			value:   "192.168.1.1",
			wantErr: true,
		},

		// excluded subnets tests
		{
			name: "IPv4 in excluded subnet",
			rules: IPRules{
				Version:         4,
				ExcludedSubnets: []string{"192.168.0.0/16"},
			},
			value:   "192.168.1.1",
			wantErr: true,
		},
		{
			name: "IPv4 not in excluded subnet",
			rules: IPRules{
				Version:         4,
				ExcludedSubnets: []string{"192.168.0.0/16"},
				AllowPrivate:    true,
			},
			value:   "172.16.1.1",
			wantErr: false,
		},
		{
			name: "IPv4 in one of multiple excluded subnets",
			rules: IPRules{
				Version:         4,
				ExcludedSubnets: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
			},
			value:   "172.16.1.1",
			wantErr: true,
		},
		{
			name: "IPv6 in excluded subnet",
			rules: IPRules{
				Version:         6,
				ExcludedSubnets: []string{"2001:db8::/32"},
			},
			value:   "2001:db8::1",
			wantErr: true,
		},
		{
			name: "IPv6 not in excluded subnet",
			rules: IPRules{
				Version:         6,
				ExcludedSubnets: []string{"2001:db8::/32"},
			},
			value:   "2002:db8::1",
			wantErr: false,
		},
		{
			name: "Invalid CIDR in excluded subnets",
			rules: IPRules{
				Version:         4,
				ExcludedSubnets: []string{"invalid-cidr"},
			},
			value:   "192.168.1.1",
			wantErr: true,
		},
		{
			name: "Combined allowed and excluded subnets - allowed wins",
			rules: IPRules{
				Version:         4,
				AllowedSubnets:  []string{"192.168.0.0/16"},
				ExcludedSubnets: []string{"192.168.0.0/24"},
				AllowPrivate:    true,
			},
			value:   "192.168.1.1",
			wantErr: false,
		},
		{
			name: "Combined allowed and excluded subnets - excluded wins",
			rules: IPRules{
				Version:         4,
				AllowedSubnets:  []string{"192.168.0.0/16"},
				ExcludedSubnets: []string{"192.168.0.0/24"},
			},
			value:   "192.168.0.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ok := tt.value.(func() *netip.Addr)
			if ok {
				tt.value = f()
			} else {
				f, ok := tt.value.(func() *net.IP)
				if ok {
					tt.value = f()
				}
			}
			err := tt.rules.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for %s with value %v", tt.name, tt.value)
			} else {
				assert.NoError(t, err, "Unexpected error for %s with value %v", tt.name, tt.value)
			}
		})
	}
}
