package parsers

import "io"

// Parser is a generic interface for parsers
type Parser[T any] interface {
	CanParse(raw io.Reader) bool
	Parse(raw io.Reader, out T) error
}
