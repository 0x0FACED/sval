package sval

import (
	"slices"
)

type EmailRuleName = string

const (
	EmailRuleNameRFC             EmailRuleName = "rfc"
	EmailRuleNameMinDomainLen    EmailRuleName = "min_domain_len"
	EmailRuleNameExcludedDomains EmailRuleName = "excluded_domains"
	EmailRuleNameAllowedDomains  EmailRuleName = "allowed_domains"
)

type EmailRules struct {
	BaseRules
	RFC             bool     `json:"rfc" yaml:"rfc"`
	MinDomainLen    int      `json:"min_domain_len" yaml:"min_domain_len"`
	ExcludedDomains []string `json:"excluded_domains" yaml:"excluded_domains"`
	AllowedDomains  []string `json:"allowed_domains" yaml:"allowed_domains"`
}

func (r *EmailRules) Validate(i any) error {
	err := NewValidationError()

	val, ok := i.(string)
	if !ok {
		err.AddError(BaseRuleType, "string", "value must be a string")
		return err
	}

	if r.Required && val == "" {
		err.AddError(BaseRuleNameRequired, r.Required, "field is required")
		return err
	}

	if r.RFC && !emailRFC(val) {
		err.AddError(EmailRuleNameRFC, r.RFC, "email does not conform to RFC standards")
	}

	domain := extractDomain(val)
	if r.MinDomainLen > 0 && len(domain) < r.MinDomainLen {
		err.AddError(EmailRuleNameMinDomainLen, r.MinDomainLen, "domain part of email is too short")
	}

	if len(r.ExcludedDomains) > 0 {
		for _, excluded := range r.ExcludedDomains {
			if domain == excluded {
				err.AddError(EmailRuleNameExcludedDomains, r.ExcludedDomains, "email domain is excluded")
			}
		}
	}

	if len(r.AllowedDomains) > 0 {
		if !slices.Contains(r.AllowedDomains, domain) {
			err.AddError(EmailRuleNameAllowedDomains, r.AllowedDomains, "email domain is not allowed")
		}
	}

	if err.HasErrors() {
		return err
	}

	return err

}

// simple rfc validation for email
func emailRFC(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false // More than one '@' found
			}
			atIndex = i
		}
	}

	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false // '@' not found or is at the start/end
	}
	return true
}

func extractDomain(email string) string {
	atIndex := len(email) - 1
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			atIndex = i
			break
		}
	}
	return email[atIndex+1:]
}
