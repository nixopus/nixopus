package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

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
	GetApplications(page int, pageSize int) ([]shared_types.Application, int, error)
	UpdateApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error
	GetApplicationById(id string) (shared_types.Application, error)
	AddApplicationDeployment(deployment *shared_types.ApplicationDeployment) error
	AddApplicationDeploymentStatus(deployment_status *shared_types.ApplicationDeploymentStatus) error
	UpdateApplicationDeploymentStatus(applicationStatus *shared_types.ApplicationDeploymentStatus) error
	UpdateApplication(application *shared_types.Application) error
	GetApplicationDeploymentById(deploymentID string) (shared_types.ApplicationDeployment, error)
	DeleteDeployment(deployment *types.DeleteDeploymentRequest, userID uuid.UUID) error
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

func (s *DeployStorage) GetApplications(page, pageSize int) ([]shared_types.Application, int, error) {
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
		Relation("Domain").
		Relation("Status").
		Relation("Logs").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(s.Ctx)

	if err != nil {
		return nil, 0, err
	}

	return applications, totalCount, nil
}

func (s *DeployStorage) GetApplicationById(id string) (shared_types.Application, error) {
	var application shared_types.Application

	err := s.DB.NewSelect().
		Model(&application).
		Relation("Status").
		Relation("Logs", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("created_at DESC").Limit(100)
		}).
		Relation("Domain").
		Relation("Deployments", func(q *bun.SelectQuery) *bun.SelectQuery { return q.Order("created_at DESC") }).
		Where("a.id = ?", id).
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
