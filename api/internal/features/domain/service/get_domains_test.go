package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func TestGetDomains_Success(t *testing.T) {
	domains := []shared_types.Domain{
		{ID: uuid.New(), Name: "app.example.com"},
		{ID: uuid.New(), Name: "api.example.com"},
	}
	stor := &mockStorage{domains: domains}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})

	got, err := svc.GetDomains(testOrgID.String(), testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 domains, got %d", len(got))
	}
}

func TestGetDomains_Empty(t *testing.T) {
	stor := &mockStorage{domains: []shared_types.Domain{}}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})

	got, err := svc.GetDomains(testOrgID.String(), testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestGetDomains_StorageError(t *testing.T) {
	stor := &mockStorage{errGetDomains: errors.New("db error")}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})

	_, err := svc.GetDomains(testOrgID.String(), testUserID)
	if err == nil {
		t.Fatal("expected error from storage")
	}
}
