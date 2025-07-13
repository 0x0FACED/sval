package sval

import (
	"errors"
)

type IPRules struct {
	BaseRules
	Version int `json:"version" yaml:"version"` // 4, 6 or 0 for both
}

func (r *IPRules) Validate(i any) error {
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
