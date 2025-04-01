package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	audit_storage "github.com/raghavyuva/nixopus-api/internal/features/audit/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type AuditService struct {
	storage *audit_storage.AuditStorage
	ctx     context.Context
	logger  logger.Logger
}

func NewAuditService(db *bun.DB, ctx context.Context, logger logger.Logger) *AuditService {
	return &AuditService{
		storage: audit_storage.NewAuditStorage(db, ctx),
		ctx:     ctx,
		logger:  logger,
	}
}

type AuditLogRequest struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Action         types.AuditAction
	ResourceType   types.AuditResourceType
	ResourceID     uuid.UUID
	OldValues      map[string]any
	NewValues      map[string]any
	Metadata       map[string]any
	IPAddress      string
	UserAgent      string
	RequestID      uuid.UUID
}

func (s *AuditService) LogAction(req *AuditLogRequest) error {
	auditLog := &types.AuditLog{
		ID:             uuid.New(),
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		Action:         req.Action,
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		OldValues:      req.OldValues,
		NewValues:      req.NewValues,
		Metadata:       req.Metadata,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		CreatedAt:      time.Now(),
		RequestID:      req.RequestID,
	}

	if err := s.storage.CreateAuditLog(auditLog); err != nil {
		s.logger.Log(logger.Error, "Failed to create audit log", err.Error())
		return err
	}

	return nil
}

func (s *AuditService) GetAuditLogs(filters map[string]interface{}, page, pageSize int) ([]*types.AuditLog, int, error) {
	return s.storage.GetAuditLogs(filters, page, pageSize)
}

func (s *AuditService) GetAuditLogsByResource(resourceType types.AuditResourceType, resourceID uuid.UUID, page, pageSize int) ([]*types.AuditLog, int, error) {
	filters := map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
	}
	return s.storage.GetAuditLogs(filters, page, pageSize)
}

func (s *AuditService) GetAuditLogsByUser(userID uuid.UUID, page, pageSize int) ([]*types.AuditLog, int, error) {
	filters := map[string]interface{}{
		"user_id": userID,
	}
	return s.storage.GetAuditLogs(filters, page, pageSize)
}

func (s *AuditService) GetAuditLogsByOrganization(orgID uuid.UUID, page, pageSize int) ([]*types.AuditLog, int, error) {
	filters := map[string]interface{}{
		"organization_id": orgID,
	}
	return s.storage.GetAuditLogs(filters, page, pageSize)
}
