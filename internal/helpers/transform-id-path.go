package helpers

import (
	"fmt"
	"reflect"
	"strings"
)

func TransformsIdToPath(targets []string, record interface{}) {
	switch recordTyped := record.(type) {
	case []map[string]interface{}:
		for index, data := range recordTyped {
			for _, target := range targets {
				newKey := strings.Replace(target, "id", "path", 1)
				var pathType string
				if strings.Contains(target, "image") {
					pathType = "images"
				} else {
					pathType = "documents"
				}
				if value, exists := data[target]; exists {
					v := reflect.ValueOf(value)
					if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
						if !v.IsNil() {
							value = v.Elem().Interface()
						} else {
							value = nil
						}
					}

					if value != nil && value != 0 {
						recordTyped[index][newKey] = fmt.Sprintf("/api/v1/%s/%v", pathType, value)
					} else {
						recordTyped[index][newKey] = nil
					}
					delete(recordTyped[index], target)
				}
			}
		}
	case map[string]interface{}:
		for _, target := range targets {
			newKey := strings.Replace(target, "id", "path", 1)
			var pathType string
			if strings.Contains(target, "image") {
				pathType = "images"
			} else {
				pathType = "documents"
			}
			if value, exists := recordTyped[target]; exists {
				v := reflect.ValueOf(value)
				if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
					if !v.IsNil() {
						value = v.Elem().Interface()
					} else {
						value = nil
					}
				}
				if value != nil && value != 0 {
					recordTyped[newKey] = fmt.Sprintf("/api/v1/%s/%v", pathType, value)
				} else {
					recordTyped[newKey] = nil
				}
				delete(recordTyped, target)
			}
		}
	}
}
