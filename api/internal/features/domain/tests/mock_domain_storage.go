package tests

import (
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/mock"
)

type MockDomainStorage struct {
	mock.Mock
}

func NewMockDomainStorage() *MockDomainStorage {
	return &MockDomainStorage{}
}

func (m *MockDomainStorage) CreateDomain(domain *types.Domain) error {
	args := m.Called(domain)
	return args.Error(0)
}

func (m *MockDomainStorage) GetDomain(id string) (*types.Domain, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*types.Domain), args.Error(1)
}

func (m *MockDomainStorage) UpdateDomain(ID string, Name string) error {
	args := m.Called(ID, Name)
	return args.Error(0)
}

func (m *MockDomainStorage) DeleteDomain(domain *types.Domain) error {
	args := m.Called(domain)
	return args.Error(0)
}

func (m *MockDomainStorage) GetDomains() ([]types.Domain, error) {
	args := m.Called()
	return args.Get(0).([]types.Domain), args.Error(1)
}

func (m *MockDomainStorage) GetDomainByName(name string) (*types.Domain, error) {
	args := m.Called(name)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*types.Domain), args.Error(1)
}

func (m *MockDomainStorage) IsDomainExists(ID string) (bool, error) {
	args := m.Called(ID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDomainStorage) GetDomainOwnerByID(ID string) (string, error) {
	args := m.Called(ID)
	return args.String(0), args.Error(1)
}

func (m *MockDomainStorage) WithGetDomainError(id string, err error) *MockDomainStorage {
	m.On("GetDomain", id).Return(nil, err)
	return m
}

func (m *MockDomainStorage) WithGetDomainByNameError(name string, err error) *MockDomainStorage {
	m.On("GetDomainByName", name).Return(nil, err)
	return m
}

func (m *MockDomainStorage) WithGetDomain(id string, domain *types.Domain, err error) *MockDomainStorage {
	m.On("GetDomain", id).Return(domain, err)
	return m
}

func (m *MockDomainStorage) WithUpdateDomain(id string, name string, err error) *MockDomainStorage {
	m.On("UpdateDomain", id, name).Return(err)
	return m
}

func (m *MockDomainStorage) WithGetDomainsError(err error) *MockDomainStorage {
	m.On("GetDomains").Return(nil, err)
	return m
}