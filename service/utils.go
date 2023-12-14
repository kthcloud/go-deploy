package service

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go-deploy/pkg/subsystems"
	"reflect"
	"time"
)

type UpdateDbSubsystem func(string, string, interface{}) error

func ResetTimeFields[T any](input T) T {
	// Create a function to recursively reset time.Time fields
	var resetTimeFields func(v reflect.Value)
	resetTimeFields = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Ptr:
			resetTimeFields(v.Elem())
		case reflect.Interface:
			resetTimeFields(v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(time.Time{}) {
					resetTimeFields(field)
				} else if field.Type() == reflect.TypeOf(time.Time{}) {
					v.Field(i).Set(reflect.ValueOf(time.Time{}))
				}
			}
		}
	}

	// Make a copy of the input struct
	output := reflect.New(reflect.TypeOf(input)).Elem()
	output.Set(reflect.ValueOf(input))

	// Recursively reset time.Time fields
	resetTimeFields(output)

	return output.Interface().(T)
}

func areTimeFieldsEqual(a, b interface{}) bool {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() != valB.Kind() {
		return false
	}

	switch valA.Kind() {
	case reflect.Ptr, reflect.Interface:
		return areTimeFieldsEqual(valA.Elem().Interface(), valB.Elem().Interface())
	case reflect.Struct:
		for i := 0; i < valA.NumField(); i++ {
			fieldA := valA.Field(i)
			fieldB := valB.Field(i)
			if fieldA.Type() == reflect.TypeOf(time.Time{}) {
				if !fieldA.Interface().(time.Time).Equal(fieldB.Interface().(time.Time)) {
					return false
				}
			} else if fieldA.Kind() == reflect.Struct {
				if !areTimeFieldsEqual(fieldA.Interface(), fieldB.Interface()) {
					return false
				}
			}
		}
	}
	return true
}

func UpdateIfDiff[T subsystems.SsResource](dbResource T, fetchFunc func() (T, error), updateFunc func(T) (T, error), recreateFunc func(T) error) error {
	liveResource, err := fetchFunc()
	if err != nil {
		return nil
	}

	if subsystems.Nil(liveResource) {
		return recreateFunc(dbResource)
	}

	dbResourceCleaned := ResetTimeFields(dbResource)
	liveResourceCleaned := ResetTimeFields(liveResource)

	if subsystems.NotNil(liveResource) {
		timeEqual := areTimeFieldsEqual(dbResource, liveResource)
		restEqual := cmp.Equal(dbResourceCleaned, liveResourceCleaned, cmpopts.EquateEmpty())

		if timeEqual && restEqual {
			return nil
		}
	}

	liveResource, err = updateFunc(dbResource)
	if err != nil {
		return err
	}

	if subsystems.NotNil(liveResource) {
		liveResourceCleaned = ResetTimeFields(liveResource)

		timeEqual := areTimeFieldsEqual(dbResource, liveResource)
		restEqual := cmp.Equal(dbResourceCleaned, liveResourceCleaned, cmpopts.EquateEmpty())

		if subsystems.NotNil(liveResource) && timeEqual && restEqual {
			return nil
		}
	}

	return recreateFunc(dbResource)
}
