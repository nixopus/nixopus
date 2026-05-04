package service

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/domain/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type DomainsService struct {
	storage storage.DomainStorageInterface
	Ctx     context.Context
	logger  logger.Logger
	dns     DNSResolver
	queue   QueueClient
}

// NewDomainsService creates a service wired to the real DNS and queue backends.
func NewDomainsService(ctx context.Context, l logger.Logger, repo storage.DomainStorageInterface) *DomainsService {
	return NewDomainsServiceWith(ctx, l, repo, NewRealDNSResolver(), &RealQueueClient{})
}

// NewDomainsServiceWith creates a service with explicit DNS and queue implementations.
// Use this in tests to inject mocks.
func NewDomainsServiceWith(
	ctx context.Context,
	l logger.Logger,
	repo storage.DomainStorageInterface,
	dns DNSResolver,
	q QueueClient,
) *DomainsService {
	return &DomainsService{
		storage: repo,
		Ctx:     ctx,
		logger:  l,
		dns:     dns,
		queue:   q,
	}
}
