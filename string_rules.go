package sval

import (
	"errors"
	"fmt"
)

type StringRules struct {
	BaseRules
	MinLen   int    `json:"min_len" yaml:"min_len"`
	MaxLen   int    `json:"max_len" yaml:"max_len"`
	Regex    string `json:"regex" yaml:"regex"`
	Alphanum bool   `json:"alphanum" yaml:"alphanum"`
}

func (r *StringRules) Validate(i any) error {
	val, ok := i.(string)
	if !ok {
		return errors.New("value must be a string")
	}

	if r.Required && val == "" {
		return errors.New("field is required")
	}

	if r.MinLen > 0 && len(val) < r.MinLen {
		return fmt.Errorf("value too short, min length: %d", r.MinLen)
	}

	if r.MaxLen > 0 && len(val) > r.MaxLen {
		return fmt.Errorf("value too long, max length: %d", r.MaxLen)
	}

	if r.Regex != "" {
		// regexp check
	}

	if r.Alphanum {
		// alphanum checks
	}

	return nil
}
