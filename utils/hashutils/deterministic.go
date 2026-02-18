package hashutils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// HashJSONWithExtras computes a SHA256 hash of v plus any additional values.
// This is deterministic even for maps.
// extras can be any Go value (string, int, struct, map, etc.)
func HashDeterministicJSON(v any, extras ...any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var intermediate any
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return "", err
	}

	normalized, err := normalize(intermediate)
	if err != nil {
		return "", err
	}

	toHash := []any{normalized}
	for _, e := range extras {
		data, err := json.Marshal(e)
		if err != nil {
			return "", err
		}
		var intermediate any
		if err := json.Unmarshal(data, &intermediate); err != nil {
			return "", err
		}
		norm, err := normalize(intermediate)
		if err != nil {
			return "", err
		}
		toHash = append(toHash, norm)
	}

	// Marshal combined slice to deterministic JSON
	finalData, err := json.Marshal(toHash)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(finalData)
	return hex.EncodeToString(hash[:]), nil
}

// normalize recursively sorts maps for deterministic JSON
func normalize(v any) (any, error) {
	switch val := v.(type) {
	case map[string]any:
		sorted := make(map[string]any, len(val))
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			norm, err := normalize(val[k])
			if err != nil {
				return nil, err
			}
			sorted[k] = norm
		}
		return sorted, nil
	case []any:
		normalizedSlice := make([]any, len(val))
		for i, elem := range val {
			norm, err := normalize(elem)
			if err != nil {
				return nil, err
			}
			normalizedSlice[i] = norm
		}
		return normalizedSlice, nil
	default:
		return val, nil
	}
}
