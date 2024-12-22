package openapi

import (
	"reflect"
	"runtime"
	"strings"
)

// FuncName returns the name of a function and the name with package path
func FuncName(f interface{}) string {
	return strings.TrimSuffix(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "-fm")
}

// transform camelCase to human-readable string
func camelToHuman(s string) string {
	result := strings.Builder{}
	for i, r := range s {
		if 'A' <= r && r <= 'Z' {
			if i > 0 {
				result.WriteRune(' ')
			}
			result.WriteRune(r + 'a' - 'A') // 'A' -> 'a
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
