package tasks

import (
	"testing"

	"github.com/nixopus/nixopus/api/internal/config"
	s3store "github.com/nixopus/nixopus/api/internal/features/deploy/s3"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func TestS3ArtifactUploadSettingsGateDefault(t *testing.T) {
	defaults := shared_types.DefaultOrganizationSettingsData()
	if defaults.S3ArtifactUploadEnabled == nil {
		t.Fatal("S3ArtifactUploadEnabled should not be nil in defaults")
	}
	if *defaults.S3ArtifactUploadEnabled != false {
		t.Errorf("S3ArtifactUploadEnabled should default to false, got %v", *defaults.S3ArtifactUploadEnabled)
	}
}

func TestS3ArtifactUploadGateLogic(t *testing.T) {
	testCases := []struct {
		name     string
		setting  *bool
		expected bool
	}{
		{
			name:     "nil setting should block upload",
			setting:  nil,
			expected: false,
		},
		{
			name:     "false setting should block upload",
			setting:  boolPtr(false),
			expected: false,
		},
		{
			name:     "true setting should allow upload",
			setting:  boolPtr(true),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			settings := shared_types.OrganizationSettingsData{
				S3ArtifactUploadEnabled: tc.setting,
			}
			allowed := isS3ArtifactUploadAllowed(settings)
			if allowed != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, allowed)
			}
		})
	}
}

func TestS3NotConfiguredBlocksExport(t *testing.T) {
	emptyCfg := shared_types.S3Config{}
	if s3store.IsConfigured(emptyCfg) {
		t.Error("empty S3 config should not be considered configured")
	}

	fullCfg := shared_types.S3Config{
		Bucket:    "test-bucket",
		Region:    "us-east-1",
		AccessKey: "access",
		SecretKey: "secret",
		Endpoint:  "s3.example.com",
	}
	if !s3store.IsConfigured(fullCfg) {
		t.Error("complete S3 config should be considered configured")
	}
}

func TestS3ConfigAndSettingsGateCombination(t *testing.T) {
	savedConfig := config.AppConfig.S3
	defer func() { config.AppConfig.S3 = savedConfig }()

	t.Run("S3 not configured - gate irrelevant", func(t *testing.T) {
		config.AppConfig.S3 = shared_types.S3Config{}
		if s3store.IsConfigured(config.AppConfig.S3) {
			t.Error("should not be configured")
		}
	})

	t.Run("S3 configured but setting disabled", func(t *testing.T) {
		config.AppConfig.S3 = shared_types.S3Config{
			Bucket:    "test",
			Region:    "us-east-1",
			AccessKey: "key",
			SecretKey: "secret",
			Endpoint:  "s3.example.com",
		}
		if !s3store.IsConfigured(config.AppConfig.S3) {
			t.Fatal("should be configured")
		}
		settings := shared_types.OrganizationSettingsData{
			S3ArtifactUploadEnabled: boolPtr(false),
		}
		if isS3ArtifactUploadAllowed(settings) {
			t.Error("upload should not be allowed when setting is false")
		}
	})

	t.Run("S3 configured and setting enabled", func(t *testing.T) {
		config.AppConfig.S3 = shared_types.S3Config{
			Bucket:    "test",
			Region:    "us-east-1",
			AccessKey: "key",
			SecretKey: "secret",
			Endpoint:  "s3.example.com",
		}
		settings := shared_types.OrganizationSettingsData{
			S3ArtifactUploadEnabled: boolPtr(true),
		}
		if !isS3ArtifactUploadAllowed(settings) {
			t.Error("upload should be allowed when S3 configured and setting is true")
		}
	})
}

func boolPtr(b bool) *bool {
	return &b
}

// isS3ArtifactUploadAllowed mirrors the gate logic in ExportAndRecordImage/ExportComposeImagesToS3.
func isS3ArtifactUploadAllowed(settings shared_types.OrganizationSettingsData) bool {
	if settings.S3ArtifactUploadEnabled == nil || !*settings.S3ArtifactUploadEnabled {
		return false
	}
	return true
}
