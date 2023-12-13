package subsystems

import "reflect"

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
