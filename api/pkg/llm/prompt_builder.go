package llm

import (
	"sort"
	"strings"
)

const (
	DefaultMaxChars   = 120_000 // ~30K tokens
	DefaultSectionCap = 25_000  // ~6K tokens per section
)

// PromptSection represents a named block within a system prompt.
// Lower Priority values survive truncation; higher values are cut first.
type PromptSection struct {
	Name     string
	Priority int
	Content  string
	MaxChars int // per-section cap; 0 = use DefaultSectionCap
}

// PromptBuilder assembles a system prompt from prioritised sections with
// per-section and total character budgets. When the total exceeds MaxChars,
// the lowest-priority (highest Priority number) sections are truncated first.
type PromptBuilder struct {
	sections []PromptSection
	MaxChars int // total budget; 0 = DefaultMaxChars
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{MaxChars: DefaultMaxChars}
}

// Add appends a section. Duplicate names are allowed (both are kept).
func (b *PromptBuilder) Add(section PromptSection) *PromptBuilder {
	b.sections = append(b.sections, section)
	return b
}

// Build assembles all sections into a single prompt string.
// Sections are emitted in Priority order (low number first).
// If total chars exceed b.MaxChars, lowest-priority sections are trimmed.
func (b *PromptBuilder) Build() string {
	if len(b.sections) == 0 {
		return ""
	}

	sorted := make([]PromptSection, len(b.sections))
	copy(sorted, b.sections)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	maxTotal := b.MaxChars
	if maxTotal <= 0 {
		maxTotal = DefaultMaxChars
	}

	// Apply per-section caps
	for i := range sorted {
		cap := sorted[i].MaxChars
		if cap <= 0 {
			cap = DefaultSectionCap
		}
		if len(sorted[i].Content) > cap {
			sorted[i].Content = sorted[i].Content[:cap] + "\n... [truncated]"
		}
	}

	total := 0
	for _, s := range sorted {
		total += len(s.Content)
	}

	// Trim from the back (lowest priority = highest index in sorted slice)
	for i := len(sorted) - 1; i >= 0 && total > maxTotal; i-- {
		excess := total - maxTotal
		if excess >= len(sorted[i].Content) {
			total -= len(sorted[i].Content)
			sorted[i].Content = ""
		} else {
			sorted[i].Content = sorted[i].Content[:len(sorted[i].Content)-excess] + "\n... [truncated]"
			total = maxTotal
		}
	}

	var sb strings.Builder
	for _, s := range sorted {
		if s.Content == "" {
			continue
		}
		sb.WriteString(s.Content)
		if !strings.HasSuffix(s.Content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// SectionCount returns the number of registered sections.
func (b *PromptBuilder) SectionCount() int {
	return len(b.sections)
}

// TotalChars returns the raw total characters across all sections before truncation.
func (b *PromptBuilder) TotalChars() int {
	total := 0
	for _, s := range b.sections {
		total += len(s.Content)
	}
	return total
}
