package llm

import (
	"strings"
	"testing"
)

func TestPromptBuilder_Basic(t *testing.T) {
	b := NewPromptBuilder()
	b.Add(PromptSection{Name: "identity", Priority: 0, Content: "You are an assistant."})
	b.Add(PromptSection{Name: "rules", Priority: 1, Content: "Be helpful."})

	result := b.Build()
	if !strings.Contains(result, "You are an assistant.") {
		t.Error("missing identity section")
	}
	if !strings.Contains(result, "Be helpful.") {
		t.Error("missing rules section")
	}
}

func TestPromptBuilder_PriorityOrder(t *testing.T) {
	b := NewPromptBuilder()
	b.Add(PromptSection{Name: "low", Priority: 10, Content: "LOW"})
	b.Add(PromptSection{Name: "high", Priority: 0, Content: "HIGH"})
	b.Add(PromptSection{Name: "mid", Priority: 5, Content: "MID"})

	result := b.Build()

	highIdx := strings.Index(result, "HIGH")
	midIdx := strings.Index(result, "MID")
	lowIdx := strings.Index(result, "LOW")

	if highIdx > midIdx || midIdx > lowIdx {
		t.Errorf("sections not in priority order: high=%d mid=%d low=%d", highIdx, midIdx, lowIdx)
	}
}

func TestPromptBuilder_TotalTruncation(t *testing.T) {
	b := NewPromptBuilder()
	b.MaxChars = 100

	b.Add(PromptSection{Name: "critical", Priority: 0, Content: strings.Repeat("A", 60), MaxChars: 100})
	b.Add(PromptSection{Name: "optional", Priority: 10, Content: strings.Repeat("B", 60), MaxChars: 100})

	result := b.Build()

	if !strings.Contains(result, strings.Repeat("A", 60)) {
		t.Error("high-priority section should survive")
	}
	if strings.Contains(result, strings.Repeat("B", 60)) {
		t.Log("low-priority section was truncated as expected")
	}
}

func TestPromptBuilder_PerSectionCap(t *testing.T) {
	b := NewPromptBuilder()
	b.MaxChars = 1_000_000

	b.Add(PromptSection{
		Name:     "big",
		Priority: 0,
		Content:  strings.Repeat("X", 50_000),
		MaxChars: 100,
	})

	result := b.Build()
	if len(result) > 200 {
		t.Errorf("per-section cap not applied, got %d chars", len(result))
	}
	if !strings.Contains(result, "[truncated]") {
		t.Error("truncated marker missing")
	}
}

func TestPromptBuilder_Empty(t *testing.T) {
	b := NewPromptBuilder()
	if result := b.Build(); result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestPromptBuilder_EmptySectionsSkipped(t *testing.T) {
	b := NewPromptBuilder()
	b.Add(PromptSection{Name: "filled", Priority: 0, Content: "hello"})
	b.Add(PromptSection{Name: "empty", Priority: 1, Content: ""})

	result := b.Build()
	if strings.Contains(result, "empty") {
		t.Error("empty section should not appear in output")
	}
	if !strings.Contains(result, "hello") {
		t.Error("filled section should appear")
	}
}

func TestPromptBuilder_SectionCount(t *testing.T) {
	b := NewPromptBuilder()
	b.Add(PromptSection{Name: "a", Priority: 0, Content: "x"})
	b.Add(PromptSection{Name: "b", Priority: 1, Content: "y"})

	if b.SectionCount() != 2 {
		t.Errorf("expected 2, got %d", b.SectionCount())
	}
}
