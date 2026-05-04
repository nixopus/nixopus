package caddy

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/nixopus/nixopus/api/internal/features/deploy/storage"
	deploy_types "github.com/nixopus/nixopus/api/internal/features/deploy/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// deployRepoTestStub satisfies storage.DeployRepository for route-builder tests.
// Only GetApplicationServers is observed; everything else returns zero values.
type deployRepoTestStub struct {
	servers         []shared_types.ApplicationServer
	err             error
	deployedApps    []shared_types.Application
	deployedAppsErr error
}

func (s *deployRepoTestStub) RunInTransaction(fn func(tx bun.Tx) error) error {
	return nil
}
func (s *deployRepoTestStub) IsNameAlreadyTaken(name string) (bool, error) {
	return false, nil
}
func (s *deployRepoTestStub) IsDomainAlreadyTaken(domain string) (bool, error) {
	return false, nil
}
func (s *deployRepoTestStub) IsPortAlreadyTaken(port int) (bool, error) {
	return false, nil
}
func (s *deployRepoTestStub) IsDomainValid(domain string) (bool, error) {
	return false, nil
}
func (s *deployRepoTestStub) AddApplication(application *shared_types.Application) error {
	return nil
}
func (s *deployRepoTestStub) AddApplicationLogs(applicationLogs *shared_types.ApplicationLogs) error {
	return nil
}
func (s *deployRepoTestStub) AddApplicationLogsBatch(logs []shared_types.ApplicationLogs) error {
	return nil
}
func (s *deployRepoTestStub) AddApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error {
	return nil
}
func (s *deployRepoTestStub) GetApplications(page int, pageSize int, sortBy string, sortDirection string, organizationID uuid.UUID, serverID *uuid.UUID) ([]shared_types.Application, int, error) {
	return nil, 0, nil
}
func (s *deployRepoTestStub) UpdateApplicationStatus(applicationStatus *shared_types.ApplicationStatus) error {
	return nil
}
func (s *deployRepoTestStub) GetApplicationById(id string, organizationID uuid.UUID) (shared_types.Application, error) {
	return shared_types.Application{}, nil
}
func (s *deployRepoTestStub) GetApplicationBasicById(id string, organizationID uuid.UUID) (shared_types.Application, error) {
	return shared_types.Application{}, nil
}
func (s *deployRepoTestStub) AddApplicationDeployment(deployment *shared_types.ApplicationDeployment) error {
	return nil
}
func (s *deployRepoTestStub) AddApplicationDeploymentStatus(deployment_status *shared_types.ApplicationDeploymentStatus) error {
	return nil
}
func (s *deployRepoTestStub) UpdateApplicationDeploymentStatus(applicationStatus *shared_types.ApplicationDeploymentStatus) error {
	return nil
}
func (s *deployRepoTestStub) UpdateApplication(application *shared_types.Application) error {
	return nil
}
func (s *deployRepoTestStub) GetApplicationDeploymentById(deploymentID string) (shared_types.ApplicationDeployment, error) {
	return shared_types.ApplicationDeployment{}, nil
}
func (s *deployRepoTestStub) DeleteDeployment(deployment *deploy_types.DeleteDeploymentRequest, userID uuid.UUID) error {
	return nil
}
func (s *deployRepoTestStub) UpdateApplicationDeployment(deployment *shared_types.ApplicationDeployment) error {
	return nil
}
func (s *deployRepoTestStub) GetApplicationDeployments(applicationID uuid.UUID) ([]shared_types.ApplicationDeployment, error) {
	return nil, nil
}
func (s *deployRepoTestStub) GetPaginatedApplicationDeployments(applicationID uuid.UUID, page, pageSize int) ([]shared_types.ApplicationDeployment, int, error) {
	return nil, 0, nil
}
func (s *deployRepoTestStub) GetLogs(applicationID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int, error) {
	return nil, 0, nil
}
func (s *deployRepoTestStub) AddApplicationDomains(applicationID uuid.UUID, domains []string) error {
	return nil
}
func (s *deployRepoTestStub) RemoveApplicationDomain(applicationID uuid.UUID, domain string) error {
	return nil
}
func (s *deployRepoTestStub) GetApplicationDomains(applicationID uuid.UUID) ([]shared_types.ApplicationDomain, error) {
	return nil, nil
}
func (s *deployRepoTestStub) GetDeploymentLogs(deploymentID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int, error) {
	return nil, 0, nil
}
func (s *deployRepoTestStub) GetApplicationByRepositoryID(repositoryID uint64) (shared_types.Application, error) {
	return shared_types.Application{}, nil
}
func (s *deployRepoTestStub) GetApplicationByRepositoryIDAndBranch(repositoryID uint64, branch string) ([]shared_types.Application, error) {
	return nil, nil
}
func (s *deployRepoTestStub) UpdateApplicationLabels(applicationID uuid.UUID, labels []string, organizationID uuid.UUID) error {
	return nil
}
func (s *deployRepoTestStub) GetProjectsByFamilyID(familyID uuid.UUID, organizationID uuid.UUID) ([]shared_types.Application, error) {
	return nil, nil
}
func (s *deployRepoTestStub) UpdateApplicationFamilyID(applicationID uuid.UUID, familyID *uuid.UUID) error {
	return nil
}
func (s *deployRepoTestStub) IsEnvironmentInFamily(familyID uuid.UUID, environment shared_types.Environment) (bool, error) {
	return false, nil
}
func (s *deployRepoTestStub) GetEnvironmentsInFamily(familyID uuid.UUID, organizationID uuid.UUID) ([]shared_types.Environment, error) {
	return nil, nil
}
func (s *deployRepoTestStub) CountFamilyMembers(familyID uuid.UUID) (int, error) {
	return 0, nil
}
func (s *deployRepoTestStub) ClearFamilyIDIfSingleMember(familyID uuid.UUID) error {
	return nil
}
func (s *deployRepoTestStub) GetLatestDeployments(organizationID uuid.UUID, limit int) ([]shared_types.ApplicationDeployment, error) {
	return nil, nil
}
func (s *deployRepoTestStub) GetDeployedApplications(organizationID uuid.UUID) ([]shared_types.Application, error) {
	if s.deployedAppsErr != nil {
		return nil, s.deployedAppsErr
	}
	return s.deployedApps, nil
}
func (s *deployRepoTestStub) GetLatestDeploymentsByServer(organizationID uuid.UUID, serverID uuid.UUID, limit int) ([]shared_types.ApplicationDeployment, error) {
	return nil, nil
}
func (s *deployRepoTestStub) GetLatestS3Deployment(applicationID uuid.UUID) (*shared_types.ApplicationDeployment, error) {
	return nil, nil
}
func (s *deployRepoTestStub) UpsertComposeServices(applicationID uuid.UUID, services []shared_types.ComposeService) error {
	return nil
}
func (s *deployRepoTestStub) GetComposeServices(applicationID uuid.UUID) ([]shared_types.ComposeService, error) {
	return nil, nil
}
func (s *deployRepoTestStub) GetComposeServiceByName(applicationID uuid.UUID, serviceName string) (*shared_types.ComposeService, error) {
	return nil, nil
}
func (s *deployRepoTestStub) AddApplicationDomainWithService(applicationID uuid.UUID, domain string, composeServiceID *uuid.UUID, port *int) error {
	return nil
}
func (s *deployRepoTestStub) UpdateApplicationDomainService(applicationID uuid.UUID, domain string, composeServiceID *uuid.UUID, port *int) error {
	return nil
}
func (s *deployRepoTestStub) GetApplicationServers(appID uuid.UUID) ([]shared_types.ApplicationServer, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.servers, nil
}
func (s *deployRepoTestStub) SetApplicationServers(appID uuid.UUID, serverIDs []uuid.UUID, primaryServerID *uuid.UUID, routingStrategy shared_types.RoutingStrategy) error {
	return nil
}
func (s *deployRepoTestStub) EnsureApplicationServers(appID uuid.UUID, orgID uuid.UUID) error {
	return nil
}
func (s *deployRepoTestStub) CopyApplicationServers(srcAppID, dstAppID uuid.UUID) error {
	return nil
}
func (s *deployRepoTestStub) DeleteApplicationDeploymentByID(id uuid.UUID) error {
	return nil
}
func (s *deployRepoTestStub) ClearDeploymentArtifactFields(deploymentID uuid.UUID) error {
	return nil
}

var _ storage.DeployRepository = (*deployRepoTestStub)(nil)
