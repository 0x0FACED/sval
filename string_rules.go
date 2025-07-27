package sval

import (
	"math"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type StringRuleName = string

const (
	StringRuleNameMinLen       StringRuleName = "min_len"
	StringRuleNameMaxLen       StringRuleName = "max_len"
	StringRuleNameRegex        StringRuleName = "regex"
	StringRuleNameOnlyDigits   StringRuleName = "only_digits"
	StringRuleNameOnlyLetters  StringRuleName = "only_letters"
	StringRuleNameNoWhitespace StringRuleName = "no_whitespace"
	StringRuleNameTrimSpace    StringRuleName = "trim_space"
	StringRuleNameStartsWith   StringRuleName = "starts_with"
	StringRuleNameEndsWith     StringRuleName = "ends_with"
	StringRuleNameContains     StringRuleName = "contains"
	StringRuleNameNotContains  StringRuleName = "not_contains"
	StringRuleNameOneOf        StringRuleName = "one_of"
	StringRuleNameMinEntropy   StringRuleName = "min_entropy"
)

type StringRules struct {
	BaseRules
	// MinLen and MaxLen in chars, not bytes
	MinLen       int      `json:"min_len" yaml:"min_len"`
	MaxLen       int      `json:"max_len" yaml:"max_len"`
	Regex        *string  `json:"regex,omitempty" yaml:"regex,omitempty"`
	OnlyDigits   bool     `json:"only_digits" yaml:"only_digits"`
	OnlyLetters  bool     `json:"only_letters" yaml:"only_letters"`
	NoWhitespace bool     `json:"no_whitespace" yaml:"no_whitespace"`
	TrimSpace    bool     `json:"trim_space" yaml:"trim_space"`
	StartsWith   *string  `json:"starts_with,omitempty" yaml:"starts_with,omitempty"`
	EndsWith     *string  `json:"ends_with,omitempty" yaml:"ends_with,omitempty"`
	Contains     []string `json:"contains,omitempty" yaml:"contains,omitempty"`
	NotContains  []string `json:"not_contains,omitempty" yaml:"not_contains,omitempty"`
	OneOf        []string `json:"one_of,omitempty" yaml:"one_of,omitempty"`
	MinEntropy   float64  `json:"min_entropy,omitempty" yaml:"min_entropy,omitempty"`
	// TODO: add compiled regex for performance
}

var (
	onlyDigitsRegex   = regexp.MustCompile(`^\d+$`)
	onlyLettersRegex  = regexp.MustCompile(`^[a-zA-Z]+$`)
	noWhitespaceRegex = regexp.MustCompile(`^\S+$`)
)

func (r *StringRules) Validate(i any) error {
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
		err.AddError(BaseRuleNameType, TypeString, i, "value must be a string")
		return err
	}

	if val == "" {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, i, FieldIsRequired)
			return err
		}
		return nil
	}

	length := utf8.RuneCountInString(val)
	if r.MinLen > 0 && length < r.MinLen {
		err.AddError(StringRuleNameMinLen, r.MinLen, i, "string too short")
	}

	if r.MaxLen > 0 && length > r.MaxLen {
		err.AddError(StringRuleNameMaxLen, r.MaxLen, i, "string too long")
	}

	if r.Regex != nil {
		// TODO: move regex compilation from validate
		re, compileErr := regexp.Compile(*r.Regex)
		if compileErr == nil && !re.MatchString(val) {
			err.AddError(StringRuleNameRegex, *r.Regex, i, "string does not match pattern")
		}
	}

	if r.OnlyDigits && !onlyDigitsRegex.MatchString(val) {
		err.AddError(StringRuleNameOnlyDigits, true, i, "string must contain only digits")
	}

	// Only Digits and Only Letters will be checked by CLI,
	// so if both are true, it will be an error
	if r.OnlyLetters && !onlyLettersRegex.MatchString(val) {
		err.AddError(StringRuleNameOnlyLetters, true, i, "string must contain only letters")
	}

	if r.NoWhitespace && !noWhitespaceRegex.MatchString(val) {
		err.AddError(StringRuleNameNoWhitespace, true, i, "string must not contain whitespace")
	}

	// strange rule, i think must be 1st, but its here xd
	if r.TrimSpace {
		val = strings.TrimSpace(val)
	}

	if r.StartsWith != nil && !strings.HasPrefix(val, *r.StartsWith) {
		err.AddError(StringRuleNameStartsWith, *r.StartsWith, i, "string must start with specified prefix")
	}

	if r.EndsWith != nil && !strings.HasSuffix(val, *r.EndsWith) {
		err.AddError(StringRuleNameEndsWith, *r.EndsWith, i, "string must end with specified suffix")
	}

	if len(r.Contains) > 0 {
		for _, substr := range r.Contains {
			if !strings.Contains(val, substr) {
				// rule value - substr or full slice?
				err.AddError(StringRuleNameContains, substr, i, "string must contain specified substrings")
				break
			} else {
				break // if one of them is found, we can break
			}
		}
	}

	if len(r.NotContains) > 0 {
		for _, substr := range r.NotContains {
			if strings.Contains(val, substr) {
				// rule value - substr or full slice?
				err.AddError(StringRuleNameNotContains, substr, i, "string must not contain specified substring")
				break
			}
		}
	}

	if len(r.OneOf) > 0 {
		if !slices.Contains(r.OneOf, val) {
			err.AddError(StringRuleNameOneOf, r.OneOf, i, "string must be one of the specified values")
		}
	}

	// i dont know is this rule needed
	if r.MinEntropy > 0 {
		entropy := entropy(val)
		if entropy < r.MinEntropy {
			err.AddError(StringRuleNameMinEntropy, r.MinEntropy, i, "string entropy is too low")
		}
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

func entropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	frequency := make(map[rune]int)
	for _, char := range s {
		frequency[char]++
	}

	var entropy float64
	length := float64(utf8.RuneCountInString(s))
	for _, count := range frequency {
		probability := float64(count) / length
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}

	return entropy
}
