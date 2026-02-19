package engine

import (
	"fmt"
	"strings"
)

// replaceVars substitutes tokens like {{ varName }} in the input string
// using the provided vars map. Missing vars are left unchanged.
func replaceVars(in string, vars map[string]interface{}) string {
	out := in
	for k, v := range vars {
		token := fmt.Sprintf("{{ %s }}", k)
		out = strings.ReplaceAll(out, token, fmt.Sprint(v))
	}
	return out
}
