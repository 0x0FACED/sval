package sval

type IntRuleName = string

const (
	IntRuleNameMin IntRuleName = "min"
	IntRuleNameMax IntRuleName = "max"
)

type IntRules struct {
	BaseRules
	Min *int `json:"min" yaml:"min"`
	Max *int `json:"max" yaml:"max"`
}

func (r *IntRules) Validate(i any) error {
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
	}

	val, ok := i.(int)
	if !ok {
		err.AddError(BaseRuleNameType, TypeInt, "value must be int")
		return err
	}

	if r.Min != nil && val < *r.Min {
		err.AddError(IntRuleNameMin, *r.Min, "value must be greater than or equal to min")
	}

	if r.Max != nil && val > *r.Max {
		err.AddError(IntRuleNameMax, *r.Max, "value must be less than or equal to max")
	}

	if err.HasErrors() {
		return err
	}

	return nil
}
