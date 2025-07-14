package sval

import (
	"encoding/json"
	"strings"
)

type ValidationError struct {
	Errors []*valError `json:"errors" yaml:"errors"`
}

type valError struct {
	ID         int    `json:"id" yaml:"id"`
	Rule       string `json:"rule" yaml:"rule"`
	RuleValues any    `json:"rule_values,omitempty" yaml:"rule_values,omitempty"`
	Message    string `json:"message" yaml:"message"`
}

// JSON formatted as string
func (e *ValidationError) Error() string {
	var sb strings.Builder
	if json.NewEncoder(&sb).Encode(e) != nil {
		return "error encoding validation errors"
	}

	return sb.String()
}

func (e *ValidationError) AddError(rule string, ruleValue any, message string) {
	e.Errors = append(e.Errors, &valError{
		ID:         len(e.Errors) + 1,
		Rule:       rule,
		RuleValues: ruleValue,
		Message:    message,
	})
}

func (e *ValidationError) AppendError(err *ValidationError) {
	if err == nil || len(err.Errors) == 0 {
		return
	}

	baseID := len(e.Errors)
	for _, verr := range err.Errors {
		newError := &valError{
			ID:         baseID + verr.ID,
			Rule:       verr.Rule,
			RuleValues: verr.RuleValues,
			Message:    verr.Message,
		}
		e.Errors = append(e.Errors, newError)
	}
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make([]*valError, 0),
	}
}
