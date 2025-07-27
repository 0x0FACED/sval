package sval

import (
	"fmt"
	"unicode/utf8"
)

// ConvertToRuneArray converts various array types to []rune.
// Supported input types:
// - []any where elements can be rune, string (single character), or int
// - []rune
// - []string where each string is a single character
// - []int where each int is a valid rune value
func ConvertToRuneArray(value any) ([]rune, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []rune:
		return v, nil
	case []string:
		return convertStringArrayToRunes(v)
	case []any:
		return convertAnyArrayToRunes(v)
	case []int:
		return convertIntArrayToRunes(v)
	case string:
		return convertStringToRunes(v)
	default:
		return nil, fmt.Errorf("unsupported type for rune array conversion: %T", value)
	}
}

// ConvertToStringArray converts various array types to []string.
// Supported input types:
// - []any where elements can be string or types that implement fmt.Stringer
// - []string
// - []rune where each rune is converted to a string
func ConvertToStringArray(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []rune:
		result := make([]string, len(v))
		for i, r := range v {
			result[i] = string(r)
		}
		return result, nil
	case []any:
		return convertAnyArrayToStrings(v)
	default:
		return nil, fmt.Errorf("unsupported type for string array conversion: %T", value)
	}
}

func convertStringArrayToRunes(arr []string) ([]rune, error) {
	result := make([]rune, 0, len(arr))
	for i, s := range arr {
		if utf8.RuneCountInString(s) != 1 {
			return nil, fmt.Errorf("string at index %d must contain exactly one character: %q", i, s)
		}
		result = append(result, []rune(s)[0])
	}
	return result, nil
}

func convertStringToRunes(str string) ([]rune, error) {
	length := utf8.RuneCountInString(str)
	result := make([]rune, 0, length)
	for _, c := range str {
		result = append(result, c)
	}

	return result, nil
}

func convertAnyArrayToRunes(arr []any) ([]rune, error) {
	result := make([]rune, 0, len(arr))
	for i, v := range arr {
		switch val := v.(type) {
		case rune:
			result = append(result, val)
		case string:
			if utf8.RuneCountInString(val) != 1 {
				return nil, fmt.Errorf("string at index %d must contain exactly one character: %q", i, val)
			}
			result = append(result, []rune(val)[0])
		case int:
			if !utf8.ValidRune(rune(val)) {
				return nil, fmt.Errorf("invalid rune value at index %d: %d", i, val)
			}
			result = append(result, rune(val))
		default:
			return nil, fmt.Errorf("unsupported type at index %d: %T", i, v)
		}
	}
	return result, nil
}

func convertIntArrayToRunes(arr []int) ([]rune, error) {
	result := make([]rune, 0, len(arr))
	for i, v := range arr {
		if !utf8.ValidRune(rune(v)) {
			return nil, fmt.Errorf("invalid rune value at index %d: %d", i, v)
		}
		result = append(result, rune(v))
	}
	return result, nil
}

func convertAnyArrayToStrings(arr []any) ([]string, error) {
	result := make([]string, 0, len(arr))
	for i, v := range arr {
		switch val := v.(type) {
		case string:
			result = append(result, val)
		case fmt.Stringer:
			result = append(result, val.String())
		default:
			return nil, fmt.Errorf("unsupported type at index %d: %T", i, v)
		}
	}
	return result, nil
}
