package raygun4go

import (
	"fmt"
	"strings"
)

// arrayMapToStringMap converts a map[string][]string to a map[string]string
// by joining all values of the containing array and wrapping them in brackets
func arrayMapToStringMap(arrayMap map[string][]string) map[string]string {
	entries := make(map[string]string)
	for k, v := range arrayMap {
		if len(v) > 1 {
			entries[k] = fmt.Sprintf("[%s]", strings.Join(v, "; "))
		} else {
			entries[k] = v[0]
		}
	}
	return entries
}
