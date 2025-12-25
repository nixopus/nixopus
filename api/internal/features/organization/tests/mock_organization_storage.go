package tests

import (
	"errors"
	"sync"
	"time"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// MockOrganizationStore implements the OrganizationRepository interface for testing
type MockOrganizationStore struct {
	mutex             sync.RWMutex
	organizations     map[string]shared_types.Organization
	organizationUsers map[string][]shared_types.OrganizationUsers
}

// NewMockOrganizationStore creates a new instance of MockOrganizationStore
func NewMockOrganizationStore() *MockOrganizationStore {
	return &MockOrganizationStore{
		organizations:     make(map[string]shared_types.Organization),
		organizationUsers: make(map[string][]shared_types.OrganizationUsers),
	}
}

// GetOrganizations returns all organizations
func (m *MockOrganizationStore) GetOrganizations() ([]shared_types.Organization, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	organizations := make([]shared_types.Organization, 0, len(m.organizations))
	for _, org := range m.organizations {
		organizations = append(organizations, org)
	}
	return organizations, nil
}

// CreateOrganization adds a new organization to the mock store
func (m *MockOrganizationStore) CreateOrganization(organization shared_types.Organization) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.organizations[organization.ID.String()]; exists {
		return errors.New("organization with this ID already exists")
	}

	// Check if organization with the same name exists
	for _, org := range m.organizations {
		if org.Name == organization.Name {
			return errors.New("organization with this name already exists")
		}
	}

	m.organizations[organization.ID.String()] = organization
	m.organizationUsers[organization.ID.String()] = []shared_types.OrganizationUsers{}
	return nil
}

// GetOrganization retrieves an organization by its ID
func (m *MockOrganizationStore) GetOrganization(id string) (*shared_types.Organization, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	org, exists := m.organizations[id]
	if !exists {
		return &shared_types.Organization{}, nil
	}
	return &org, nil
}

// UpdateOrganization updates an existing organization
func (m *MockOrganizationStore) UpdateOrganization(organization *shared_types.Organization) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.organizations[organization.ID.String()]; !exists {
		return errors.New("organization not found")
	}

	for id, org := range m.organizations {
		if id != organization.ID.String() && org.Name == organization.Name {
			return errors.New("organization with this name already exists")
		}
	}

	m.organizations[organization.ID.String()] = *organization
	return nil
}

// DeleteOrganization removes an organization and its users
func (m *MockOrganizationStore) DeleteOrganization(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.organizations[id]; !exists {
		return errors.New("organization not found")
	}

	delete(m.organizations, id)
	delete(m.organizationUsers, id)
	return nil
}

// GetOrganizationUsers retrieves all users for an organization
func (m *MockOrganizationStore) GetOrganizationUsers(id string) ([]shared_types.OrganizationUsers, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, exists := m.organizations[id]; !exists {
		return []shared_types.OrganizationUsers{}, errors.New("organization not found")
	}

	users, exists := m.organizationUsers[id]
	if !exists {
		return []shared_types.OrganizationUsers{}, nil
	}

	usersCopy := make([]shared_types.OrganizationUsers, len(users))
	copy(usersCopy, users)
	return usersCopy, nil
}

// AddUserToOrganization adds a user to an organization
func (m *MockOrganizationStore) AddUserToOrganization(organizationUser shared_types.OrganizationUsers) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.organizations[organizationUser.OrganizationID.String()]; !exists {
		return errors.New("organization not found")
	}

	for _, user := range m.organizationUsers[organizationUser.OrganizationID.String()] {
		if user.UserID == organizationUser.UserID && user.DeletedAt == nil {
			return errors.New("user already exists in this organization")
		}
	}

	m.organizationUsers[organizationUser.OrganizationID.String()] = append(
		m.organizationUsers[organizationUser.OrganizationID.String()],
		organizationUser,
	)
	return nil
}

// RemoveUserFromOrganization removes a user from an organization
func (m *MockOrganizationStore) RemoveUserFromOrganization(userID string, organizationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.organizations[organizationID]; !exists {
		return errors.New("organization not found")
	}

	users := m.organizationUsers[organizationID]
	for i, user := range users {
		if user.UserID.String() == userID {
			users[i] = users[len(users)-1]
			users = users[:len(users)-1]
			m.organizationUsers[organizationID] = users
			return nil
		}
	}

	return errors.New("user not found in organization")
}

// FindUserInOrganization finds a user in an organization
func (m *MockOrganizationStore) FindUserInOrganization(userID string, organizationID string) (*shared_types.OrganizationUsers, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, exists := m.organizations[organizationID]; !exists {
		return &shared_types.OrganizationUsers{}, errors.New("organization not found")
	}

	for _, user := range m.organizationUsers[organizationID] {
		if user.UserID.String() == userID && user.DeletedAt == nil {
			userCopy := user
			return &userCopy, nil
		}
	}

	return &shared_types.OrganizationUsers{}, nil
}

// GetOrganizationByName retrieves an organization by its name
func (m *MockOrganizationStore) GetOrganizationByName(name string) (*shared_types.Organization, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, org := range m.organizations {
		if org.Name == name {
			orgCopy := org
			return &orgCopy, nil
		}
	}

	return &shared_types.Organization{}, nil
}

// MarkUserDeletedInOrganization is a helper method that marks a user as deleted in an organization
// This simulates a soft delete that might be used in a real application
func (m *MockOrganizationStore) MarkUserDeletedInOrganization(userID string, organizationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.organizations[organizationID]; !exists {
		return errors.New("organization not found")
	}

	users := m.organizationUsers[organizationID]
	for i, user := range users {
		if user.UserID.String() == userID {
			now := time.Now()
			users[i].DeletedAt = &now
			m.organizationUsers[organizationID] = users
			return nil
		}
	}

	return errors.New("user not found in organization")
}
