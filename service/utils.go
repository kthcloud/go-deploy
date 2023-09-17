package service

import (
	"reflect"
	"time"
)

func deepEqual[T any](a, b *T) bool {
	aValue := reflect.ValueOf(*a)
	bValue := reflect.ValueOf(*b)

	for i := 0; i < aValue.NumField(); i++ {
		aField := aValue.Field(i)
		bField := bValue.Field(i)
		fieldType := aValue.Type().Field(i)

		if fieldType.PkgPath != "" {
			continue
		}

		if fieldType.Type == reflect.TypeOf(time.Time{}) {
			if !aField.Interface().(time.Time).Equal(bField.Interface().(time.Time)) {
				return false
			}
		} else if fieldType.Type.Kind() == reflect.Struct {
			aInt := aField.Interface()
			bInt := bField.Interface()

			if deepEqual(&aInt, &bInt) {
				return true
			}
		} else {
			if aField.Interface() != bField.Interface() {
				return false
			}
		}
	}

	return true
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

func UpdateIfDiff[T any](dbResource T, fetchFunc func() (*T, error), updateFunc func(*T) error, recreateFunc func(*T) error) error {
	liveResource, err := fetchFunc()
	if err != nil {
		return nil
	}

	dbResourceCleaned := ResetTimeFields(dbResource)
	liveResourceCleaned := ResetTimeFields(*liveResource)

	if liveResource != nil && areTimeFieldsEqual(dbResource, *liveResource) && reflect.DeepEqual(liveResourceCleaned, dbResourceCleaned) {
		return nil
	}

	err = updateFunc(&dbResource)
	if err != nil {
		return err
	}

	liveResource, err = fetchFunc()
	if err != nil {
		return nil
	}
	liveResourceCleaned = ResetTimeFields(*liveResource)

	if liveResource != nil && areTimeFieldsEqual(dbResource, *liveResource) && reflect.DeepEqual(liveResourceCleaned, dbResourceCleaned) {
		return nil
	}

	return recreateFunc(&dbResource)
}
