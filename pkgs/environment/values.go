package environment

import (
	"maps"
	"strings"

	"github.com/reMarkable/helmfile-nix/pkgs/transform"
)

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
	var err error
	if strings.Contains(dottedKey, ".") {
		// Nested value
		// Split the key into parts
		mref := m
		keys := strings.Split(dottedKey, ".")
		for i, key := range keys {
			if i == len(keys)-1 {
				if mref[key], err = transform.UnmarshalOption(value); err != nil {
					return err
				}
			} else {
				if _, ok := mref[key]; !ok {
					mref[key] = make(map[string]any)
				}
				mref = mref[key].(map[string]any)
			}
		}
	} else {
		if m[dottedKey], err = transform.UnmarshalOption(value); err != nil {
			return err
		}
	}
	return nil
}
