package utils

import (
	"gopkg.in/yaml.v3"
)

func UnmarshalOption(val string) (any, error) {
	var v any
	err := yaml.Unmarshal([]byte(val), &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func MergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}
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
