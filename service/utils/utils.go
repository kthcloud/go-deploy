package utils

import (
	"reflect"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems"
)

// UpdateDbSubsystem is a common interface that services use to update subsystem resources
type UpdateDbSubsystem func(string, string, interface{}) error

// ResetTimeFields resets all time.Time fields in a struct to the zero value
// This is done when comparing structs to ignore time.Time fields, since they are not
// guaranteed to be equal using == (they have their own Equal() method)
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
	// If it is a pointer, we need to dereference it firs
	inputCopy := reflect.ValueOf(input)
	wasPointer := false
	if inputCopy.Kind() == reflect.Ptr {
		inputCopy = inputCopy.Elem()
		wasPointer = true
	}

	// Create a new instance of the input struct
	output := reflect.New(inputCopy.Type()).Elem()

	// Copy the input struct to the output struct
	output.Set(inputCopy)

	// Recursively reset time.Time fields
	resetTimeFields(output)

	if wasPointer {
		return output.Addr().Interface().(T)
	} else {
		return output.Interface().(T)
	}
}

// areTimeFieldsEqual compares two structs time.Time fields
// This is done when comparing structs to ignore time.Time fields, since they are not
// guaranteed to be equal using == (they have their own Equal() method)
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

// UpdateIfDiff is a common function that services use to update subsystem resources if there is a diff
// in the database and live model.
//
// This is the core functionality that enables repairing subsystem resources.
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
		} else {
			log.Debugln("Resources are not equal, updating")
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
		} else {
			log.Debugln("Resources are not equal after update, recreating")
		}
	}

	return recreateFunc(dbResource)
}

// GetFirstOrDefault gets the first element of a slice, or the default value of the type if the slice is empty
func GetFirstOrDefault[T any](variable []T) T {
	if len(variable) > 0 {
		return variable[0]
	}
	defaultVal := new(T)
	return *defaultVal
}

func FirstNonZero[T comparable](vars ...T) T {
	var zero T
	for _, v := range vars {
		if v != zero {
			return v
		}
	}
	return zero
}
