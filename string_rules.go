package sval

import "regexp"

type StringRuleName = string

const (
	StringRuleNameMinLen   StringRuleName = "min_len"
	StringRuleNameMaxLen   StringRuleName = "max_len"
	StringRuleNameRegex    StringRuleName = "regex"
	StringRuleNameAlphanum StringRuleName = "alphanum"
)

type StringRules struct {
	BaseRules
	MinLen   int    `json:"min_len" yaml:"min_len"`
	MaxLen   int    `json:"max_len" yaml:"max_len"`
	Regex    string `json:"regex" yaml:"regex"`
	Alphanum bool   `json:"alphanum" yaml:"alphanum"`
}

func (r *StringRules) Validate(i any) error {
	err := NewValidationError()

	if i == nil && r.Required {
		err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
		return err
	}

	if ptr, ok := i.(*string); ok {
		if ptr == nil && r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
			return err
		}
		i = *ptr
	}

	val, ok := i.(string)
	if !ok {
		err.AddError(BaseRuleNameType, TypeString, i, "value must be a string")
		return err
	}

	if r.Required && val == "" {
		err.AddError(BaseRuleNameRequired, r.Required, i, "field is required")
		return err
	}

	if r.MinLen > 0 && len(val) < r.MinLen {
		err.AddError(StringRuleNameMinLen, r.MinLen, i, "value too short")
	}

	if r.MaxLen > 0 && len(val) > r.MaxLen {
		err.AddError(StringRuleNameMaxLen, r.MaxLen, i, "value too long")
	}

	if r.Regex != "" {
		// TODO: move regex compilation from validate
		re, compileErr := regexp.Compile(r.Regex)
		if compileErr == nil && !re.MatchString(val) {
			err.AddError(StringRuleNameRegex, r.Regex, i, "value does not match pattern")
		}
	}

	if r.Alphanum {
		// TODO: add aplhanum checks
		// alphanum checks
	}

	if err.HasErrors() {
		return err
	}

	return err
}
