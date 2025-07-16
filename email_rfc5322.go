package sval

import (
	"strings"
	"unicode"
)

const (
	maxEmailLength  = 254
	maxLocalLength  = 64
	maxDomainLength = 255
	maxLabelLength  = 63
)

// validateEmailRFC5322 checks email with RFC 5322 standard
func validateEmailRFC5322(email string) bool {
	if len(email) > maxEmailLength || len(email) == 0 {
		return false
	}

	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	local := email[:atIndex]
	domain := email[atIndex+1:]

	return validateLocal(local) && validateDomain(domain)
}

func validateLocal(local string) bool {
	if len(local) > maxLocalLength || len(local) == 0 {
		return false
	}

	if strings.HasPrefix(local, "\"") {
		if !strings.HasSuffix(local, "\"") {
			return false // Unmatched quotes
		}
		return validateQuotedLocal(local)
	}

	if strings.Contains(local, "\"") {
		return false
	}

	return validateUnquotedLocal(local)
}

func validateUnquotedLocal(local string) bool {
	if local == "" {
		return false
	}

	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return false
	}

	if strings.Contains(local, "..") {
		return false
	}

	for _, c := range local {
		if !isAllowedLocalChar(c) {
			return false
		}
	}

	return true
}

func validateQuotedLocal(local string) bool {
	// Remove surrounding quotes
	content := local[1 : len(local)-1]

	for i := 0; i < len(content); i++ {
		c := content[i]

		if c == '\\' {
			i++
			if i >= len(content) {
				return false // Backslash at end
			}
			if content[i] != '"' && content[i] != '\\' {
				return false
			}
			continue
		}

		if c == '"' {
			return false
		}

		if c < 32 || c > 126 {
			return false
		}
	}

	return true
}

const localAllowedChars = "!#$%&'*+-/=?^_`{|}~."

func isAllowedLocalChar(c rune) bool {
	return unicode.IsLetter(c) ||
		unicode.IsDigit(c) ||
		strings.ContainsRune(localAllowedChars, c)
}

func validateDomain(domain string) bool {
	if len(domain) > maxDomainLength || len(domain) == 0 {
		return false
	}

	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}

	for _, label := range labels {
		if !validateLabel(label) {
			return false
		}
	}

	return true
}

func validateLabel(label string) bool {
	if len(label) == 0 || len(label) > maxLabelLength {
		return false
	}

	if label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}

	for _, c := range label {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
			return false
		}
	}

	return true
}
