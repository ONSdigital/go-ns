package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/schema"
)

// ErrFormValidationFailed if form validation has failed, if this is ever returned,
// corresponding call to GetFieldErrors should be made to check for individual field errors
var ErrFormValidationFailed = errors.New("form validation failed, check field errors")

// FormValidator represents a form validator with appropriate validator method
type FormValidator struct {
	rulesJSON []byte
	fieldErrs map[string][]error
}

// New creates a new FormValidator
func New(r io.Reader) (fv FormValidator, err error) {
	fv.fieldErrs = make(map[string][]error)

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, r)
	if err == nil {
		fv.rulesJSON = buf.Bytes()
	}

	return
}

// GetFieldErrors returns a list of field errors (if any) after validation
func (fv FormValidator) GetFieldErrors() map[string][]error {
	return fv.fieldErrs
}

// Validate will validate a request form parameters against a provided struct
func (fv FormValidator) Validate(req *http.Request, s interface{}) error {
	if err := decodeRequest(req, s); err != nil {
		return err
	}

	v := getValue(s)

	var flds []Field
	if err := json.Unmarshal(fv.rulesJSON, &flds); err != nil {
		return err
	}

	for i := 0; i < v.NumField(); i++ {
		tag := string(v.Type().Field(i).Tag.Get("schema"))

		if tag == "" {
			log.Debug("field missing schema tag", log.Data{"field": v.Type().Field(i).Name})
			continue
		}

		for _, fld := range flds {
			if fld.ID == tag {
				for _, rule := range fld.Rules {
					fn, ok := RulesList[rule.Name]
					if !ok {
						return fmt.Errorf("rule name: %s, missing corresponding validation function", rule.Name)
					}

					if err := fn(v.Field(i).Interface(), rule.Value); err != nil {
						if _, ok := err.(FieldValidationErr); !ok {
							return err
						}
						fv.fieldErrs[tag] = append(fv.fieldErrs[tag], err)
					}
				}
			}
		}
	}

	if len(fv.fieldErrs) > 0 {
		return ErrFormValidationFailed
	}

	return nil
}

func getValue(s interface{}) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(s))
}

func decodeRequest(req *http.Request, s interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	decoder := schema.NewDecoder()
	return decoder.Decode(s, req.PostForm)
}
