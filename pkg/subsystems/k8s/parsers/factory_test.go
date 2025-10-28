package parsers_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra"
)

type mockParser struct {
	canParseResult bool
	parseResult    dra.OpaqueParams
	parseErr       error
}

func (m mockParser) CanParse(r io.Reader) bool {
	return m.canParseResult
}

func (m mockParser) Parse(r io.Reader) (dra.OpaqueParams, error) {
	return m.parseResult, m.parseErr
}

type mockParams struct {
	Name       string
	APIVersion string
	Kind       string
}

func (m mockParams) MetaAPIVersion() string {
	return m.APIVersion
}

func (m mockParams) MetaKind() string {
	return m.Kind
}

func (m mockParams) Opaque() {}

func TestParse_Success(t *testing.T) {
	// Override global parser slice for testing
	original := parsers.GpuOpaqueParamParsers
	defer func() { parsers.GpuOpaqueParamParsers = original }()

	expected := mockParams{Name: "test", APIVersion: "resource.test.kth.se/v1", Kind: "Mock"}
	parsers.GpuOpaqueParamParsers = []dra.ParametersParser{
		mockParser{
			canParseResult: true,
			parseResult:    expected,
			parseErr:       nil,
		},
	}

	data := bytes.NewReader([]byte("some raw data"))
	result, err := parsers.Parse[mockParams](data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != expected.Name {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParse_NoParserFound(t *testing.T) {
	original := parsers.GpuOpaqueParamParsers
	defer func() { parsers.GpuOpaqueParamParsers = original }()

	parsers.GpuOpaqueParamParsers = []dra.ParametersParser{
		mockParser{canParseResult: false},
	}

	data := bytes.NewReader([]byte("data"))
	_, err := parsers.Parse[mockParams](data)
	if !errors.Is(err, parsers.ErrParserNotFound) {
		t.Errorf("expected ErrParserNotFound, got %v", err)
	}
}

func TestParse_CastFailed(t *testing.T) {
	original := parsers.GpuOpaqueParamParsers
	defer func() { parsers.GpuOpaqueParamParsers = original }()

	parsers.GpuOpaqueParamParsers = []dra.ParametersParser{
		mockParser{
			canParseResult: true,
			parseResult:    nil,
			parseErr:       nil,
		},
	}

	data := bytes.NewReader([]byte("data"))
	_, err := parsers.Parse[mockParams](data)
	if !errors.Is(err, parsers.ErrParserCastFailed) {
		t.Errorf("expected ErrParserCastFailed, got %v", err)
	}
}

func TestParse_ParseError(t *testing.T) {
	original := parsers.GpuOpaqueParamParsers
	defer func() { parsers.GpuOpaqueParamParsers = original }()

	expectedErr := errors.New("parse failed")
	parsers.GpuOpaqueParamParsers = []dra.ParametersParser{
		mockParser{
			canParseResult: true,
			parseResult:    nil,
			parseErr:       expectedErr,
		},
	}

	data := bytes.NewReader([]byte("data"))
	_, err := parsers.Parse[mockParams](data)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
