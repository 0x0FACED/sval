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
	EmailRuleNameRegexp          EmailRuleName = "regex"
)

var (
	// TODO: remove global regex, use compiled regex in rules
	emailRegexp *regexp.Regexp
)

type EmailRules struct {
	BaseRules
	Strategy        string   `json:"strategy" yaml:"strategy"`
	MinDomainLen    int      `json:"min_domain_len" yaml:"min_domain_len"`
	ExcludedDomains []string `json:"excluded_domains" yaml:"excluded_domains"`
	AllowedDomains  []string `json:"allowed_domains" yaml:"allowed_domains"`
	Regex           *string  `json:"regex,omitempty" yaml:"regex,omitempty"`
	// TODO: add compiled regex for performance
}

func (r *EmailRules) Validate(i any) error {
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
		err.AddError(BaseRuleNameType, TypeEmail, i, "value must be a string")
		return err
	}

	if val == "" {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
			return err
		}
		return nil
	}

	if r.Strategy != "" {
		if !validateEmail(val, EmailValidationStrategy(r.Strategy)) {
			err.AddError(EmailRuleNameStrategy, r.Strategy, i, "email does not conform to chosen strategy")
		}
	}

	atIndex := strings.LastIndex(val, "@")
	if atIndex == -1 {
		return err
	}

	domain := val[atIndex+1:]
	if r.MinDomainLen > 0 && len(domain) < r.MinDomainLen {
		err.AddError(EmailRuleNameMinDomainLen, r.MinDomainLen, i, "domain part of email is too short")
	}

	if len(r.ExcludedDomains) > 0 {
		for _, excluded := range r.ExcludedDomains {
			if domain == excluded {
				err.AddError(EmailRuleNameExcludedDomains, r.ExcludedDomains, i, "email domain is excluded")
			}
		}
	}

	if len(r.AllowedDomains) > 0 {
		if !slices.Contains(r.AllowedDomains, domain) {
			err.AddError(EmailRuleNameAllowedDomains, r.AllowedDomains, i, "email domain is not allowed")
		}
	}

	if r.Regex != nil {
		// TODO: compilation will be removed to avoid performance issues
		re, compileErr := regexp.Compile(*r.Regex)
		if compileErr == nil && !re.MatchString(val) {
			err.AddError(EmailRuleNameRegexp, *r.Regex, i, "email does not match the regex pattern")
		}
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

func matchRegex(value string) bool {
	return emailRegexp.MatchString(value)
}
