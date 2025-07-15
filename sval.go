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
	Rules map[string]RuleConfig `yaml:"rules" json:"rules"`
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
	TypeNumber   RuleType = "number"
	TypeIP       RuleType = "ip"
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
	case string(TypeNumber):
		return parseNumberRules(cfg.Params)
	case string(TypeIP):
		return parseIPRules(cfg.Params)
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
	default:
		return 0, false
	}
}

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
			rules.Regex = regex
		}
	}

	if v, ok := params[StringRuleNameAlphanum]; ok {
		if alphanum, ok := v.(bool); ok {
			rules.Alphanum = alphanum
		}
	}

	return rules, nil
}

func parseEmailRules(params map[string]any) (*EmailRules, error) {
	rules := &EmailRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[EmailRuleNameRFC]; ok {
		if rfc, ok := v.(bool); ok {
			rules.RFC = rfc
		}
	}

	if v, ok := params[EmailRuleNameMinDomainLen]; ok {
		if minLen, ok := toInt(v); ok {
			rules.MinDomainLen = minLen
		}
	}

	if v, ok := params[EmailRuleNameExcludedDomains]; ok {
		if domains, ok := v.([]any); ok {
			for _, d := range domains {
				if domain, ok := d.(string); ok {
					rules.ExcludedDomains = append(rules.ExcludedDomains, domain)
				}
			}
		}
	}

	if v, ok := params[EmailRuleNameAllowedDomains]; ok {
		if domains, ok := v.([]any); ok {
			for _, d := range domains {
				if domain, ok := d.(string); ok {
					rules.AllowedDomains = append(rules.AllowedDomains, domain)
				}
			}
		}
	}

	return rules, nil
}

func parsePasswordRules(params map[string]any) (*PasswordRules, error) {
	rules := &PasswordRules{}

	if v, ok := params["required"]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params["min_len"]; ok {
		if minLen, ok := toInt(v); ok {
			rules.MinLen = minLen
		}
	}

	if v, ok := params["max_len"]; ok {
		if maxLen, ok := toInt(v); ok {
			rules.MaxLen = maxLen
		}
	}

	if v, ok := params["require_upper"]; ok {
		if require, ok := v.(bool); ok {
			rules.RequireUpper = require
		}
	}

	if v, ok := params["require_lower"]; ok {
		if require, ok := v.(bool); ok {
			rules.RequireLower = require
		}
	}

	if v, ok := params["require_number"]; ok {
		if require, ok := v.(bool); ok {
			rules.RequireNumber = require
		}
	}

	if v, ok := params["require_special"]; ok {
		if require, ok := v.(bool); ok {
			rules.RequireSpecial = require
		}
	}

	return rules, nil
}

func parseNumberRules(params map[string]any) (RuleSet, error) {
	rules := &NumberRules{}

	if v, ok := params[BaseRuleNameRequired]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params[NumberRuleNameMin]; ok {
		if min, ok := v.(float64); ok {
			rules.Min = &min
		} else if min, ok := v.(int); ok {
			fmin := float64(min)
			rules.Min = &fmin
		}
	}

	if v, ok := params[NumberRuleNameMax]; ok {
		if max, ok := v.(float64); ok {
			rules.Max = &max
		} else if max, ok := v.(int); ok {
			fmax := float64(max)
			rules.Max = &fmax
		}
	}

	return rules, nil
}

// TODO: Implement parseIPRules function
func parseIPRules(params map[string]any) (RuleSet, error) {
	return &IPRules{}, nil
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
