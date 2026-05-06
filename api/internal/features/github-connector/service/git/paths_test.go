package git

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/require"
)

func TestResolveClonePath_SFTPPoolError(t *testing.T) {
	t.Parallel()

	orgID := uuid.New().String()
	pool := utils.NewSFTPPool(5*time.Minute, func(string, *ssh.SSHManager) (*sftp.Client, error) {
		return nil, context.DeadlineExceeded
	})
	sshMgr := ssh.NewSSHManager()
	defer sshMgr.Close()

	ctx := utils.WithTestSFTPPool(context.Background(), orgID, pool, sshMgr)

	_, _, err := ResolveClonePath(ctx, logger.NewLogger(), "user1", "production", "app1")
	require.Error(t, err)
}
