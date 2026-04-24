package utils

import (
	"errors"
	"io"
	"testing"

	"github.com/melbahja/goph"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSFTPClientWithRetry_connectFails(t *testing.T) {
	mgr := ssh.NewSSHManager()
	_, err := CreateSFTPClientWithRetry(mgr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect")
}

func TestCreateSFTPClientWithRetry_success(t *testing.T) {
	old := newSftpFromGophClient
	t.Cleanup(func() { newSftpFromGophClient = old })

	mgr := ssh.NewSSHManagerForTest(func(id string) (*goph.Client, error) {
		return &goph.Client{}, nil
	}, 0)
	newSftpFromGophClient = func(_ *goph.Client) (*sftp.Client, error) {
		return newInMemSFTPClient(t), nil
	}
	c, err := CreateSFTPClientWithRetry(mgr)
	require.NoError(t, err)
	require.NotNil(t, c)
	t.Cleanup(func() { _ = c.Close() })
}

func TestCreateSFTPClientWithRetry_recoversAfterClosedSftp(t *testing.T) {
	old := newSftpFromGophClient
	t.Cleanup(func() { newSftpFromGophClient = old })

	mgr := ssh.NewSSHManagerForTest(func(id string) (*goph.Client, error) {
		return &goph.Client{}, nil
	}, 0)
	var n int
	newSftpFromGophClient = func(_ *goph.Client) (*sftp.Client, error) {
		n++
		if n < 2 {
			return nil, io.EOF
		}
		return newInMemSFTPClient(t), nil
	}
	c, err := CreateSFTPClientWithRetry(mgr)
	require.NoError(t, err)
	require.NotNil(t, c)
	t.Cleanup(func() { _ = c.Close() })
	assert.GreaterOrEqual(t, n, 2)
}

func TestCreateSFTPClientWithRetry_givesUpAfterClosedErrors(t *testing.T) {
	old := newSftpFromGophClient
	t.Cleanup(func() { newSftpFromGophClient = old })

	mgr := ssh.NewSSHManagerForTest(func(id string) (*goph.Client, error) {
		return &goph.Client{}, nil
	}, 0)
	newSftpFromGophClient = func(_ *goph.Client) (*sftp.Client, error) {
		return nil, io.EOF
	}
	_, err := CreateSFTPClientWithRetry(mgr)
	require.Error(t, err)
}

func TestCreateSFTPClientWithRetry_nonRetriableNewSftp(t *testing.T) {
	old := newSftpFromGophClient
	t.Cleanup(func() { newSftpFromGophClient = old })

	mgr := ssh.NewSSHManagerForTest(func(id string) (*goph.Client, error) {
		return &goph.Client{}, nil
	}, 0)
	newSftpFromGophClient = func(_ *goph.Client) (*sftp.Client, error) {
		return nil, errors.New("sftp: permission denied")
	}
	_, err := CreateSFTPClientWithRetry(mgr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}
