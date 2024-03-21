package subsystems

import "reflect"

// SsResource is an interface that all subsystem resources should implement.
// It is used to ensure that all subsystem resources can be compared and copied,
// which is essential when, for example, repairing them or storing them as placeholders
// in the database.
type SsResource interface {
	Created() bool
	IsPlaceholder() bool
}

// CopyValue copies the value of src to dst.
func CopyValue(src, dst SsResource) {
	reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(src).Elem())
}

// Nil returns true if the model is nil.
func Nil(resource SsResource) bool {
	return !NotNil(resource)
}

// Created returns true if the model is created.
func Created(resource SsResource) bool {
	return !NotCreated(resource)
}

// NotNil returns true if the model is not nil.
func NotNil(resource SsResource) bool {
	if resource == nil || (reflect.ValueOf(resource).Kind() == reflect.Ptr && reflect.ValueOf(resource).IsNil()) {
		return false
	}
	return true
}

// NotCreated returns true if the model is not created.
func NotCreated(resource SsResource) bool {
	return Nil(resource) || !resource.Created()
}
