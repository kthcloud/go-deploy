package service

import (
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"reflect"
	"time"
)

type UpdateDbSubsystem func(string, string, interface{}) error

func Created(resource k8sModels.K8sResource) bool {
	return !NotCreated(resource)
}

func NotCreated(resource k8sModels.K8sResource) bool {
	if resource == nil || (reflect.ValueOf(resource).Kind() == reflect.Ptr && reflect.ValueOf(resource).IsNil()) {
		return true
	}

	return !resource.Created()
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

	if liveResource == nil {
		return recreateFunc(&dbResource)
	}

	dbResourceCleaned := ResetTimeFields(dbResource)
	liveResourceCleaned := ResetTimeFields(*liveResource)

	if liveResource != nil {
		timeEqual := areTimeFieldsEqual(dbResource, *liveResource)
		restEqual := reflect.DeepEqual(dbResourceCleaned, liveResourceCleaned)

		if timeEqual && restEqual {
			return nil
		}
	}

	err = updateFunc(&dbResource)
	if err != nil {
		return err
	}

	liveResource, err = fetchFunc()
	if err != nil {
		return nil
	}

	if liveResource != nil {
		liveResourceCleaned = ResetTimeFields(*liveResource)
		if liveResource != nil && areTimeFieldsEqual(dbResource, *liveResource) && reflect.DeepEqual(liveResourceCleaned, dbResourceCleaned) {
			return nil
		}
	}

	return recreateFunc(&dbResource)
}
