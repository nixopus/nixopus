package types

import "errors"

var (
	ErrDomainNotFound      = errors.New("domain not found")
	ErrInvalidRequestType  = errors.New("invalid request type")
	ErrMissingDomainName   = errors.New("domain name is required")
	ErrInvalidDomainID     = errors.New("invalid domain id")
	ErrMissingDomainID     = errors.New("domain id is required")
	ErrDomainAlreadyExists = errors.New("domain already exists")
	ErrNotAllowed          = errors.New("request not allowed")
	ErrDomainNameTooLong   = errors.New("domain name too long")
	ErrDomainNameTooShort  = errors.New("domain name too short")
)

type CreateDomainRequest struct {
	Name string `json:"name"`
}

type UpdateDomainRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DeleteDomainRequest struct {
	ID string `json:"id"`
}

type CreateDomainResponse struct {
	ID string `json:"id"`
}
