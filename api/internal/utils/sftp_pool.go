package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	defaultSFTPIdleTimeout = 5 * time.Minute
)

type pooledSFTP struct {
	client   *sftp.Client
	lastUsed time.Time
}

// SFTPPool provides org-scoped SFTP client reuse to avoid connection churn.
// Clients are cached per organization and evicted on idle timeout or connection errors.
type SFTPPool struct {
	mu          sync.RWMutex
	clients     map[string]*pooledSFTP
	idleTimeout time.Duration
}

var globalSFTPPool = &SFTPPool{
	clients:     make(map[string]*pooledSFTP),
	idleTimeout: defaultSFTPIdleTimeout,
}

// WithSFTPClientFromPool runs fn with an SFTP client from the org-scoped pool.
// Context must have types.OrganizationIDKey set. Falls back to local staging (no SFTP) is not applicable.
// Evicts stale clients on connection errors; creates new client when pool empty or evicted.
func WithSFTPClientFromPool(ctx context.Context, fn func(*sftp.Client) error) error {
	orgIDAny := ctx.Value(types.OrganizationIDKey)
	if orgIDAny == nil {
		return fmt.Errorf("organization ID required for SFTP pool")
	}
	var orgID string
	switch v := orgIDAny.(type) {
	case string:
		orgID = v
	case uuid.UUID:
		orgID = v.String()
	default:
		return fmt.Errorf("invalid organization ID type: %T", orgIDAny)
	}
	if orgID == "" {
		return fmt.Errorf("empty organization ID")
	}

	sshMgr, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		return err
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		client, fromPool := globalSFTPPool.getOrCreate(ctx, orgID, sshMgr)
		if client == nil {
			return fmt.Errorf("failed to get SFTP client for org %s", orgID)
		}
		err = fn(client)
		if err != nil {
			if isClosedConnectionError(err) {
				globalSFTPPool.evict(orgID, client)
				if fromPool {
					sshMgr.CloseConnection("")
				}
				if attempt < maxRetries-1 {
					continue
				}
			}
			return err
		}
		globalSFTPPool.touch(orgID)
		return nil
	}
	return fmt.Errorf("SFTP operation failed after %d attempts", maxRetries)
}

func (p *SFTPPool) getOrCreate(ctx context.Context, orgID string, sshMgr *ssh.SSHManager) (*sftp.Client, bool) {
	p.mu.Lock()
	if entry, ok := p.clients[orgID]; ok {
		if time.Since(entry.lastUsed) <= p.idleTimeout {
			client := entry.client
			p.mu.Unlock()
			return client, true
		}
		entry.client.Close()
		delete(p.clients, orgID)
	}
	p.mu.Unlock()

	// Create new client outside lock to avoid blocking other orgs
	sshClient, err := sshMgr.Connect()
	if err != nil {
		return nil, false
	}
	sftpClient, err := sshClient.NewSftp()
	if err != nil {
		if isClosedConnectionError(err) {
			sshMgr.CloseConnection("")
		}
		return nil, false
	}

	p.mu.Lock()
	if existing, ok := p.clients[orgID]; ok {
		// Another goroutine added one while we were creating
		p.mu.Unlock()
		sftpClient.Close()
		return existing.client, true
	}
	p.clients[orgID] = &pooledSFTP{client: sftpClient, lastUsed: time.Now()}
	p.mu.Unlock()
	return sftpClient, false
}

func (p *SFTPPool) evict(orgID string, c *sftp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.clients[orgID]; ok && entry.client == c {
		entry.client.Close()
		delete(p.clients, orgID)
	}
}

func (p *SFTPPool) touch(orgID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.clients[orgID]; ok {
		entry.lastUsed = time.Now()
	}
}
