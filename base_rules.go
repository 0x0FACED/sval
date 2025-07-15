package sval

type BaseRuleName = string

const (
	BaseRuleNameRequired BaseRuleName = "required"
	BaseRuleNameType     BaseRuleName = "type"
)

type BaseRules struct {
	Required bool `json:"required" yaml:"required"`
}
