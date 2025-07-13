package sval

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

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

	if v, ok := params["regex"]; ok {
		if regex, ok := v.(string); ok {
			rules.Regex = regex
		}
	}

	if v, ok := params["alphanum"]; ok {
		if alphanum, ok := v.(bool); ok {
			rules.Alphanum = alphanum
		}
	}

	return rules, nil
}

func parseEmailRules(params map[string]any) (*EmailRules, error) {
	rules := &EmailRules{}

	if v, ok := params["required"]; ok {
		if required, ok := v.(bool); ok {
			rules.Required = required
		}
	}

	if v, ok := params["rfc"]; ok {
		if rfc, ok := v.(bool); ok {
			rules.RFC = rfc
		}
	}

	if v, ok := params["min_domain_len"]; ok {
		if minLen, ok := toInt(v); ok {
			rules.MinDomainLen = minLen
		}
	}

	if v, ok := params["excluded_domains"]; ok {
		if domains, ok := v.([]any); ok {
			for _, d := range domains {
				if domain, ok := d.(string); ok {
					rules.ExcludedDomains = append(rules.ExcludedDomains, domain)
				}
			}
		}
	}

	if v, ok := params["allowed_domains"]; ok {
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
	return &NumberRules{}, nil
}

func parseIPRules(params map[string]any) (RuleSet, error) {
	return &IPRules{}, nil
}

func (v *validator) Validate(data any) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.New("validator supports only struct types")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		tag := field.Tag.Get("sval")
		if tag == "" {
			continue
		}

		ruleSet, exists := v.rules[tag]
		if !exists {
			continue
		}

		var value any
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				value = nil
			} else {
				value = fieldValue.Elem().Interface()
			}
		} else {
			value = fieldValue.Interface()
		}

		if err := ruleSet.Validate(value); err != nil {
			return fmt.Errorf("%s: %w", field.Name, err)
		}
	}

	return nil
}
