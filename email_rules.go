package sval

import (
	"regexp"
	"slices"
	"strings"
)

type EmailRuleName = string

const (
	EmailRuleNameStrategy        EmailRuleName = "strategy"
	EmailRuleNameMinDomainLen    EmailRuleName = "min_domain_len"
	EmailRuleNameExcludedDomains EmailRuleName = "excluded_domains"
	EmailRuleNameAllowedDomains  EmailRuleName = "allowed_domains"
	EmailRuleNameRegexp          EmailRuleName = "regexp"
)

var (
	emailRegexp *regexp.Regexp
)

type EmailRules struct {
	BaseRules
	Strategy        string   `json:"strategy" yaml:"strategy"`
	MinDomainLen    int      `json:"min_domain_len" yaml:"min_domain_len"`
	ExcludedDomains []string `json:"excluded_domains" yaml:"excluded_domains"`
	AllowedDomains  []string `json:"allowed_domains" yaml:"allowed_domains"`
	Regex           *string  `json:"regex" yaml:"regex"`
}

func (r *EmailRules) Validate(i any) error {
	err := NewValidationError()

	if i == nil && r.Required {
		err.AddError(BaseRuleNameRequired, r.Required, FieldIsRequired)
		return err
	}

	if ptr, ok := i.(*string); ok {
		if ptr == nil && r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, FieldIsRequired)
			return err
		}
		i = *ptr
	}

	val, ok := i.(string)
	if !ok {
		err.AddError(BaseRuleNameType, "string", "value must be a string")
		return err
	}

	if r.Required && val == "" {
		err.AddError(BaseRuleNameRequired, r.Required, FieldIsRequired)
		return err
	}

	if r.Strategy != "" {
		if !validateEmail(val, EmailValidationStrategy(r.Strategy)) {
			err.AddError(EmailRuleNameStrategy, r.Strategy, "email does not conform to chosen strategy")
		}
	}

	atIndex := strings.LastIndex(val, "@")
	if atIndex == -1 {
		return err
	}

	domain := val[atIndex+1:]
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

	if r.Regex != nil {
		if !matchRegex(val) {
			err.AddError(EmailRuleNameRegexp, *r.Regex, "email does not match the regex pattern")
		}
	}

	if err.HasErrors() {
		return err
	}

	return err
}

func matchRegex(value string) bool {
	return emailRegexp.MatchString(value)
}
