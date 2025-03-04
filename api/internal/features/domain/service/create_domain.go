package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateDomain creates a new domain in the application.
//
// It takes a CreateDomainRequest, which contains the domain name, and a user ID.
// The user ID is used to associate the domain with a user.
//
// It returns a CreateDomainResponse, which is always empty, and an error.
// The error is either ErrDomainAlreadyExists, or any error that occurred
// while creating the domain in the storage layer.
func (s *DomainsService) CreateDomain(req types.CreateDomainRequest, userID string) (types.CreateDomainResponse, error) {
	fmt.Printf("create domain request received: domain_name=%s, user_id=%s\n", req.Name, userID)

	existing_domain, err := s.storage.GetDomainByName(req.Name)
	if err != nil {
		fmt.Printf("error while retrieving domain: error=%s\n", err.Error())
		return types.CreateDomainResponse{}, err
	}

	if existing_domain != nil {
		fmt.Printf("domain already exists: domain_name=%s\n", req.Name)
		return types.CreateDomainResponse{}, types.ErrDomainAlreadyExists
	}

	domain := &shared_types.Domain{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
		Name:      req.Name,
	}

	if err := s.storage.CreateDomain(domain); err != nil {
		fmt.Printf("error while creating domain: error=%s\n", err.Error())
		return types.CreateDomainResponse{}, err
	}

	fmt.Printf("domain created successfully: domain_id=%s\n", domain.ID)

	return types.CreateDomainResponse{}, nil
}
