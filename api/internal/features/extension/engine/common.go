package engine

import (
	"fmt"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// replaceVars substitutes tokens like {{ varName }} in the input string
// using the provided vars map. Missing vars are left unchanged.
// Variable values are shell-quoted to prevent injection.
func replaceVars(in string, vars map[string]interface{}) string {
	out := in
	for k, v := range vars {
		token := fmt.Sprintf("{{ %s }}", k)
		out = strings.ReplaceAll(out, token, fmt.Sprint(v))
	}
	return out
}

// replaceVarsSafe substitutes tokens like {{ varName }} in the input string,
// shell-quoting each value to prevent command injection.
func replaceVarsSafe(in string, vars map[string]interface{}) string {
	out := in
	for k, v := range vars {
		token := fmt.Sprintf("{{ %s }}", k)
		out = strings.ReplaceAll(out, token, utils.ShellQuote(fmt.Sprint(v)))
	}
	return out
}

// validateShellArgs validates that all string values going into a shell command
// do not contain shell metacharacters.
func validateShellArgs(args map[string]string) error {
	for name, value := range args {
		if value == "" {
			continue
		}
		if err := utils.ValidateShellArg(value, name); err != nil {
			return err
		}
	}
	return nil
}
