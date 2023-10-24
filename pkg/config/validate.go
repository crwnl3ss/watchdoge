package config

import (
	"fmt"
	"reflect"

	"github.com/crwnl3ss/watchdoge/pkg/check"
)

// TODO: need recurcive calls
func Validate(k check.Checker) error {
	T := reflect.TypeOf(k).Elem()
	V := reflect.Indirect(reflect.ValueOf(k))
	for i := 0; i < T.NumField(); i++ {
		validation := T.Field(i).Tag.Get("validate")
		val := V.Field(i)
		switch validation {
		case "non-empty":
			if val.String() == "" {
				return fmt.Errorf("validation error, field %s is empty", T.Field(i).Name)
			}
		case "non-nil":
			if val.IsNil() {
				return fmt.Errorf("validation error, field %s is empty", T.Field(i).Name)
			}
		case "non-zero":
			if val.IsZero() {
				return fmt.Errorf("validation error, field %s is zero", T.Field(i).Name)
			}
		}
	}
	return nil
}
