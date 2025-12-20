package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

type DeployRepository interface {
	IsNameAlreadyTaken(name string) (bool, error)
	IsDomainAlreadyTaken(domain string) (bool, error)
	IsPortAlreadyTaken(port int) (bool, error)
	IsDomainValid(domain string) (bool, error)
	AddApplication(application *shared_types.Application) error
	AddApplicationLogs(applicationLogs *shared_types.ApplicationLogs) error
	AddApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error
	GetApplications(page int, pageSize int, organizationID uuid.UUID) ([]shared_types.Application, int, error)
	UpdateApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error
	GetApplicationById(id string, organizationID uuid.UUID) (shared_types.Application, error)
	AddApplicationDeployment(deployment *shared_types.ApplicationDeployment) error
	AddApplicationDeploymentStatus(deployment_status *shared_types.ApplicationDeploymentStatus) error
	UpdateApplicationDeploymentStatus(applicationStatus *shared_types.ApplicationDeploymentStatus) error
	UpdateApplication(application *shared_types.Application) error
	GetApplicationDeploymentById(deploymentID string) (shared_types.ApplicationDeployment, error)
	DeleteDeployment(deployment *types.DeleteDeploymentRequest, userID uuid.UUID) error
	UpdateApplicationDeployment(deployment *shared_types.ApplicationDeployment) error
	GetApplicationDeployments(applicationID uuid.UUID) ([]shared_types.ApplicationDeployment, error)
	GetPaginatedApplicationDeployments(applicationID uuid.UUID, page, pageSize int) ([]shared_types.ApplicationDeployment, int, error)
	GetLogs(applicationID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int, error)
	GetDeploymentLogs(deploymentID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int, error)
	GetApplicationByRepositoryID(repositoryID uint64) (shared_types.Application, error)
	GetApplicationByRepositoryIDAndBranch(repositoryID uint64, branch string) ([]shared_types.Application, error)
	UpdateApplicationLabels(applicationID uuid.UUID, labels []string, organizationID uuid.UUID) error
}

func (s *DeployStorage) IsNameAlreadyTaken(name string) (bool, error) {
	var count int
	err := s.DB.NewSelect().
		TableExpr("applications").
		ColumnExpr("count(*)").
		Where("name = ?", name).
		Scan(s.Ctx, &count)

	return count > 0, err
}

func (s *DeployStorage) IsDomainAlreadyTaken(domain string) (bool, error) {
	var count int
	err := s.DB.NewSelect().
		TableExpr("applications").
		ColumnExpr("count(*)").
		Where("domain = ?", domain).
		Scan(s.Ctx, &count)

	return count > 0, err
}

func (s *DeployStorage) IsPortAlreadyTaken(port int) (bool, error) {
	var count int
	err := s.DB.NewSelect().
		TableExpr("applications").
		ColumnExpr("count(*)").
		Where("port = ?", port).
		Scan(s.Ctx, &count)

	return count > 0, err
}

func (s *DeployStorage) IsDomainValid(domain string) (bool, error) {
	return true, nil
}

func (s *DeployStorage) AddApplication(application *shared_types.Application) error {
	_, err := s.DB.NewInsert().Model(application).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) UpdateApplication(application *shared_types.Application) error {
	_, err := s.DB.NewUpdate().
		Model(application).
		OmitZero().
		WherePK().
		Exec(s.Ctx)

	return err
}

func (s *DeployStorage) AddApplicationDeployment(deployment *shared_types.ApplicationDeployment) error {
	_, err := s.DB.NewInsert().Model(deployment).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) UpdateApplicationDeployment(deployment *shared_types.ApplicationDeployment) error {
	_, err := s.DB.NewUpdate().Model(deployment).Where("id = ?", deployment.ID).OmitZero().Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) AddApplicationDeploymentStatus(deployment_status *shared_types.ApplicationDeploymentStatus) error {
	_, err := s.DB.NewInsert().Model(deployment_status).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) UpdateApplicationDeploymentStatus(applicationStatus *shared_types.ApplicationDeploymentStatus) error {
	_, err := s.DB.NewUpdate().Model(applicationStatus).Where("id = ?", applicationStatus.ID).OmitZero().Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) AddApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error {
	_, err := s.DB.NewInsert().Model(applicationStatus).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) UpdateApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error {
	_, err := s.DB.NewUpdate().Model(applicationStatus).Where("id = ?", applicationStatus.ID).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) AddApplicationLogs(applicationLogs *shared_types.ApplicationLogs) error {
	_, err := s.DB.NewInsert().Model(applicationLogs).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *DeployStorage) GetApplications(page, pageSize int, organizationID uuid.UUID) ([]shared_types.Application, int, error) {
	var applications []shared_types.Application

	offset := (page - 1) * pageSize

	totalCount, err := s.DB.NewSelect().
		Model((*shared_types.Application)(nil)).
		Count(s.Ctx)

	if err != nil {
		return nil, 0, err
	}

	err = s.DB.NewSelect().
		Model(&applications).
		Relation("Status").
		Relation("Logs").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Where("organization_id = ?", organizationID).
		Scan(s.Ctx)

	if err != nil {
		return nil, 0, err
	}

	return applications, totalCount, nil
}

func (s *DeployStorage) GetApplicationById(id string, organizationID uuid.UUID) (shared_types.Application, error) {
	var application shared_types.Application

	err := s.DB.NewSelect().
		Model(&application).
		Relation("Status").
		Where("a.id = ? AND a.organization_id = ?", id, organizationID).
		Scan(s.Ctx)

	if err != nil {
		return shared_types.Application{}, err
	}

	return application, nil
}

func (s *DeployStorage) GetApplicationDeploymentById(deploymentID string) (shared_types.ApplicationDeployment, error) {
	var deployment shared_types.ApplicationDeployment

	err := s.DB.NewSelect().
		Model(&deployment).
		Relation("Status").
		Relation("Logs").
		Where("ad.id = ?", deploymentID).
		Scan(s.Ctx)

	if err != nil {
		return shared_types.ApplicationDeployment{}, err
	}

	return deployment, nil
}

func (s *DeployStorage) DeleteDeployment(deployment *types.DeleteDeploymentRequest, userID uuid.UUID) error {
	var count int
	err := s.DB.NewSelect().
		TableExpr("applications").
		ColumnExpr("count(*)").
		Where("id = ? AND user_id = ?", deployment.ID, userID).
		Scan(s.Ctx, &count)

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("application not found or not authorized")
	}

	_, err = s.DB.NewDelete().
		Table("application_logs").
		Where("application_deployment_id IN (SELECT id FROM application_deployment WHERE application_id = ?)", deployment.ID).
		Exec(s.Ctx)

	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Table("application_deployment_status").
		Where("application_deployment_id IN (SELECT id FROM application_deployment WHERE application_id = ?)", deployment.ID).
		Exec(s.Ctx)

	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Table("application_deployment").
		Where("application_id = ?", deployment.ID).
		Exec(s.Ctx)

	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Table("applications").
		Where("id = ?", deployment.ID).
		Exec(s.Ctx)
	return err
}

func (s *DeployStorage) GetApplicationDeployments(applicationID uuid.UUID) ([]shared_types.ApplicationDeployment, error) {
	var deployments []shared_types.ApplicationDeployment
	err := s.DB.NewSelect().
		Model(&deployments).
		Where("application_id = ?", applicationID).
		Scan(s.Ctx)
	return deployments, err
}

func (s *DeployStorage) GetPaginatedApplicationDeployments(applicationID uuid.UUID, page, pageSize int) ([]shared_types.ApplicationDeployment, int, error) {
	var deployments []shared_types.ApplicationDeployment
	offset := (page - 1) * pageSize

	totalCount, err := s.DB.NewSelect().
		Model((*shared_types.ApplicationDeployment)(nil)).
		Where("application_id = ?", applicationID).
		Count(s.Ctx)

	if err != nil {
		return nil, 0, err
	}

	err = s.DB.NewSelect().
		Model(&deployments).
		Relation("Status").
		Where("application_id = ?", applicationID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(s.Ctx)

	if err != nil {
		return nil, 0, err
	}

	return deployments, totalCount, nil
}

func (s *DeployStorage) GetLogs(applicationID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int, error) {
	offset := (page - 1) * pageSize

	query := s.DB.NewSelect().
		Model((*shared_types.ApplicationLogs)(nil)).
		Where("application_id = ?", applicationID)

	if level != "" {
		query = query.Where("LOWER(log) LIKE LOWER(?)", "%"+level+"%")
	}

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if searchTerm != "" {
		query = query.Where("LOWER(log) LIKE LOWER(?)", "%"+searchTerm+"%")
	}

	totalCount, err := query.Count(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	var logs []shared_types.ApplicationLogs
	err = query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(s.Ctx, &logs)

	if err != nil {
		return nil, 0, err
	}

	return logs, totalCount, nil
}

func (s *DeployStorage) GetDeploymentLogs(deploymentID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int, error) {
	offset := (page - 1) * pageSize

	query := s.DB.NewSelect().
		Model((*shared_types.ApplicationLogs)(nil)).
		Where("application_deployment_id = ?", deploymentID)

	if level != "" {
		query = query.Where("LOWER(log) LIKE LOWER(?)", "%"+level+"%")
	}

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if searchTerm != "" {
		query = query.Where("LOWER(log) LIKE LOWER(?)", "%"+searchTerm+"%")
	}

	totalCount, err := query.Count(s.Ctx)
	if err != nil {
		return nil, 0, err
	}

	var logs []shared_types.ApplicationLogs
	err = query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(s.Ctx, &logs)

	if err != nil {
		return nil, 0, err
	}

	return logs, totalCount, nil
}

func (s *DeployStorage) GetApplicationByRepositoryID(repositoryID uint64) (shared_types.Application, error) {
	var application shared_types.Application
	err := s.DB.NewSelect().
		Model(&application).
		Relation("Status").
		Relation("Deployments", func(q *bun.SelectQuery) *bun.SelectQuery { return q.Order("created_at DESC") }).
		Relation("Deployments.Status").
		Where("repository = ?", fmt.Sprintf("%d", repositoryID)).
		Scan(s.Ctx)

	if err != nil {
		return shared_types.Application{}, fmt.Errorf("failed to get application by repository ID: %w", err)
	}

	return application, nil
}

func (s *DeployStorage) GetApplicationByRepositoryIDAndBranch(repositoryID uint64, branch string) ([]shared_types.Application, error) {
	var applications []shared_types.Application
	err := s.DB.NewSelect().
		Model(&applications).
		Relation("Status").
		Relation("Deployments", func(q *bun.SelectQuery) *bun.SelectQuery { return q.Order("created_at DESC") }).
		Relation("Deployments.Status").
		Where("repository = ? AND branch = ?", fmt.Sprintf("%d", repositoryID), branch).
		Scan(s.Ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get applications by repository ID and branch: %w", err)
	}

	return applications, nil
}

func (s *DeployStorage) UpdateApplicationLabels(applicationID uuid.UUID, labels []string, organizationID uuid.UUID) error {
	_, err := s.DB.NewUpdate().
		Model((*shared_types.Application)(nil)).
		Set("labels = ?", pgdialect.Array(labels)).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("id = ? AND organization_id = ?", applicationID, organizationID).
		Exec(s.Ctx)
	return err
}
