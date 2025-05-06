package nixchart

import (
	"errors"
	"fmt"
	"os"
	"reflect"
)

func RenderCharts(obj map[string]any) (map[string]any, error) {
	releasesValue := reflect.ValueOf(obj["releases"])
	if releasesValue.Kind() != reflect.Slice {
		return nil, errors.New("releases is not a slice")
	}
	for i := range releasesValue.Len() {
		element := releasesValue.Index(i)
		if element.Kind() != reflect.Map {
			mappedElement, ok := element.Interface().(map[string]any)
			if !ok {
				return nil, fmt.Errorf("release at index %d is not a map[string]interface{}: %v", i, element)
			}
			element = reflect.ValueOf(mappedElement)
		}
		nixChart := element.MapIndex(reflect.ValueOf("nixChart"))
		if nixChart.IsValid() {
			evalChart(nixChart.String())
			fmt.Println("Rendering chart:", nixChart)
			element.SetMapIndex(reflect.ValueOf("chart"), reflect.ValueOf(evalChart))
			element.SetMapIndex(reflect.ValueOf("nixChart"), reflect.ValueOf(nil))
		}
	}
	return obj, nil
}

func evalChart(chart string) string {
	chartDir := os.TempDir()
	return chartDir
}
