package sval

type BaseRuleName = string

const (
	BaseRuleNameRequired BaseRuleName = "required"
	BaseRuleType         BaseRuleName = "type"
)

type BaseRules struct {
	Required bool `json:"required" yaml:"required"`
}
