package sval

import "errors"

type PasswordRules struct {
	BaseRules
	MinLen         int  `json:"min_len" yaml:"min_len"`
	MaxLen         int  `json:"max_len" yaml:"max_len"`
	RequireUpper   bool `json:"require_upper" yaml:"require_upper"`
	RequireLower   bool `json:"require_lower" yaml:"require_lower"`
	RequireNumber  bool `json:"require_number" yaml:"require_number"`
	RequireSpecial bool `json:"require_special" yaml:"require_special"`
}

func (r *PasswordRules) Validate(i any) error {
	val, ok := i.(string)
	if !ok {
		return errors.New("value must be a string")
	}

	if r.Required && val == "" {
		return errors.New("password is required")
	}

	// other validations

	return nil
}
