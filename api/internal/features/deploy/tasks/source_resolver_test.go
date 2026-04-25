package tasks

import (
	"fmt"
	"testing"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func TestGetSourceResolver_PublicGit(t *testing.T) {
	svc := &TaskService{}

	resolver := svc.GetSourceResolver(shared_types.SourcePublicGit)

	if resolver == nil {
		t.Fatal("expected non-nil resolver for SourcePublicGit")
	}
	if _, ok := resolver.(*PublicGitSourceResolver); !ok {
		t.Errorf("expected *PublicGitSourceResolver, got %T", resolver)
	}
}

func TestGetSourceResolver_DefaultIsGithub(t *testing.T) {
	svc := &TaskService{}

	resolver := svc.GetSourceResolver(shared_types.SourceGithub)

	if resolver == nil {
		t.Fatal("expected non-nil resolver for SourceGithub")
	}
	if _, ok := resolver.(*GithubSourceResolver); !ok {
		t.Errorf("expected *GithubSourceResolver, got %T", resolver)
	}
}

func TestGetSourceResolver_AllSourceTypes(t *testing.T) {
	svc := &TaskService{}

	tests := []struct {
		source       shared_types.Source
		expectedType string
	}{
		{shared_types.SourceS3, "*tasks.S3SourceResolver"},
		{shared_types.SourceZip, "*tasks.ZipSourceResolver"},
		{shared_types.SourceStaging, "*tasks.StagingSourceResolver"},
		{shared_types.SourceTemplate, "*tasks.TemplateSourceResolver"},
		{shared_types.SourcePublicGit, "*tasks.PublicGitSourceResolver"},
		{shared_types.SourceGithub, "*tasks.GithubSourceResolver"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			resolver := svc.GetSourceResolver(tt.source)
			if resolver == nil {
				t.Fatalf("expected non-nil resolver for %s", tt.source)
			}
			actualType := fmt.Sprintf("%T", resolver)
			if actualType != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, actualType)
			}
		})
	}
}
