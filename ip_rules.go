package sval

import (
	"net"
	"net/netip"
)

type IPRuleName = string

const (
	IPRuleNameVersion         IPRuleName = "version"
	IPRuleNameAllowPrivate    IPRuleName = "allow_private"
	IPRuleNameAllowedSubnets  IPRuleName = "allowed_subnets"
	IPRuleNameExcludedSubnets IPRuleName = "excluded_subnets"
)

type IPRules struct {
	BaseRules
	Version         int      `json:"version" yaml:"version"` // 4, 6 or 0 for both
	AllowPrivate    bool     `json:"allow_private" yaml:"allow_private"`
	AllowedSubnets  []string `json:"allowed_subnets" yaml:"allowed_subnets"`
	ExcludedSubnets []string `json:"excluded_subnets" yaml:"excluded_subnets"`
}

func (r *IPRules) Validate(i any) error {
	err := NewValidationError()

	if i == nil {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
			return err
		}
		return nil
	}

	if ptr, ok := i.(*string); ok {
		if ptr == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
				return err
			}
			return nil
		}
		i = *ptr
	}

	val, ok := i.(string)
	if !ok {
		err.AddError(BaseRuleNameType, TypeIP, i, "value must be a string")
		return err
	}

	if val == "" {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
			return err
		}
		return nil
	}

	ip, errParse := netip.ParseAddr(val)
	if errParse != nil {
		err.AddError(BaseRuleNameType, TypeIP, i, "invalid IP address format")
		return err
	}

	if !r.validateVersion(ip) {
		err.AddError(IPRuleNameVersion, r.Version, i, "IP version mismatch")
		return err
	}

	if !r.AllowPrivate && (ip.IsPrivate() || ip.IsLinkLocalUnicast()) {
		err.AddError(IPRuleNameAllowPrivate, r.AllowPrivate, i, "private or link-local IPs are not allowed")
		return err
	}

	if len(r.AllowedSubnets) > 0 {
		// TODO: separate after cli will be implemented.
		// TEMP. In the future will be cli that will validate sval config files.
		// So cli will validate that all allowed subnets are valid CIDR notations.
		allowed := false
		for _, subnet := range r.AllowedSubnets {
			_, netIP, errParse := net.ParseCIDR(subnet)
			if errParse != nil {
				err.AddError(IPRuleNameAllowedSubnets, r.AllowedSubnets, i, "invalid allowed subnet format")
				return err
			}
			if netIP.Contains(ip.AsSlice()) {
				allowed = true
				break
			}
		}
		if !allowed {
			err.AddError(IPRuleNameAllowedSubnets, r.AllowedSubnets, i,
				"IP is not in any of the allowed subnets")
			return err
		}
	}

	if len(r.ExcludedSubnets) > 0 {
		// TODO: same as above.
		for _, subnet := range r.ExcludedSubnets {
			_, netIP, errParse := net.ParseCIDR(subnet)
			if errParse != nil {
				err.AddError(IPRuleNameExcludedSubnets, r.ExcludedSubnets, i, "invalid excluded subnet format")
				return err
			}
			if netIP.Contains(ip.AsSlice()) {
				err.AddError(IPRuleNameExcludedSubnets, r.ExcludedSubnets, i,
					"IP is in an excluded subnet")
				return err
			}
		}
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

func (r *IPRules) validateVersion(ip netip.Addr) bool {
	switch r.Version {
	case 4:
		return ip.Is4()
	case 6:
		return ip.Is6()
	case 0:
		return ip.Is4() || ip.Is6()
	default:
		return false
	}
}

// deadcode (mb will use it later)
func isValidIPv4(ip string) bool {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return false
	}

	if netIP.To4() == nil {
		return false
	}

	return true
}

// deadcode (mb will use it later)
func isValidIPv6(ip string) bool {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return false
	}

	if netIP.To16() == nil {
		return false
	}

	return true
}
