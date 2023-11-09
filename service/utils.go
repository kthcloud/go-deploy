package service

import (
	"reflect"
	"time"
)

type UpdateDbSubsystem func(string, string, interface{}) error

type SsResource interface {
	Created() bool
	IsPlaceholder() bool
}

func CopyValue(src, dst SsResource) {
	reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(src).Elem())
}

func Nil(resource SsResource) bool {
	return !NotNil(resource)
}

func Created(resource SsResource) bool {
	return !NotCreated(resource)
}

func NotNil(resource SsResource) bool {
	if resource == nil || (reflect.ValueOf(resource).Kind() == reflect.Ptr && reflect.ValueOf(resource).IsNil()) {
		return false
	}
	return true
}

func NotCreated(resource SsResource) bool {
	return Nil(resource) || !resource.Created()
}

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

func UpdateIfDiff[T SsResource](dbResource T, fetchFunc func() (T, error), updateFunc func(T) (T, error), recreateFunc func(T) error) error {
	liveResource, err := fetchFunc()
	if err != nil {
		return nil
	}

	if Nil(liveResource) {
		return recreateFunc(dbResource)
	}

	dbResourceCleaned := ResetTimeFields(dbResource)
	liveResourceCleaned := ResetTimeFields(liveResource)

	if NotNil(liveResource) {
		timeEqual := areTimeFieldsEqual(dbResource, liveResource)
		restEqual := reflect.DeepEqual(dbResourceCleaned, liveResourceCleaned)

		if timeEqual && restEqual {
			return nil
		}
	}

	liveResource, err = updateFunc(dbResource)
	if err != nil {
		return err
	}

	if NotNil(liveResource) {
		liveResourceCleaned = ResetTimeFields(liveResource)
		if NotNil(liveResource) && areTimeFieldsEqual(dbResource, liveResource) && reflect.DeepEqual(liveResourceCleaned, dbResourceCleaned) {
			return nil
		}
	}

	return recreateFunc(dbResource)
}
