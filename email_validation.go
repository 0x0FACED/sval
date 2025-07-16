package sval

import (
	"regexp"
	"strings"
	"unicode"
)

type EmailValidationStrategy string

const (
	// RFC5321 - smtp validation
	RFC5321 EmailValidationStrategy = "rfc5321"
	// RFC5322 - more liberal validation for message format
	RFC5322 EmailValidationStrategy = "rfc5322"
	// HTMLInput - simple validation like in HTML5 input[type=email]
	HTMLInput EmailValidationStrategy = "html"
)

func validateStrategy(strategy EmailValidationStrategy) bool {
	switch strategy {
	case RFC5321, RFC5322, HTMLInput:
		return true
	default:
		return false
	}
}

// HTML5 email regex from WHATWG
// https://html.spec.whatwg.org/multipage/input.html#valid-e-mail-address
var htmlEmailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func validateEmail(email string, strategy EmailValidationStrategy) bool {
	switch strategy {
	case RFC5321:
		return validateEmailRFC5321(email)
	case RFC5322:
		return validateEmailRFC5322(email)
	case HTMLInput:
		return validateEmailHTML(email)
	default:
		return validateEmailRFC5322(email)
	}
}

func validateEmailHTML(email string) bool {
	if len(email) > 254 {
		return false
	}
	return htmlEmailRegex.MatchString(email)
}

func validateEmailRFC5321(email string) bool {
	if len(email) > 254 || len(email) == 0 {
		return false
	}

	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	local := email[:atIndex]
	domain := email[atIndex+1:]

	return validateLocalRFC5321(local) && validateDomainRFC5321(domain)
}

func validateLocalRFC5321(local string) bool {
	if len(local) > 64 || len(local) == 0 {
		return false
	}

	if strings.Contains(local, "\"") {
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

func validateDomainRFC5321(domain string) bool {
	if len(domain) > 255 || len(domain) == 0 {
		return false
	}

	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}

	for _, label := range labels {
		if !validateSMTPLabel(label) {
			return false
		}
	}

	lastLabel := labels[len(labels)-1]
	for _, c := range lastLabel {
		if !unicode.IsLetter(c) {
			return false
		}
	}

	return true
}

func validateSMTPLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 {
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
