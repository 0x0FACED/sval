package sval

import (
	"slices"
	"unicode"
	"unicode/utf8"
)

type PasswordRuleName = string

const (
	PasswordRuleNameMinLen               PasswordRuleName = "min_len"                // Min len of password (in symbols, not bytes)
	PasswordRuleNameMaxLen               PasswordRuleName = "max_len"                // Max len of password (in symbols, not bytes)
	PasswordRuleNameMinUpper             PasswordRuleName = "min_upper"              // Min count of upper characters
	PasswordRuleNameMinLower             PasswordRuleName = "min_lower"              // Min count of lower characters
	PasswordRuleNameMinNumbers           PasswordRuleName = "min_numbers"            // Min count of numbers
	PasswordRuleNameMinSpecial           PasswordRuleName = "min_special"            // Min count of special characters
	PasswordRuleNameSpecialChars         PasswordRuleName = "special_chars"          // List of special characters, that must be present in the password
	PasswordRuleNameAllowedChars         PasswordRuleName = "allowed_chars"          // List of allowed characters
	PasswordRuleNameDisallowedChars      PasswordRuleName = "disallowed_chars"       // List of disallowed characters
	PasswordRuleNameMaxRepeatRun         PasswordRuleName = "max_repeat_run"         // Max consecutive repeating characters
	PasswordRuleNameDetectLinearPatterns PasswordRuleName = "detect_linear_patterns" // Active detection of linear patterns (e.g., asdfgh, 12345678)
	PasswordRuleNameBlacklist            PasswordRuleName = "blacklist"              // Blacklist of passwords
	PasswordRuleNameMinEntropy           PasswordRuleName = "min_entropy"            // Min password entropy
)

var patterns = []string{
	"abcdefghijklmnopqrstuvwxyz",
	"qwertyuiopasdfghjklzxcvbnm",
	"1234567890",
	"0987654321",
}

type PasswordRules struct {
	BaseRules
	MinLen               int      `json:"min_len" yaml:"min_len"`                               // 100% need
	MaxLen               int      `json:"max_len" yaml:"max_len"`                               // 100% need
	MinUpper             int      `json:"min_upper" yaml:"min_upper"`                           // 100% need if 0 = ignore
	MinLower             int      `json:"min_lower" yaml:"min_lower"`                           // 100% need if 0 = ignore
	MinNumbers           int      `json:"min_numbers" yaml:"min_numbers"`                       // 100% need if 0 = ignore
	MinSpecial           int      `json:"min_special" yaml:"min_special"`                       // 100% need if 0 = ignore
	SpecialChars         []rune   `json:"special_chars" yaml:"special_chars"`                   // if not empry - password must contain at least one of these
	AllowedChars         []rune   `json:"allowed_chars" yaml:"allowed_chars"`                   // if {'a', 'b', 'c'} - password must contain only these chars
	DisallowedChars      []rune   `json:"disallowed_chars" yaml:"disallowed_chars"`             // if {'a', 'b', 'c'} - password must not contain these chars
	MaxRepeatRun         int      `json:"max_repeat_run" yaml:"max_repeat_run"`                 // aaaaa, bbbbbbb, 11111 etc
	DetectLinearPatterns bool     `json:"detect_linear_patterns" yaml:"detect_linear_patterns"` // asdfgh, 12345678, qwerty etc
	Blacklist            []string `json:"blacklist" yaml:"blacklist"`                           // idunno
	MinEntropy           float64  `json:"min_entropy" yaml:"min_entropy"`                       // if 0 = ignore
}

func (r *PasswordRules) Validate(i any) error {
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
		err.AddError(StringRuleNameMinLen, r.MinLen, i, "string too short")
	}

	var (
		hasUpper   int
		hasLower   int
		hasNumber  int
		hasSpecial int
	)

	for _, char := range val {
		switch {
		case unicode.IsUpper(char):
			hasUpper++
		case unicode.IsLower(char):
			hasLower++
		case unicode.IsNumber(char):
			hasNumber++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			if len(r.SpecialChars) > 0 {
				found := false
				for _, special := range r.SpecialChars {
					if char == special {
						found = true
						hasSpecial++
						break
					}
				}
				if !found {
					err.AddError(PasswordRuleNameSpecialChars, r.SpecialChars, char, "password must contain at least one of the allowed special characters")
				}
			} else {
				hasSpecial++
			}
		}

		if len(r.DisallowedChars) > 0 {
			if slices.Contains(r.DisallowedChars, char) {
				err.AddError(PasswordRuleNameDisallowedChars, r.DisallowedChars, char, "password must not contain disallowed characters")
			}
		}

		if len(r.AllowedChars) > 0 {
			if !slices.Contains(r.AllowedChars, char) {
				err.AddError(PasswordRuleNameAllowedChars, r.AllowedChars, char, "password must contain allowed characters only")
			}
		}

	}

	if r.MinUpper > 0 && hasUpper < r.MinUpper {
		err.AddError(PasswordRuleNameMinUpper, r.MinUpper, i, "password must contain uppercase characters")
	}

	if r.MinLower > 0 && hasLower < r.MinLower {
		err.AddError(PasswordRuleNameMinLower, r.MinLower, i, "password must contain lowwercase characters")
	}

	if r.MinNumbers > 0 && hasNumber < r.MinNumbers {
		err.AddError(PasswordRuleNameMinNumbers, r.MinNumbers, i, "password must contain numbers")
	}

	if r.MinSpecial > 0 && hasSpecial < r.MinSpecial {
		err.AddError(PasswordRuleNameMinSpecial, r.MinSpecial, i, "password must contain special characters")
	}

	if r.MaxRepeatRun > 0 {
		var lastChar rune
		count := 1
		for _, char := range val {
			if char == lastChar {
				count++
				if count > r.MaxRepeatRun {
					err.AddError(PasswordRuleNameMaxRepeatRun, r.MaxRepeatRun, i, "too many consecutive identical characters")
					return err
				}
			} else {
				lastChar = char
				count = 1
			}
		}
	}

	if r.DetectLinearPatterns {
		// TODO: implement linear patterns detection
	}

	if len(r.Blacklist) > 0 {
		if slices.Contains(r.Blacklist, val) {
			err.AddError(PasswordRuleNameBlacklist, r.Blacklist, i, "password is in the blacklist")
			return err
		}
	}

	if r.MinEntropy > 0 {
		entropy := entropy(val)
		if entropy < r.MinEntropy {
			err.AddError(PasswordRuleNameMinEntropy, r.MinEntropy, i, "password entropy is too low")
			return err
		}
	}

	if err.HasErrors() {
		return err
	}

	return nil
}
