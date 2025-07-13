package sval

import (
	"errors"
)

type EmailRules struct {
	BaseRules
	RFC             bool     `json:"rfc" yaml:"rfc"`
	MinDomainLen    int      `json:"min_domain_len" yaml:"min_domain_len"`
	ExcludedDomains []string `json:"excluded_domains" yaml:"excluded_domains"`
	AllowedDomains  []string `json:"allowed_domains" yaml:"allowed_domains"`
}

func (r *EmailRules) Validate(i any) error {
	val, ok := i.(string)
	if !ok {
		return errors.New("value must be a string")
	}

	if r.Required && val == "" {
		return errors.New("field is required")
	}

	// other validations

	return nil
}
