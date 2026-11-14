//go:build !goverter

package model

func init() {
	dtoConverter = &ConverterImpl{}
}
