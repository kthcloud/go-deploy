package parsers

import (
	"bytes"
	"io"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra/nvidia"
)

var (
	// If we add other vendors that support DRA and have driver specific
	// configuration, we can add it here and just implement the
	// dra.ParamerersParser interface.
	//
	// Currently only NVIDIA has opaque driver parameters,
	// This is used to configure things like Sharing configuration.
	// Like decide what sharing strategy should be used,
	// and also provide configurations for that configuration.
	GpuOpaqueParamParsers = []dra.ParametersParser{

		nvidia.GPUConfigParametersParserImpl{},
	}
)

// Parse attempts to parse the provided raw data into the desired type T.
// It iterates through all registered vendor-specific parsers and uses CanParse().
// The first parser that returns true will be used for parsing.
//
// The type T must implement dra.OpaqueParams; otherwise, the function returns
// the zero value of T.
//
// Returns:
//   - Parsed value of type T if a suitable parser is found and parsing succeeds.
//   - ErrParserNotFound if no parser claims the data.
//   - ErrParserCastFailed if the parser returns a value that cannot be cast to T.
//   - Any error returned by the underlying parser.
//
// The raw io.Reader will be automatically rewound between CanParse and Parse,
// so you can safely pass a bytes.Reader or similar.
func Parse[T any](raw io.Reader) (T, error) {
	var zero T

	// Wrap raw in a bytes.Buffer if it's not already a ReadSeeker
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, raw); err != nil {
		return zero, err
	}
	reader := bytes.NewReader(buf.Bytes())

	// Ensure T implements dra.OpaqueParams
	if _, ok := any(zero).(dra.OpaqueParams); ok {
		for _, vendor := range GpuOpaqueParamParsers {
			// Rewind reader before CanParse
			if _, err := reader.Seek(0, io.SeekStart); err != nil {
				return zero, err
			}

			if vendor.CanParse(reader) {
				// Rewind again before Parse
				if _, err := reader.Seek(0, io.SeekStart); err != nil {
					return zero, err
				}

				parsed, err := vendor.Parse(reader)
				if err != nil {
					return zero, err
				}

				if casted, ok := parsed.(T); ok {
					return casted, nil
				}
				return zero, ErrParserCastFailed
			}
		}
	}

	return zero, ErrParserNotFound
}
