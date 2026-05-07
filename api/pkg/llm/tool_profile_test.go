package llm

import (
	"context"
	"encoding/json"
	"testing"
)

func TestToolProfileBuilder_CoreAlwaysPresent(t *testing.T) {
	pb := NewToolProfileBuilder(func(reg *ToolRegistry) {
		reg.Register(ToolDefinition{
			Name:       "core_tool",
			Parameters: json.RawMessage(`{}`),
			Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
				return nil, nil
			},
		})
	})

	reg := pb.Build(ProfileDeploy)
	if _, ok := reg.Get("core_tool"); !ok {
		t.Error("core tool should be present in deploy profile")
	}

	reg2 := pb.Build(ProfileDiagnostic)
	if _, ok := reg2.Get("core_tool"); !ok {
		t.Error("core tool should be present in diagnostic profile")
	}
}

func TestToolProfileBuilder_AddonApplied(t *testing.T) {
	pb := NewToolProfileBuilder(func(reg *ToolRegistry) {
		reg.Register(ToolDefinition{
			Name:       "core_tool",
			Parameters: json.RawMessage(`{}`),
			Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
				return nil, nil
			},
		})
	})

	pb.RegisterProfile(ProfileDeploy, func(reg *ToolRegistry) {
		reg.Register(ToolDefinition{
			Name:       "deploy_only",
			Parameters: json.RawMessage(`{}`),
			Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
				return nil, nil
			},
		})
	})

	deployReg := pb.Build(ProfileDeploy)
	if _, ok := deployReg.Get("deploy_only"); !ok {
		t.Error("deploy addon tool should be present in deploy profile")
	}
	if _, ok := deployReg.Get("core_tool"); !ok {
		t.Error("core tool should also be present in deploy profile")
	}

	diagReg := pb.Build(ProfileDiagnostic)
	if _, ok := diagReg.Get("deploy_only"); ok {
		t.Error("deploy addon tool should NOT be present in diagnostic profile")
	}
}

func TestToolProfileBuilder_NilCoreFn(t *testing.T) {
	pb := NewToolProfileBuilder(nil)
	reg := pb.Build(ProfileCore)
	if reg == nil {
		t.Error("should return non-nil registry even with nil core function")
	}
}
