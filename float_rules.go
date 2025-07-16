package sval

type FloatRuleName = string

const (
	FloatRuleNameMin FloatRuleName = "min"
	FloatRuleNameMax FloatRuleName = "max"
)

type FloatRules struct {
	BaseRules
	Min *float64 `json:"min" yaml:"min"`
	Max *float64 `json:"max" yaml:"max"`
}

func (r *FloatRules) Validate(i any) error {
	err := NewValidationError()

	if i == nil {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
		}
		return err
	}

	if ptr, ok := i.(*float64); ok {
		if ptr == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
			}
			return err
		}
	}

	val, ok := i.(float64)
	if !ok {
		err.AddError(BaseRuleNameType, TypeFloat, i, "value must be a float")
		return err
	}

	if r.Min != nil && val < *r.Min {
		err.AddError(FloatRuleNameMin, *r.Min, i, "value must be greater than or equal to min")
	}

	if r.Max != nil && val > *r.Max {
		err.AddError(FloatRuleNameMax, *r.Max, i, "value must be less than or equal to max")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}
