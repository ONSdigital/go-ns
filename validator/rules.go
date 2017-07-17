package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

const emailRegex = `^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,6}$`

// FieldValidationErr is returned when a field fails validation
type FieldValidationErr struct {
	error
}

// Field represents a field to validate
type field struct {
	ID    string `json:"id"`
	Rules []rule `json:"rules"`
}

// Rule represents a rule to validate within a field
type rule struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value,omitempty"`
}

// RulesList is an extendable map containing rule functions which return an error if the
// condition is not met
var RulesList = map[string]func(...interface{}) error{
	"min_length": minLength,
	"max_length": maxLength,
	"email":      email,
	"not_empty":  notEmpty,
}

func minLength(vars ...interface{}) error {
	var s string
	var l float64
	var ok bool

	if s, ok = vars[0].(string); !ok {
		return errors.New("first param to minLength must be string")
	}

	if l, ok = vars[1].(float64); !ok {
		return errors.New("second param to minLength must be number")
	}

	if len(s) < int(l) {
		return FieldValidationErr{fmt.Errorf("value: %s, must be at least %d characters", s, int(l))}
	}
	return nil
}

func maxLength(vars ...interface{}) error {
	var s string
	var l float64
	var ok bool

	if s, ok = vars[0].(string); !ok {
		return errors.New("first param to maxLength must be string")
	}

	if l, ok = vars[1].(float64); !ok {
		return errors.New("second param to maxLength must be number")
	}

	if len(s) > int(l) {
		return FieldValidationErr{fmt.Errorf("value: %s, must be at most %d characters", s, int(l))}
	}
	return nil
}

func email(vars ...interface{}) error {
	var e string
	var ok bool

	if e, ok = vars[0].(string); !ok {
		return errors.New("first parameter to email must be a string")
	}

	if ok, err := regexp.MatchString(emailRegex, e); !ok || err != nil {
		return FieldValidationErr{fmt.Errorf("email: %s is not a valid email address", e)}
	}
	return nil
}

func notEmpty(vars ...interface{}) error {
	v := reflect.ValueOf(vars[0])

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return errors.New("first param to notEmpty must be a slice or array")
	}

	if v.Len() == 0 {
		return FieldValidationErr{errors.New("slice must not be empty")}
	}
	return nil
}
