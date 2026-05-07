package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
)

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"-"`
}

type SkillStore struct {
	mu     sync.RWMutex
	skills map[string]Skill
}

func NewSkillStore() *SkillStore {
	return &SkillStore{skills: make(map[string]Skill)}
}

func (s *SkillStore) Register(skill Skill) error {
	if skill.Name == "" {
		return fmt.Errorf("llm: skill name cannot be empty")
	}
	if skill.Content == "" {
		return fmt.Errorf("llm: skill %q content cannot be empty", skill.Name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.skills[skill.Name] = skill
	return nil
}

func (s *SkillStore) Get(name string) (Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	skill, ok := s.skills[name]
	return skill, ok
}

func (s *SkillStore) List() []Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]Skill, 0, len(s.skills))
	for _, skill := range s.skills {
		list = append(list, skill)
	}
	return list
}

func (s *SkillStore) Catalog() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	catalog := make([]string, 0, len(s.skills))
	for _, skill := range s.skills {
		entry := skill.Name
		if skill.Description != "" {
			entry += " — " + skill.Description
		}
		catalog = append(catalog, entry)
	}
	return catalog
}

func (s *SkillStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.skills)
}

// LoadFromFS loads all SKILL.md files from an fs.FS.
// Expected structure: <name>/SKILL.md
func (s *SkillStore) LoadFromFS(fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "SKILL.md" {
			return nil
		}

		content, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("llm: read skill %s: %w", p, err)
		}

		name := path.Dir(p)
		if name == "." {
			name = strings.TrimSuffix(d.Name(), ".md")
		}

		skill := parseSkillContent(name, string(content))
		s.mu.Lock()
		s.skills[skill.Name] = skill
		s.mu.Unlock()

		return nil
	})
}

func parseSkillContent(name, content string) Skill {
	skill := Skill{Name: name, Content: content}

	// Parse YAML frontmatter if present
	if strings.HasPrefix(content, "---\n") {
		parts := strings.SplitN(content[4:], "\n---\n", 2)
		if len(parts) == 2 {
			frontmatter := parts[0]
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					val := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
					if val != "" {
						skill.Name = val
					}
				} else if strings.HasPrefix(line, "description:") {
					skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
			skill.Content = strings.TrimSpace(parts[1])
		}
	}

	return skill
}

// Tool returns a ToolDefinition that lets the agent read skills by name.
func (s *SkillStore) Tool() ToolDefinition {
	return ToolDefinition{
		Name:        "read_skill",
		Description: "Load task-specific instructions by skill name. Use when you need domain guidance for a specific task.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Skill name to load"}},"required":["name"]}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			skill, ok := s.Get(input.Name)
			if !ok {
				available := s.Catalog()
				return json.Marshal(map[string]interface{}{
					"error":     fmt.Sprintf("skill %q not found", input.Name),
					"available": available,
				})
			}

			return json.Marshal(map[string]string{
				"name":    skill.Name,
				"content": skill.Content,
			})
		},
	}
}
