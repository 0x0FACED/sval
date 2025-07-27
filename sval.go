package sval

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	FieldIsRequired = "field is required"
)

var indexRegex = regexp.MustCompile(`\[\d+\]`)

type validator struct {
	rules map[string]RuleSet
}

type ValidatorConfig struct {
	Version int                   `yaml:"version" json:"version"`
	Rules   map[string]RuleConfig `yaml:"rules" json:"rules"`
}

type ConfigLoader interface {
	Load() (ValidatorConfig, error)
}

type FileConfigLoader struct {
	Path string
}

func (l *FileConfigLoader) Load() (ValidatorConfig, error) {
	data, err := os.ReadFile(l.Path)
	if err != nil {
		return ValidatorConfig{}, err
	}

	var config ValidatorConfig

	switch {
	case strings.HasSuffix(l.Path, ".yaml"), strings.HasSuffix(l.Path, ".yml"):
		err = yaml.Unmarshal(data, &config)
	case strings.HasSuffix(l.Path, ".json"):
		err = json.Unmarshal(data, &config)
	default:
		return ValidatorConfig{}, errors.New("unsupported config format")
	}

	return config, err
}

func DefaultConfigLoader() ConfigLoader {
	paths := []string{
		"sval.yaml",
		"sval.yml",
		"sval.json",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return &FileConfigLoader{Path: p}
		}
	}
	return nil
}

func NewWithConfig(loader ConfigLoader) (*validator, error) {
	config, err := loader.Load()
	if err != nil {
		return nil, err
	}
	return NewValidatorFromConfig(config)
}

func New() (*validator, error) {
	loader := DefaultConfigLoader()
	if loader == nil {
		return nil, errors.New("no config file found")
	}
	return NewWithConfig(loader)
}

type RuleSet interface {
	Validate(i any) error
}

func (v *validator) AddRule(fieldName string, rules RuleSet) {
	if v.rules == nil {
		v.rules = make(map[string]RuleSet)
	}
	v.rules[fieldName] = rules
}

type RuleType string

const (
	TypeString   RuleType = "string"
	TypeEmail    RuleType = "email"
	TypePassword RuleType = "password"
	TypeInt      RuleType = "int"
	TypeFloat    RuleType = "float"
	TypeIP       RuleType = "ip"
	TypeMAC      RuleType = "mac"
)

type RuleConfig struct {
	Type   string         `json:"type" yaml:"type"`
	Params map[string]any `json:"params" yaml:"params"`
}

func NewValidatorFromConfig(config ValidatorConfig) (*validator, error) {
	v := &validator{
		rules: make(map[string]RuleSet),
	}

	for field, ruleCfg := range config.Rules {
		ruleSet, err := createRuleSet(ruleCfg)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", field, err)
		}

		v.AddRule(field, ruleSet)
	}

	return v, nil
}

func createRuleSet(cfg RuleConfig) (RuleSet, error) {
	switch strings.ToLower(cfg.Type) {
	case string(TypeString):
		return parseStringRules(cfg.Params)
	case string(TypeEmail):
		return parseEmailRules(cfg.Params)
	case string(TypePassword):
		return parsePasswordRules(cfg.Params)
	case string(TypeInt):
		return parseIntRules(cfg.Params)
	case string(TypeFloat):
		return parseFloatRules(cfg.Params)
	case string(TypeIP):
		return parseIPRules(cfg.Params)
	case string(TypeMAC):
		return parseMACRules(cfg.Params)
	default:
		return nil, fmt.Errorf("unknown rule type: %s", cfg.Type)
	}
}

func toInt(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	case *int:
		return *v, true
	case *float32:
		return int(*v), true
	case *float64:
		return int(*v), true
	default:
		return 0, false
	}
}

// TODO: add validating parsed rules
func parseStringRules(params map[string]any) (*StringRules, error) {
	rules := &StringRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[StringRuleNameMinLen]; ok {
		if minLen, ok := toInt(v); ok {
			rules.MinLen = minLen
		}
	}

	if v, ok := params[StringRuleNameMaxLen]; ok {
		if maxLen, ok := toInt(v); ok {
			rules.MaxLen = maxLen
		}
	}

	if v, ok := params[StringRuleNameRegex]; ok {
		if regex, ok := v.(string); ok {
			rules.Regex = &regex
		}
	}

	if v, ok := params[StringRuleNameOnlyDigits]; ok {
		if onlyDigits, ok := v.(bool); ok {
			rules.OnlyDigits = onlyDigits
		}
	}

	if v, ok := params[StringRuleNameOnlyLetters]; ok {
		if onlyLetters, ok := v.(bool); ok {
			rules.OnlyLetters = onlyLetters
		}
	}

	if v, ok := params[StringRuleNameNoWhitespace]; ok {
		if noWhitespace, ok := v.(bool); ok {
			rules.NoWhitespace = noWhitespace
		}
	}

	if v, ok := params[StringRuleNameTrimSpace]; ok {
		if trimSpace, ok := v.(bool); ok {
			rules.TrimSpace = trimSpace
		}
	}

	if v, ok := params[StringRuleNameContains]; ok {
		contains, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid contains values: %w", err)
		}
		rules.Contains = contains
	}

	if v, ok := params[StringRuleNameNotContains]; ok {
		notContains, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid not contains values: %w", err)
		}
		rules.NotContains = notContains
	}

	if v, ok := params[StringRuleNameOneOf]; ok {
		oneOf, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid one of values: %w", err)
		}
		rules.OneOf = oneOf
	}

	if v, ok := params[StringRuleNameStartsWith]; ok {
		if startsWith, ok := v.(string); ok {
			rules.StartsWith = &startsWith
		}
	}

	if v, ok := params[StringRuleNameEndsWith]; ok {
		if endsWith, ok := v.(string); ok {
			rules.EndsWith = &endsWith
		}
	}

	if v, ok := params[StringRuleNameMinEntropy]; ok {
		if minEntropy, ok := v.(float64); ok {
			rules.MinEntropy = minEntropy
		}
	}

	return rules, nil
}

func parseMACRules(params map[string]any) (*MACRules, error) {
	rules := &MACRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[MACRuleNameFormat]; ok {
		formats, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid formats values: %w", err)
		}
		rules.Formats = formats
	}

	if v, ok := params[MACRuleNameCase]; ok {
		cases, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid cases values: %w", err)
		}
		rules.Cases = cases
	}

	if v, ok := params[MACRuleNameType]; ok {
		types, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid types values: %w", err)
		}
		rules.Types = types
	}

	if v, ok := params[MACRuleNameAllowZero]; ok {
		if allowZero, ok := v.(bool); ok {
			rules.AllowZero = &allowZero
		} else {
			if allowZero, ok := v.(*bool); ok {
				rules.AllowZero = allowZero
			}
		}
	}

	if v, ok := params[MACRuleNameAllowBroad]; ok {
		if allowBroad, ok := v.(bool); ok {
			rules.AllowBroadcast = &allowBroad
		} else {
			if allowBroad, ok := v.(*bool); ok {
				rules.AllowBroadcast = allowBroad
			}
		}
	}

	if v, ok := params[MACRuleNameAllowMulti]; ok {
		if allowMulti, ok := v.(bool); ok {
			rules.AllowMulticast = &allowMulti
		} else {
			if allowMulti, ok := v.(*bool); ok {
				rules.AllowMulticast = allowMulti
			}
		}
	}

	if v, ok := params[MACRuleNameOUI]; ok {
		oui, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid oui values: %w", err)
		}
		rules.OUIWhitelist = oui
	}

	if v, ok := params[MACRuleNameBlacklist]; ok {
		blacklist, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid not contains values: %w", err)
		}
		rules.Blacklist = blacklist
	}

	if v, ok := params[MACRuleNameMaxOctets]; ok {
		if maxOctets, ok := toInt(v); ok {
			rules.MaxOctets = &maxOctets
		} else {
			return nil, fmt.Errorf("invalid max octets value: %v", v)
		}
	}

	return rules, nil
}

// TODO: add validating parsed rules
func parseEmailRules(params map[string]any) (*EmailRules, error) {
	rules := &EmailRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[EmailRuleNameStrategy]; ok {
		if strategy, ok := v.(string); ok {
			if !validateStrategy(EmailValidationStrategy(strategy)) {
				return nil, fmt.Errorf("invalid email validation strategy: %s", strategy)
			}
			rules.Strategy = strategy
		}
	}

	if v, ok := params[EmailRuleNameMinDomainLen]; ok {
		if minLen, ok := toInt(v); ok {
			rules.MinDomainLen = minLen
		}
	}

	if v, ok := params[EmailRuleNameExcludedDomains]; ok {
		domains, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid excluded domains: %w", err)
		}
		rules.ExcludedDomains = domains
	}

	if v, ok := params[EmailRuleNameAllowedDomains]; ok {
		domains, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid allowed domains: %w", err)
		}
		rules.AllowedDomains = domains
	}

	if v, ok := params[EmailRuleNameRegexp]; ok {
		// global regex for email validation
		if regex, ok := v.(*string); ok {
			emailRegexp = regexp.MustCompile(*regex)
			rules.Regex = regex
		} else {
			if regex, ok := v.(string); ok {
				emailRegexp = regexp.MustCompile(regex)
				rules.Regex = &regex
			}
		}
	}

	return rules, nil
}

// TODO: add validating parsed rules
func parsePasswordRules(params map[string]any) (*PasswordRules, error) {
	rules := &PasswordRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[PasswordRuleNameMinLen]; ok {
		if minLen, ok := toInt(v); ok {
			rules.MinLen = minLen
		}
	}

	if v, ok := params[PasswordRuleNameMaxLen]; ok {
		if maxLen, ok := toInt(v); ok {
			rules.MaxLen = maxLen
		}
	}

	if v, ok := params[PasswordRuleNameMinUpper]; ok {
		if minUpper, ok := v.(int); ok {
			rules.MinUpper = minUpper
		}
	}

	if v, ok := params[PasswordRuleNameMinLower]; ok {
		if minLower, ok := v.(int); ok {
			rules.MinLower = minLower
		}
	}

	if v, ok := params[PasswordRuleNameMinDigits]; ok {
		if minNumbers, ok := v.(int); ok {
			rules.MinDigits = minNumbers
		}
	}

	if v, ok := params[PasswordRuleNameMinSpecial]; ok {
		if minSpecial, ok := v.(int); ok {
			rules.MinSpecial = minSpecial
		}
	}

	if v, ok := params[PasswordRuleNameSpecialChars]; ok {
		chars, err := ConvertToRuneArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid special chars: %w", err)
		}
		rules.SpecialChars = chars
	}

	if v, ok := params[PasswordRuleNameAllowedChars]; ok {
		chars, err := ConvertToRuneArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid allowed chars: %w", err)
		}
		rules.AllowedChars = chars
	}

	if v, ok := params[PasswordRuleNameDisallowedChars]; ok {
		chars, err := ConvertToRuneArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid disallowed chars: %w", err)
		}
		rules.DisallowedChars = chars
	}

	if v, ok := params[PasswordRuleNameMaxRepeatRun]; ok {
		if maxRepeat, ok := toInt(v); ok {
			rules.MaxRepeatRun = maxRepeat
		}
	}

	if v, ok := params[PasswordRuleNameDetectLinearPatterns]; ok {
		if detectLinearPatterns, ok := v.(bool); ok {
			rules.DetectLinearPatterns = detectLinearPatterns
		}
	}

	if v, ok := params[PasswordRuleNameBlacklist]; ok {
		blacklist, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid blacklist: %w", err)
		}
		rules.Blacklist = blacklist
	}

	if v, ok := params[PasswordRuleNameMinEntropy]; ok {
		if minEntropy, ok := v.(float64); ok {
			rules.MinEntropy = minEntropy
		}
	}

	return rules, nil
}

// TODO: add validating parsed rules
func parseIntRules(params map[string]any) (RuleSet, error) {
	rules := &IntRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[IntRuleNameMin]; ok {
		if min, ok := v.(int); ok {
			rules.Min = &min
		}
	}

	if v, ok := params[IntRuleNameMax]; ok {
		if max, ok := v.(int); ok {
			rules.Max = &max
		}
	}

	return rules, nil
}

// TODO: add validating parsed rules
func parseFloatRules(params map[string]any) (RuleSet, error) {
	rules := &FloatRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[FloatRuleNameMin]; ok {
		if min, ok := v.(float64); ok {
			rules.Min = &min
		}
	}

	if v, ok := params[FloatRuleNameMax]; ok {
		if max, ok := v.(float64); ok {
			rules.Max = &max
		}
	}

	return rules, nil
}

// TODO: add validating parsed rules
func parseIPRules(params map[string]any) (RuleSet, error) {
	rules := &IPRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[IPRuleNameVersion]; ok {
		if version, ok := v.(int); ok {
			rules.Version = version
		}
	}

	if v, ok := params[IPRuleNameAllowPrivate]; ok {
		if private, ok := v.(bool); ok {
			rules.AllowPrivate = private
		}
	}

	if v, ok := params[IPRuleNameAllowedSubnets]; ok {
		subnets, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid allowed subnets: %w", err)
		}
		rules.AllowedSubnets = subnets
	}

	if v, ok := params[IPRuleNameExcludedSubnets]; ok {
		subnets, err := ConvertToStringArray(v)
		if err != nil {
			return nil, fmt.Errorf("invalid excluded subnets: %w", err)
		}
		rules.ExcludedSubnets = subnets
	}

	return rules, nil
}

type validationContext struct {
	Path string
}

func (v *validator) Validate(data any) error {
	return v.validateRecursive(reflect.ValueOf(data), validationContext{Path: ""})
}

func (v *validator) validateRecursive(val reflect.Value, ctx validationContext) error {
	normalized := normalizePath(ctx.Path)
	ruleSet, hasRules := v.rules[normalized]

	if val.Kind() == reflect.Ptr {
		if val.IsNil() && hasRules {
			if err := ruleSet.Validate(nil); err != nil {
				return err
			}

			return nil
		}

		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		errs := NewValidationError()
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)

			tag := field.Tag.Get("sval")
			if tag == "" {
				continue
			}

			currentPath := tag
			if ctx.Path != "" {
				currentPath = ctx.Path + "." + tag
			}
			currentCtx := validationContext{Path: currentPath}

			if err := v.validateRecursive(fieldValue, currentCtx); err != nil {
				errs.AppendError(err.(*ValidationError))
			}
		}

		if errs.HasErrors() {
			return errs
		}
		return nil

	case reflect.Slice, reflect.Array:
		return v.validateSlice(val, ctx)

	default:
		normalized := normalizePath(ctx.Path)
		ruleSet, exists := v.rules[normalized]
		if !exists {
			return nil
		}

		var value any
		if val.CanInterface() {
			value = val.Interface()
		}

		if err := ruleSet.Validate(value); err != nil {
			return err
		}
		return nil
	}
}

func (v *validator) validateSlice(slice reflect.Value, ctx validationContext) error {
	errs := NewValidationError()
	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		newPath := ctx.Path + "[" + strconv.Itoa(i) + "]"
		newCtx := validationContext{Path: newPath}

		if err := v.validateRecursive(elem, newCtx); err != nil {
			errs.AppendError(err.(*ValidationError))
		}
	}

	if errs.HasErrors() {
		return errs
	}

	return nil
}

func normalizePath(path string) string {
	return indexRegex.ReplaceAllString(path, "[]")
}

func (v validator) String() string {
	var sb strings.Builder
	for field, rules := range v.rules {
		sb.WriteString(fmt.Sprintf("%s: %T\n", field, rules))
	}
	return sb.String()
}
