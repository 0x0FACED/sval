package sval

import (
	"errors"
	"fmt"
)

type NumberRules struct {
	BaseRules
	Min *float64 `json:"min" yaml:"min"`
	Max *float64 `json:"max" yaml:"max"`
}

func (r *NumberRules) Validate(i any) error {
	val, ok := i.(int)
	if !ok {
		return errors.New("value must be a string")
	}

	if r.Required {
		return errors.New("field is required")
	}

	if val < int(*r.Min) {
		return fmt.Errorf("value too short, min length: %d", r.Min)
	}

	if val > int(*r.Min) {
		return fmt.Errorf("value too long, max length: %d", r.Max)
	}

	// other validations

	return nil
}
