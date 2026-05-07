package llm

// ToolProfile defines a named set of tool capabilities.
type ToolProfile string

const (
	ProfileCore       ToolProfile = "core"
	ProfileDeploy     ToolProfile = "deploy"
	ProfileDiagnostic ToolProfile = "diagnostic"
	ProfileGitHub     ToolProfile = "github"
	ProfileNotify     ToolProfile = "notify"
	ProfileMachine    ToolProfile = "machine"
)

// ToolProfileBuilder assembles a ToolRegistry from a base profile plus addons.
type ToolProfileBuilder struct {
	coreFn func(*ToolRegistry)
	addons map[ToolProfile]func(*ToolRegistry)
}

// NewToolProfileBuilder creates a builder with the given core tool registrar.
// The core function is called for every profile.
func NewToolProfileBuilder(coreFn func(*ToolRegistry)) *ToolProfileBuilder {
	return &ToolProfileBuilder{
		coreFn: coreFn,
		addons: make(map[ToolProfile]func(*ToolRegistry)),
	}
}

// RegisterProfile registers additional tools for a specific profile.
func (b *ToolProfileBuilder) RegisterProfile(profile ToolProfile, fn func(*ToolRegistry)) {
	b.addons[profile] = fn
}

// Build creates a ToolRegistry for the given profile by running the core
// registrar followed by the profile-specific addon (if any).
func (b *ToolProfileBuilder) Build(profile ToolProfile) *ToolRegistry {
	reg := NewToolRegistry()
	if b.coreFn != nil {
		b.coreFn(reg)
	}
	if fn, ok := b.addons[profile]; ok {
		fn(reg)
	}
	return reg
}
