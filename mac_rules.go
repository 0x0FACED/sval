package sval

import (
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type MACRuleName = string

const (
	MACRuleNameFormat     MACRuleName = "formats"
	MACRuleNameMaxOctets  MACRuleName = "max_octets"
	MACRuleNameCase       MACRuleName = "cases"
	MACRuleNameType       MACRuleName = "types"
	MACRuleNameOUI        MACRuleName = "oui_whitelist"
	MACRuleNameBlacklist  MACRuleName = "blacklist"
	MACRuleNameAllowZero  MACRuleName = "allow_zero"
	MACRuleNameAllowBroad MACRuleName = "allow_broadcast"
	MACRuleNameAllowMulti MACRuleName = "allow_multicast"
)

type MACFormat = string

const (
	MACFormatAny    MACFormat = "any"    // any of formats are allowed. used by default
	MACFormatColon  MACFormat = "colon"  // 00:00:5e:00:53:01
	MACFormatHyphen MACFormat = "hyphen" // 00-00-5e-00-53-01
	MACFormatDot    MACFormat = "dot"    // 0000.5e00.5301
	MACFormatRaw    MACFormat = "raw"    // 00005e005301
)

type MACCase = string

const (
	MACCaseAny   MACCase = "any" // by default
	MACCaseLower MACCase = "lower"
	MACCaseUpper MACCase = "upper"
	MACCaseCamel MACCase = "camel" // Cisco-style (0000.5E00.5301)
)

type MACAddressType = string

const (
	MACTypeUnicast   MACAddressType = "unicast"   // bit 0 == 0
	MACTypeMulticast MACAddressType = "multicast" // bit 0 == 1
	MACTypeUniversal MACAddressType = "universal" // bit 1 == 0
	MACTypeLocal     MACAddressType = "local"     // bit 1 == 1
)

type MACRules struct {
	BaseRules
	Formats        []MACFormat      `json:"formats,omitempty" yaml:"formats"`                 // check MACFormat for available values
	Cases          []MACCase        `json:"cases,omitempty" yaml:"cases"`                     // check MACCase for available values
	Types          []MACAddressType `json:"types,omitempty" yaml:"types"`                     // check MACAddressType for available values
	AllowZero      *bool            `json:"allow_zero,omitempty" yaml:"allow_zero"`           // does 00:00:00:00:00:00 allowed
	AllowBroadcast *bool            `json:"allow_broadcast,omitempty" yaml:"allow_broadcast"` // does FF:FF:FF:FF:FF:FF allowed
	AllowMulticast *bool            `json:"allow_multicast,omitempty" yaml:"allow_multicast"` // does 01:00:... allowed
	OUIWhitelist   []string         `json:"oui_whitelist,omitempty" yaml:"oui_whitelist"`
	Blacklist      []string         `json:"blacklist,omitempty" yaml:"blacklist"`
	MaxOctets      *int             `json:"max_octets,omitempty" yaml:"max_octets"`
}

func (r *MACRules) Validate(i any) error {
	err := NewValidationError()

	if i == nil {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
			return err
		}
		return nil
	}

	switch v := i.(type) {
	case *string:
		if v == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
				return err
			}
			return nil
		}
		i = *v
	case string:
		if v == "" {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
				return err
			}
			return nil
		}
	case net.HardwareAddr:
		if v == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
				return err
			}
			return nil
		}
		i = v.String()
	case *net.HardwareAddr:
		if v == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
				return err
			}
			return nil
		}
		if *v == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
				return err
			}
			return nil
		}
		i = (*v).String()
	default:
		err.AddError(BaseRuleNameType, TypeIP, i, "value must be a string or net.HardwareAddr or ptr of them")
		return err
	}

	val, ok := i.(string)
	if !ok {
		err.AddError(BaseRuleNameType, TypeMAC, i, "value must be a string")
		return err
	}

	normalized := r.normalizeMAC(val)
	if normalized == "" {
		if len(r.Formats) > 0 {
			err.AddError(MACRuleNameFormat, r.Formats, i, "invalid MAC address format")
			return err
		}
		err.AddError(MACRuleNameFormat, MACFormatAny, i, "invalid MAC address format")
		return err
	}

	if r.MaxOctets != nil {
		octets := len(normalized) / 2
		if octets > *r.MaxOctets {
			err.AddError(MACRuleNameMaxOctets, r.MaxOctets, i, "too many octets in MAC address")
			return err
		}
	}

	if len(r.Formats) > 0 {
		if !r.validateFormat(val) {
			err.AddError(MACRuleNameFormat, r.Formats, i, "invalid MAC address format")
			return err
		}
	}

	if len(r.Cases) > 0 {
		if !r.validateCase(val) {
			err.AddError(MACRuleNameCase, r.Cases, i, "incorrect MAC address case")
			return err
		}
	}

	if len(r.Types) > 0 {
		valid := false
		for _, t := range r.Types {
			if r.validateType(normalized, t) {
				valid = true
				break
			}
		}
		if !valid {
			err.AddError(MACRuleNameType, r.Types, i, "MAC address does not match any of the required types")
			return err
		}
	}

	if len(r.OUIWhitelist) > 0 {
		valid := false
		oui := normalized[:6]
		for _, prefix := range r.OUIWhitelist {
			if strings.EqualFold(oui, prefix) {
				valid = true
				break
			}
		}
		if !valid {
			err.AddError(MACRuleNameOUI, r.OUIWhitelist, i, "MAC address OUI not in allowed list")
			return err
		}
	}

	if len(r.Blacklist) > 0 {
		for _, blocked := range r.Blacklist {
			if strings.HasPrefix(strings.ToLower(normalized), strings.ToLower(blocked)) {
				err.AddError(MACRuleNameBlacklist, r.Blacklist, i, "MAC address is blacklisted")
				return err
			}
		}
	}

	if isZeroMAC(normalized) && !(r.AllowZero != nil && *r.AllowZero) {
		err.AddError(MACRuleNameAllowZero, false, i, "zero MAC address is not allowed")
		return err
	}
	if isBroadcastMAC(normalized) {
		if !(r.AllowBroadcast != nil && *r.AllowBroadcast) {
			err.AddError(MACRuleNameAllowBroad, false, i, "broadcast MAC address is not allowed")
			return err
		}
		return nil // If broadcast is allowed, we don't need to check multicast
	}
	if isMulticastMAC(normalized) && !(r.AllowMulticast != nil && *r.AllowMulticast) {
		err.AddError(MACRuleNameAllowMulti, false, i, "multicast MAC address is not allowed")
		return err
	}

	return nil
}

// TODO: remove regexp and use strings directly or move regexp compilation to global scope
func (r *MACRules) normalizeMAC(mac string) string {
	normalized := strings.ToLower(strings.NewReplacer(":", "", "-", "", ".", "").Replace(mac))

	if !regexp.MustCompile("^[0-9a-f]+$").MatchString(normalized) {
		return ""
	}

	return normalized
}

// TODO: remove regexp and use strings directly or move regexp compilation to global scope
func (r *MACRules) validateFormat(mac string) bool {
	if slices.Contains(r.Formats, MACFormatAny) {
		return true
	}

	formatMap := make(map[MACFormat]*regexp.Regexp)
	formatMap[MACFormatColon] = regexp.MustCompile("^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$")
	formatMap[MACFormatHyphen] = regexp.MustCompile("^([0-9A-Fa-f]{2}-){5}[0-9A-Fa-f]{2}$")
	formatMap[MACFormatDot] = regexp.MustCompile("^[0-9A-Fa-f]{4}.[0-9A-Fa-f]{4}.[0-9A-Fa-f]{4}$")
	formatMap[MACFormatRaw] = regexp.MustCompile("^[0-9A-Fa-f]{12}$")

	for _, format := range r.Formats {
		// in the future this check wont be necessary, because rules will be validated
		reg, ok := formatMap[format]
		if ok && reg.MatchString(mac) {
			return true
		}
	}

	return false
}

func (r *MACRules) validateCase(mac string) bool {
	// if any - always true
	if slices.Contains(r.Cases, MACCaseAny) {
		return true
	}

	for _, _case := range r.Cases {
		if r.validateOneCase(mac, _case) {
			return true
		}
	}

	return false
}

func (r *MACRules) validateOneCase(mac string, _case MACCase) bool {
	switch _case {
	case MACCaseLower:
		return mac == strings.ToLower(mac)
	case MACCaseUpper:
		return mac == strings.ToUpper(mac)
	case MACCaseCamel:
		letters := regexp.MustCompile("[A-Fa-f]").FindAllString(mac, -1)
		for _, letter := range letters {
			if letter != strings.ToUpper(letter) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (r *MACRules) validateType(mac string, typ MACAddressType) bool {
	firstByte, err := strconv.ParseInt(mac[:2], 16, 8)
	if err != nil {
		return false
	}

	switch typ {
	case MACTypeUnicast:
		return (firstByte & 0x01) == 0
	case MACTypeMulticast:
		return (firstByte & 0x01) == 1
	case MACTypeUniversal:
		return (firstByte & 0x02) == 0
	case MACTypeLocal:
		return (firstByte & 0x02) == 2
	default:
		return false
	}
}

func isZeroMAC(mac string) bool {
	return strings.ToLower(mac) == strings.Repeat("0", len(mac))
}

func isBroadcastMAC(mac string) bool {
	return strings.ToLower(mac) == strings.Repeat("f", len(mac))
}

func isMulticastMAC(mac string) bool {
	// Don't treat broadcast as multicast
	if isBroadcastMAC(mac) {
		return false
	}
	b, _ := strconv.ParseInt(mac[:2], 16, 8)
	return b&0x01 == 1
}
