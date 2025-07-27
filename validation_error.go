package sval

import (
	"encoding/json"
)

type ValidationError struct {
	Errors []*valError `json:"errors" yaml:"errors"`
}

type valError struct {
	Field      string `json:"field" yaml:"field"`
	Rule       string `json:"rule" yaml:"rule"`
	RuleValues any    `json:"rule_values,omitempty" yaml:"rule_values,omitempty"`
	Provided   any    `json:"provided,omitempty" yaml:"provided,omitempty"`
	Message    string `json:"message" yaml:"message"`
}

// JSON formatted as string
func (e *ValidationError) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return "error encoding validation errors: " + err.Error()
	}
	return string(data)
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make([]*valError, 0),
	}
}

func NewValidationErrorWithField(field string) *ValidationError {
	err := NewValidationError()
	err.AddContextToErrors(field)
	return err
}

func (e *ValidationError) AddContextToErrors(field string) {
	for _, err := range e.Errors {
		if err.Field == "" {
			err.Field = field
		} else if field != "" {
			err.Field = field + "." + err.Field
		}
	}
}

func (e *ValidationError) AddError(rule string, ruleValue, provided any, message string) {
	e.Errors = append(e.Errors, &valError{
		Field:      "",
		Rule:       rule,
		RuleValues: ruleValue,
		Provided:   provided,
		Message:    message,
	})
}

func (e *ValidationError) AppendError(err *ValidationError) {
	if err == nil || len(err.Errors) == 0 {
		return
	}

	if !e.HasErrors() {
		e.Errors = make([]*valError, 0, len(err.Errors))
	}

	e.Errors = append(e.Errors, err.Errors...)

}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}
