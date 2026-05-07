package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

type SearchableToolEntry struct {
	Definition ToolDefinition
	Keywords   []string
}

type ToolSearchConfig struct {
	TopK     int
	MinScore float64
}

type ToolSearch struct {
	entries []SearchableToolEntry
	config  ToolSearchConfig
}

func NewToolSearch(config ToolSearchConfig) *ToolSearch {
	if config.TopK == 0 {
		config.TopK = 6
	}
	if config.MinScore == 0 {
		config.MinScore = 0.1
	}
	return &ToolSearch{config: config}
}

func (ts *ToolSearch) Add(entry SearchableToolEntry) {
	if len(entry.Keywords) == 0 {
		entry.Keywords = extractKeywords(entry.Definition.Name + " " + entry.Definition.Description)
	}
	ts.entries = append(ts.entries, entry)
}

func (ts *ToolSearch) AddTool(def ToolDefinition, keywords ...string) {
	ts.Add(SearchableToolEntry{Definition: def, Keywords: keywords})
}

func (ts *ToolSearch) Search(query string) []ToolDefinition {
	queryTerms := extractKeywords(query)
	if len(queryTerms) == 0 {
		return nil
	}

	type scored struct {
		def   ToolDefinition
		score float64
	}

	var results []scored
	for _, entry := range ts.entries {
		score := bm25Score(queryTerms, entry.Keywords)
		if score >= ts.config.MinScore {
			results = append(results, scored{def: entry.Definition, score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	limit := ts.config.TopK
	if limit > len(results) {
		limit = len(results)
	}

	tools := make([]ToolDefinition, limit)
	for i := 0; i < limit; i++ {
		tools[i] = results[i].def
	}
	return tools
}

func (ts *ToolSearch) Count() int {
	return len(ts.entries)
}

// Tool returns a "load_tool" ToolDefinition that lets the agent search and load tools.
func (ts *ToolSearch) Tool(registry *ToolRegistry) ToolDefinition {
	return ToolDefinition{
		Name:        "load_tool",
		Description: "Search for and load specialized tools by describing what you need. Returns matching tools that become available for use.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"Describe the tool capability you need"}},"required":["query"]}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			results := ts.Search(input.Query)
			if len(results) == 0 {
				return json.Marshal(map[string]string{"message": "no matching tools found"})
			}

			loaded := make([]string, 0, len(results))
			for _, def := range results {
				if _, exists := registry.Get(def.Name); !exists {
					registry.Register(def)
				}
				loaded = append(loaded, def.Name+" — "+def.Description)
			}

			return json.Marshal(map[string]interface{}{
				"loaded": loaded,
				"count":  len(loaded),
			})
		},
	}
}

// Processor returns an InputProcessor that injects relevant tools based on the latest user message.
func (ts *ToolSearch) Processor() InputProcessor {
	return &toolSearchProcessor{ts: ts}
}

type toolSearchProcessor struct {
	ts *ToolSearch
}

func (p *toolSearchProcessor) ProcessInput(ctx context.Context, input *StepInput) error {
	if len(input.Messages) == 0 {
		return nil
	}

	// Find the last user message for search context
	var query string
	for i := len(input.Messages) - 1; i >= 0; i-- {
		if input.Messages[i].Role == RoleUser {
			query = input.Messages[i].Content
			break
		}
	}
	if query == "" {
		return nil
	}

	results := p.ts.Search(query)
	for _, def := range results {
		input.Tools = append(input.Tools, def.ToLLMTool())
	}
	return nil
}

// --- BM25 scoring ---

func bm25Score(queryTerms, docTerms []string) float64 {
	if len(docTerms) == 0 {
		return 0
	}

	termFreq := make(map[string]int)
	for _, t := range docTerms {
		termFreq[t]++
	}

	dl := float64(len(docTerms))
	avgDl := 10.0 // assumed average document length
	k1 := 1.5
	b := 0.75

	var score float64
	for _, qt := range queryTerms {
		tf := float64(termFreq[qt])
		if tf == 0 {
			continue
		}
		idf := math.Log(2.0) // simplified: all terms equally important
		numerator := tf * (k1 + 1)
		denominator := tf + k1*(1-b+b*(dl/avgDl))
		score += idf * (numerator / denominator)
	}
	return score
}

func extractKeywords(text string) []string {
	text = strings.ToLower(text)
	// Split on non-alpha chars
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	})

	// Filter stop words and short words
	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) > 2 && !isStopWord(w) {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

func isStopWord(w string) bool {
	switch w {
	case "the", "and", "for", "are", "but", "not", "you", "all",
		"can", "had", "her", "was", "one", "our", "out", "has",
		"have", "been", "this", "that", "with", "from", "they":
		return true
	}
	return false
}
