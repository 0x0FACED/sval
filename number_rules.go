package sval

type NumberRuleName = string

const (
	NumberRuleNameMin NumberRuleName = "min"
	NumberRuleNameMax NumberRuleName = "max"
)

type NumberRules struct {
	BaseRules
	Min *float64 `json:"min" yaml:"min"`
	Max *float64 `json:"max" yaml:"max"`
}

func (r *NumberRules) Validate(i any) error {
	err := NewValidationError()

	if i == nil {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, FieldIsRequired)
		}
		return err
	}

	if ptr, ok := i.(*int); ok {
		if ptr == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, FieldIsRequired)
			}
			return err
		}
		i = *ptr
	}

	val, ok := i.(int)
	if !ok {
		err.AddError(BaseRuleNameType, "number", "value must be a number")
		return err
	}

	if r.Required && val == 0 {
		err.AddError(BaseRuleNameRequired, r.Required, FieldIsRequired)
	}

	if r.Min != nil && float64(val) < *r.Min {
		err.AddError(NumberRuleNameMin, *r.Min, "value must be greater than or equal to min")
	}

	if r.Max != nil && float64(val) > *r.Max {
		err.AddError(NumberRuleNameMax, *r.Max, "value must be less than or equal to max")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}
