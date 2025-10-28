package dra

import "io"

type OpaqueParams interface {
	MetaAPIVersion() string
	MetaKind() string
}

type ParametersParser interface {
	CanParse(raw io.Reader) bool
	Parse(raw io.Reader) (OpaqueParams, error)
}

func OptParseToOpaqueParams() OpaqueParams {
	var opt OpaqueParams
	return opt
}
