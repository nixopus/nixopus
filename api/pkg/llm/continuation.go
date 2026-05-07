package llm

import (
	"regexp"
	"strings"
)

// planTailRe matches trailing sentences that indicate the LLM described its
// next action but stopped before emitting the tool call.
var planTailRe = regexp.MustCompile(
	`(?i)(I'll\s+(now\s+)?(deploy|create|set up|configure|proceed|call|use|send|update|check|try|run|invoke|start|execute|add|remove|delete)` +
		`|Let me\s+(now\s+)?(deploy|create|set up|configure|proceed|call|use|send|update|check|try|run|invoke|start|execute|add|remove|delete)` +
		`|Now I'll|I will now|I'm going to|Proceeding to|Let's go ahead)`,
)

// looksLikeUnfinishedPlan returns true if the assistant response ends with
// text that describes a next action without having executed it. This heuristic
// inspects the last ~300 characters looking for "I'll now deploy...", "Let me
// create...", and similar patterns that indicate the model stopped before making
// the intended tool call.
func looksLikeUnfinishedPlan(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}

	// Only inspect the tail of the response
	tail := content
	if len(tail) > 300 {
		tail = tail[len(tail)-300:]
	}

	return planTailRe.MatchString(tail)
}
