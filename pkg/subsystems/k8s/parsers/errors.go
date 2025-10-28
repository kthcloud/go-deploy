package parsers

import "errors"

var (
	ErrParserNotFound   = errors.New("no parser found")
	ErrParserCastFailed = errors.New("parsed value cannot be cast to T")
)
