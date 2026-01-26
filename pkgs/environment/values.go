package environment

import (
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/reMarkable/helmfile-nix/pkgs/transform"
)

// ErrNestedKeyNotMap is returned when a nested key is not a map.
var ErrNestedKeyNotMap = errors.New("nested key is not a map")

func MergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	maps.Copy(out, a)
	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func SetNestedMapValue(m map[string]any, dottedKey string, value string) error {
	if !strings.Contains(dottedKey, ".") {
		val, err := transform.UnmarshalOption(value)
		if err != nil {
			return err
		}
		m[dottedKey] = val
		return nil
	}

	return setNestedValue(m, strings.Split(dottedKey, "."), value)
}

func setNestedValue(m map[string]any, keys []string, value string) error {
	mref := m
	for i, key := range keys {
		if i == len(keys)-1 {
			val, err := transform.UnmarshalOption(value)
			if err != nil {
				return err
			}
			mref[key] = val
			return nil
		}

		nestedMap, err := ensureNestedMap(mref, key)
		if err != nil {
			return err
		}
		mref = nestedMap
	}
	return nil
}

func ensureNestedMap(m map[string]any, key string) (map[string]any, error) {
	if _, ok := m[key]; !ok {
		newMap := make(map[string]any)
		m[key] = newMap
		return newMap, nil
	}
	nestedMap, ok := m[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNestedKeyNotMap, key)
	}
	return nestedMap, nil
}
