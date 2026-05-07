package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"
)

func TestSkillStore_Register(t *testing.T) {
	store := NewSkillStore()
	err := store.Register(Skill{Name: "deploy", Description: "Deploy apps", Content: "# Deploy\nSteps..."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Count() != 1 {
		t.Errorf("expected 1 skill, got %d", store.Count())
	}
}

func TestSkillStore_RegisterEmptyName(t *testing.T) {
	store := NewSkillStore()
	err := store.Register(Skill{Name: "", Content: "stuff"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSkillStore_RegisterEmptyContent(t *testing.T) {
	store := NewSkillStore()
	err := store.Register(Skill{Name: "test", Content: ""})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestSkillStore_Get(t *testing.T) {
	store := NewSkillStore()
	store.Register(Skill{Name: "node-deploy", Content: "# Node Deploy"})

	skill, ok := store.Get("node-deploy")
	if !ok {
		t.Fatal("expected skill to be found")
	}
	if skill.Content != "# Node Deploy" {
		t.Errorf("unexpected content: %s", skill.Content)
	}

	_, ok = store.Get("nonexistent")
	if ok {
		t.Error("expected skill to not be found")
	}
}

func TestSkillStore_List(t *testing.T) {
	store := NewSkillStore()
	store.Register(Skill{Name: "a", Content: "A"})
	store.Register(Skill{Name: "b", Content: "B"})

	list := store.List()
	if len(list) != 2 {
		t.Errorf("expected 2 skills, got %d", len(list))
	}
}

func TestSkillStore_Catalog(t *testing.T) {
	store := NewSkillStore()
	store.Register(Skill{Name: "deploy", Description: "Deploy stuff", Content: "content"})
	store.Register(Skill{Name: "debug", Content: "content"})

	catalog := store.Catalog()
	if len(catalog) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(catalog))
	}

	hasWithDesc := false
	hasWithoutDesc := false
	for _, entry := range catalog {
		if entry == "deploy — Deploy stuff" {
			hasWithDesc = true
		}
		if entry == "debug" {
			hasWithoutDesc = true
		}
	}
	if !hasWithDesc {
		t.Error("expected 'deploy — Deploy stuff' in catalog")
	}
	if !hasWithoutDesc {
		t.Error("expected 'debug' (no description) in catalog")
	}
}

func TestSkillStore_LoadFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"node-deploy/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: node-deploy\ndescription: Deploy Node.js apps\nmetadata:\n  version: \"1.0\"\n---\n\n# Node Deploy\n\nSteps here."),
		},
		"go-deploy/SKILL.md": &fstest.MapFile{
			Data: []byte("# Go Deploy\n\nBuild and deploy Go apps."),
		},
		"other-file.txt": &fstest.MapFile{
			Data: []byte("ignored"),
		},
	}

	store := NewSkillStore()
	err := store.LoadFromFS(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.Count() != 2 {
		t.Errorf("expected 2 skills, got %d", store.Count())
	}

	skill, ok := store.Get("node-deploy")
	if !ok {
		t.Fatal("node-deploy not found")
	}
	if skill.Description != "Deploy Node.js apps" {
		t.Errorf("unexpected description: %s", skill.Description)
	}
	if skill.Content != "# Node Deploy\n\nSteps here." {
		t.Errorf("unexpected content: %q", skill.Content)
	}

	skill, ok = store.Get("go-deploy")
	if !ok {
		t.Fatal("go-deploy not found")
	}
	if skill.Content != "# Go Deploy\n\nBuild and deploy Go apps." {
		t.Errorf("unexpected content: %q", skill.Content)
	}
}

func TestSkillStore_LoadFromFS_Error(t *testing.T) {
	fsys := &failFS{}
	store := NewSkillStore()
	err := store.LoadFromFS(fsys)
	if err == nil {
		t.Fatal("expected error from failing FS")
	}
}

func TestSkillStore_LoadFromFS_ReadError(t *testing.T) {
	fsys := fstest.MapFS{
		"broken/SKILL.md": &fstest.MapFile{
			Data: []byte("content"),
			Mode: 0000,
		},
	}

	store := NewSkillStore()
	// fstest.MapFS doesn't enforce permissions, so this will succeed
	// Test with a custom FS that returns read errors
	err := store.LoadFromFS(fsys)
	if err != nil {
		t.Logf("got error (expected on some systems): %v", err)
	}
}

func TestSkillStore_Tool_Found(t *testing.T) {
	store := NewSkillStore()
	store.Register(Skill{Name: "deploy-flow", Content: "# Deploy Flow\nSteps..."})

	tool := store.Tool()
	if tool.Name != "read_skill" {
		t.Errorf("unexpected tool name: %s", tool.Name)
	}

	result, err := tool.Handler(context.Background(), json.RawMessage(`{"name":"deploy-flow"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]string
	json.Unmarshal(result, &output)
	if output["name"] != "deploy-flow" {
		t.Errorf("unexpected name in result: %s", output["name"])
	}
	if output["content"] != "# Deploy Flow\nSteps..." {
		t.Errorf("unexpected content in result: %s", output["content"])
	}
}

func TestSkillStore_Tool_NotFound(t *testing.T) {
	store := NewSkillStore()
	store.Register(Skill{Name: "a", Content: "A"})

	tool := store.Tool()
	result, err := tool.Handler(context.Background(), json.RawMessage(`{"name":"missing"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]interface{}
	json.Unmarshal(result, &output)
	if output["error"] == nil {
		t.Error("expected error field in result")
	}
	if output["available"] == nil {
		t.Error("expected available field in result")
	}
}

func TestSkillStore_Tool_InvalidArgs(t *testing.T) {
	store := NewSkillStore()
	tool := store.Tool()
	_, err := tool.Handler(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON args")
	}
}

func TestParseSkillContent_NoFrontmatter(t *testing.T) {
	skill := parseSkillContent("test", "# Just content\n\nNo frontmatter here.")
	if skill.Name != "test" {
		t.Errorf("expected name 'test', got %q", skill.Name)
	}
	if skill.Description != "" {
		t.Errorf("expected empty description, got %q", skill.Description)
	}
	if skill.Content != "# Just content\n\nNo frontmatter here." {
		t.Errorf("unexpected content: %q", skill.Content)
	}
}

func TestParseSkillContent_WithFrontmatter(t *testing.T) {
	raw := "---\nname: custom-name\ndescription: A custom skill\n---\n\n# Content\n\nBody here."
	skill := parseSkillContent("fallback", raw)
	if skill.Name != "custom-name" {
		t.Errorf("expected 'custom-name', got %q", skill.Name)
	}
	if skill.Description != "A custom skill" {
		t.Errorf("expected 'A custom skill', got %q", skill.Description)
	}
	if skill.Content != "# Content\n\nBody here." {
		t.Errorf("unexpected content: %q", skill.Content)
	}
}

func TestParseSkillContent_FrontmatterEmptyName(t *testing.T) {
	raw := "---\nname: \ndescription: Desc\n---\n\nBody"
	skill := parseSkillContent("fallback", raw)
	// Empty name in frontmatter should keep the fallback
	if skill.Name != "fallback" {
		t.Errorf("expected 'fallback', got %q", skill.Name)
	}
}

func TestParseSkillContent_IncompleteYAML(t *testing.T) {
	raw := "---\nname: partial\nno closing delimiter"
	skill := parseSkillContent("fallback", raw)
	// No closing ---, treat as no frontmatter
	if skill.Name != "fallback" {
		t.Errorf("expected 'fallback', got %q", skill.Name)
	}
	if skill.Content != raw {
		t.Errorf("expected raw content preserved")
	}
}

func TestSkillStore_LoadFromFS_RootSkillFile(t *testing.T) {
	fsys := fstest.MapFS{
		"SKILL.md": &fstest.MapFile{
			Data: []byte("# Root Skill\n\nA skill at root level."),
		},
	}

	store := NewSkillStore()
	err := store.LoadFromFS(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Name should be derived from filename without .md
	skill, ok := store.Get("SKILL")
	if !ok {
		t.Fatal("expected root skill to be loaded")
	}
	if skill.Content != "# Root Skill\n\nA skill at root level." {
		t.Errorf("unexpected content: %q", skill.Content)
	}
}

func TestSkillStore_LoadFromFS_ReadFileError(t *testing.T) {
	store := NewSkillStore()
	err := store.LoadFromFS(&readFailFS{})
	if err == nil {
		t.Fatal("expected error from read failure")
	}
}

type failFS struct{}

func (f *failFS) Open(name string) (fs.File, error) {
	return nil, fmt.Errorf("simulated FS error")
}

type readFailFS struct{}

func (f *readFailFS) Open(name string) (fs.File, error) {
	if name == "." {
		return &fakeDir{entries: []fs.DirEntry{&fakeDirEntry{name: "test", isDir: true}}}, nil
	}
	if name == "test" {
		return &fakeDir{entries: []fs.DirEntry{&fakeDirEntry{name: "SKILL.md", isDir: false}}}, nil
	}
	if name == "test/SKILL.md" {
		return nil, fmt.Errorf("permission denied")
	}
	return nil, fmt.Errorf("not found")
}

type fakeDir struct {
	entries []fs.DirEntry
	pos     int
}

func (d *fakeDir) Stat() (fs.FileInfo, error) { return &fakeDirInfo{}, nil }
func (d *fakeDir) Read([]byte) (int, error)   { return 0, fmt.Errorf("is directory") }
func (d *fakeDir) Close() error               { return nil }
func (d *fakeDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.pos >= len(d.entries) {
		return nil, nil
	}
	entries := d.entries[d.pos:]
	d.pos = len(d.entries)
	return entries, nil
}

type fakeDirInfo struct{}

func (i *fakeDirInfo) Name() string       { return "." }
func (i *fakeDirInfo) Size() int64        { return 0 }
func (i *fakeDirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0755 }
func (i *fakeDirInfo) IsDir() bool        { return true }
func (i *fakeDirInfo) Sys() interface{}   { return nil }
func (i *fakeDirInfo) ModTime() time.Time { return time.Time{} }

type fakeDirEntry struct {
	name  string
	isDir bool
}

func (e *fakeDirEntry) Name() string               { return e.name }
func (e *fakeDirEntry) IsDir() bool                { return e.isDir }
func (e *fakeDirEntry) Type() fs.FileMode          { return 0 }
func (e *fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }
