package hashutils_test

import (
	"testing"

	"github.com/kthcloud/go-deploy/utils/hashutils"
	"github.com/stretchr/testify/require"
)

func TestHashDeterministicJSON(t *testing.T) {
	type Example struct {
		Name  string         `json:"name"`
		Attrs map[string]any `json:"attrs"`
	}

	e1 := Example{
		Name: "Alice",
		Attrs: map[string]any{
			"b": 2,
			"a": 1,
		},
	}

	e2 := Example{
		Name: "Alice",
		Attrs: map[string]any{
			"a": 1,
			"b": 2,
		},
	}

	h1, err := hashutils.HashDeterministicJSON(e1)
	require.NoError(t, err)

	h2, err := hashutils.HashDeterministicJSON(e2)
	require.NoError(t, err)

	require.Equal(t, h1, h2, "Hashes should be equal even if map keys are reordered")

	extra := map[string]any{"version": 1}
	h3, err := hashutils.HashDeterministicJSON(e1, extra)
	require.NoError(t, err)

	h4, err := hashutils.HashDeterministicJSON(e2, extra)
	require.NoError(t, err)

	require.Equal(t, h3, h4, "Hashes with extras should also match")

	require.NotEqual(t, h1, h3, "Hashes with extras should differ from without extras")

	nested1 := map[string]any{
		"user": map[string]any{
			"name": "Alice",
			"age":  30,
		},
		"tags": []any{"dev", "ops"},
	}

	nested2 := map[string]any{
		"tags": []any{"dev", "ops"},
		"user": map[string]any{
			"age":  30,
			"name": "Alice",
		},
	}

	h5, err := hashutils.HashDeterministicJSON(nested1)
	require.NoError(t, err)
	h6, err := hashutils.HashDeterministicJSON(nested2)
	require.NoError(t, err)
	require.Equal(t, h5, h6, "Hashes should match for nested maps with different key orders")

	h7, err := hashutils.HashDeterministicJSON(nested1, extra, "extra string")
	require.NoError(t, err)

	h8, err := hashutils.HashDeterministicJSON(nested2, extra, "extra string")
	require.NoError(t, err)
	require.Equal(t, h7, h8, "Hashes with multiple extras should match")
}
