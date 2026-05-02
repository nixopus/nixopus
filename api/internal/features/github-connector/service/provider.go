package service

import (
	"context"

	gc_types "github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// GitConnectorProvider is the provider-agnostic contract for a git hosting connector.
//
// The current implementation is GithubConnectorService (GitHub App).
// Future implementations can target GitLab, Bitbucket, Gitea, etc. by satisfying
// this interface without touching any caller code.
type GitConnectorProvider interface {
	// Connector lifecycle
	CreateConnector(req *gc_types.CreateGithubConnectorRequest, userID string) error
	UpdateConnectorInstallation(installationID, userID, connectorID string) error
	DeleteConnector(connectorID, userID string) error
	GetConnector(connectorID string) (*shared_types.GithubConnector, error)
	GetAllConnectors(userID string) ([]shared_types.GithubConnector, error)

	// Remote repository operations
	GetRepositoriesPaginated(userID string, page, pageSize int, connectorID, search, sortBy, sortDirection string) ([]shared_types.GithubRepository, int, error)
	GetRepositoryBranches(userID, repositoryName string) ([]shared_types.GithubRepositoryBranch, error)
	GetRepositoryByID(userID string, repoID uint64) (*shared_types.GithubRepository, error)
	GetRepositoryFileContent(userID, repository, branch, filePath string) ([]byte, error)

	// Local git operations (SSH-side)
	CloneRepository(ctx context.Context, cfg CloneRepositoryConfig, commitHash *string) (string, error)
	CreateAuthenticatedRepoURL(repoURL, accessToken string) (string, error)
	GetClonePath(ctx context.Context, userID, environment, applicationID string) (string, bool, error)
	RemoveRepository(ctx context.Context, repoPath string) error
}

// Compile-time assertion: GithubConnectorService must satisfy GitConnectorProvider.
var _ GitConnectorProvider = (*GithubConnectorService)(nil)
