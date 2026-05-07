package deploy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
)

var ecosystemPatterns = map[string]*regexp.Regexp{
	"nextjs":  regexp.MustCompile(`(?i)(next\.js|nextjs|next\.config)`),
	"nuxt":    regexp.MustCompile(`(?i)(nuxt|nuxt\.config)`),
	"react":   regexp.MustCompile(`(?i)(react|create-react-app|react-scripts)`),
	"vite":    regexp.MustCompile(`(?i)(vite|vite\.config)`),
	"django":  regexp.MustCompile(`(?i)(django|manage\.py|wsgi)`),
	"flask":   regexp.MustCompile(`(?i)(flask|gunicorn.*flask)`),
	"fastapi": regexp.MustCompile(`(?i)(fastapi|uvicorn)`),
	"rails":   regexp.MustCompile(`(?i)(rails|ruby on rails|Gemfile)`),
	"laravel": regexp.MustCompile(`(?i)(laravel|artisan|composer\.json.*laravel)`),
	"express": regexp.MustCompile(`(?i)(express|express\.js)`),
	"nestjs":  regexp.MustCompile(`(?i)(nestjs|@nestjs)`),
	"go":      regexp.MustCompile(`(?i)(go\.mod|golang|gin-gonic|fiber)`),
	"rust":    regexp.MustCompile(`(?i)(cargo\.toml|rust|actix|axum)`),
	"svelte":  regexp.MustCompile(`(?i)(svelte|sveltekit|svelte\.config)`),
	"astro":   regexp.MustCompile(`(?i)(astro|astro\.config)`),
}

// DetectEcosystem scans messages for ecosystem indicators.
func DetectEcosystem(messages []memory.StoredMessage, currentInput string) string {
	text := currentInput
	for _, m := range messages {
		text += " " + m.Content
	}

	for eco, pattern := range ecosystemPatterns {
		if pattern.MatchString(text) {
			return eco
		}
	}
	return ""
}

// DetectEcosystemFromLLM works with llm.Message slice.
func DetectEcosystemFromLLM(messages []llm.Message, currentInput string) string {
	text := currentInput
	for _, m := range messages {
		text += " " + m.Content
	}
	for eco, pattern := range ecosystemPatterns {
		if pattern.MatchString(text) {
			return eco
		}
	}
	return ""
}

// FormatPatterns builds the [deploy-patterns] block for injection.
func FormatPatterns(ecosystem string, patterns []DeployPattern) string {
	if len(patterns) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[deploy-patterns] ecosystem:%s\n", ecosystem))

	fixes := filterByType(patterns, "failure_fix")
	pitfalls := filterByType(patterns, "pitfall")
	fastPaths := filterByType(patterns, "fast_path")

	if len(fixes) > 0 {
		sb.WriteString("known_fixes:\n")
		for _, p := range fixes {
			sb.WriteString(fmt.Sprintf("- \"%s\" → %s (confidence:%d%%, seen:%d)\n",
				p.Signature, p.Resolution, int(p.Confidence*100), p.HitCount+p.MissCount))
		}
	}
	if len(pitfalls) > 0 {
		sb.WriteString("pitfalls:\n")
		for _, p := range pitfalls {
			sb.WriteString(fmt.Sprintf("- %s → %s (confidence:%d%%)\n",
				p.Signature, p.Resolution, int(p.Confidence*100)))
		}
	}
	if len(fastPaths) > 0 {
		sb.WriteString("fast_paths:\n")
		for _, p := range fastPaths {
			sb.WriteString(fmt.Sprintf("- %s → %s (confidence:%d%%)\n",
				p.Signature, p.Resolution, int(p.Confidence*100)))
		}
	}

	sb.WriteString("[/deploy-patterns]")
	return sb.String()
}

// ClassifyOutcome determines the deploy outcome from status and messages.
func ClassifyOutcome(lastStatus string, messages []memory.StoredMessage) string {
	switch lastStatus {
	case "running", "active", "healthy":
		return "success"
	case "build_failed", "failed", "error", "crashed":
		return "failed"
	}

	for i := len(messages) - 1; i >= 0 && i >= len(messages)-5; i-- {
		content := strings.ToLower(messages[i].Content)
		if strings.Contains(content, "deployed successfully") ||
			strings.Contains(content, "is now live") ||
			strings.Contains(content, "running at") {
			return "success"
		}
		if strings.Contains(content, "failed") ||
			strings.Contains(content, "error") ||
			strings.Contains(content, "could not deploy") {
			return "failed"
		}
		if strings.Contains(content, "rollback") ||
			strings.Contains(content, "rolled back") {
			return "rollback"
		}
	}
	return ""
}

func filterByType(patterns []DeployPattern, patternType string) []DeployPattern {
	var result []DeployPattern
	for _, p := range patterns {
		if p.PatternType == patternType {
			result = append(result, p)
		}
	}
	return result
}
