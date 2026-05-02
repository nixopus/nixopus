package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

// mockStorage implements DomainStorageInterface with configurable return values.
type mockStorage struct {
	domains      []shared_types.Domain
	domainByID   *shared_types.Domain
	domainByName *shared_types.Domain
	sshKey       *shared_types.SSHKey
	provision    *shared_types.UserProvisionDetails

	errGetDomains              error
	errGetDomainByID           error
	errGetDomainByName         error
	errCreate                  error
	errUpdateStatus            error
	errUpdateVerification      error
	errDelete                  error
	errGetSSHKey               error
	errGetProvisionByKey       error
	errGetProvisionBySubdomain error
}

func (m *mockStorage) GetDomains(_ string, _ uuid.UUID) ([]shared_types.Domain, error) {
	return m.domains, m.errGetDomains
}
func (m *mockStorage) CreateCustomDomain(_ *shared_types.Domain) error { return m.errCreate }
func (m *mockStorage) GetCustomDomainsByOrg(_ uuid.UUID) ([]shared_types.Domain, error) {
	return m.domains, m.errGetDomains
}
func (m *mockStorage) GetCustomDomainByID(_ uuid.UUID, _ uuid.UUID) (*shared_types.Domain, error) {
	return m.domainByID, m.errGetDomainByID
}
func (m *mockStorage) GetCustomDomainByName(_ string) (*shared_types.Domain, error) {
	return m.domainByName, m.errGetDomainByName
}
func (m *mockStorage) UpdateCustomDomainStatus(_ uuid.UUID, _ string) error {
	return m.errUpdateStatus
}
func (m *mockStorage) UpdateCustomDomainVerification(_ uuid.UUID, _ string, _ *string) error {
	return m.errUpdateVerification
}
func (m *mockStorage) DeleteCustomDomain(_ uuid.UUID) error { return m.errDelete }
func (m *mockStorage) GetDefaultSSHKeyByOrg(_ uuid.UUID) (*shared_types.SSHKey, error) {
	return m.sshKey, m.errGetSSHKey
}
func (m *mockStorage) GetProvisionDetailsBySSHKeyAndOrg(_, _ uuid.UUID) (*shared_types.UserProvisionDetails, error) {
	return m.provision, m.errGetProvisionByKey
}
func (m *mockStorage) GetProvisionDetailsBySubdomain(_ string) (*shared_types.UserProvisionDetails, error) {
	return m.provision, m.errGetProvisionBySubdomain
}

// mockDNS implements DNSResolver with configurable return values.
type mockDNS struct {
	provider              string
	token                 string
	verifyDNSConfigResult bool
	verifyDNSConfigErr    error
	verifyARecordResult   bool
	verifyARecordErr      error
	propagation           string
	propagationBYOS       string
	dnsInstructions       []types.DNSInstruction
	dnsInstructionsBYOS   []types.DNSInstruction
}

func (m *mockDNS) DetectProvider(_ string) (string, error) { return m.provider, nil }
func (m *mockDNS) GenerateToken() string                   { return m.token }
func (m *mockDNS) VerifyDNSConfig(_, _ string) (bool, error) {
	return m.verifyDNSConfigResult, m.verifyDNSConfigErr
}
func (m *mockDNS) VerifyARecord(_, _ string) (bool, error) {
	return m.verifyARecordResult, m.verifyARecordErr
}
func (m *mockDNS) CheckPropagation(_ string) (string, error) { return m.propagation, nil }
func (m *mockDNS) CheckPropagationBYOS(_, _ string) (string, error) {
	return m.propagationBYOS, nil
}
func (m *mockDNS) GenerateDNSInstructions(_, _, _ string) []types.DNSInstruction {
	return m.dnsInstructions
}
func (m *mockDNS) GenerateDNSInstructionsBYOS(_, _, _ string) []types.DNSInstruction {
	return m.dnsInstructionsBYOS
}

// mockQueue implements QueueClient with configurable return values.
type mockQueue struct {
	errEnqueueRegister error
	errEnqueueRemove   error
	registerCalled     bool
	removeCalled       bool
}

func (m *mockQueue) EnqueueRegisterCustomDomain(_ context.Context, _ queue.CustomDomainPayload) error {
	m.registerCalled = true
	return m.errEnqueueRegister
}
func (m *mockQueue) EnqueueRemoveCustomDomain(_ context.Context, _ queue.RemoveCustomDomainPayload) error {
	m.removeCalled = true
	return m.errEnqueueRemove
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestService(s *mockStorage, d *mockDNS, q *mockQueue) *DomainsService {
	l := logger.NewLogger()
	return NewDomainsServiceWith(context.Background(), l, s, d, q)
}

func ptrStr(s string) *string        { return &s }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

var (
	testUserID = uuid.New()
	testOrgID  = uuid.New()
	testDomID  = uuid.New()
	serverID   = uuid.New()
)

func validSSHKey(host string) *shared_types.SSHKey {
	return &shared_types.SSHKey{ID: uuid.New(), Host: ptrStr(host)}
}

func managedProvision(subdomain string) *shared_types.UserProvisionDetails {
	return &shared_types.UserProvisionDetails{
		Type:      "managed",
		Subdomain: ptrStr(subdomain),
		ServerID:  ptrUUID(serverID),
		GuestIP:   ptrStr("10.0.0.1"),
	}
}

func byosProvision() *shared_types.UserProvisionDetails {
	return &shared_types.UserProvisionDetails{Type: "user_owned"}
}

func existingDomain(target string) *shared_types.Domain {
	return &shared_types.Domain{
		ID:              testDomID,
		Name:            "app.example.com",
		Status:          "pending_dns",
		TargetSubdomain: ptrStr(target),
		DNSProvider:     ptrStr("cloudflare"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// ---------------------------------------------------------------------------
// AddCustomDomain tests
// ---------------------------------------------------------------------------

func TestAddCustomDomain_InvalidName_TooShort(t *testing.T) {
	svc := newTestService(&mockStorage{}, &mockDNS{}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "ab")
	if err == nil {
		t.Fatal("expected error for too-short domain name")
	}
}

func TestAddCustomDomain_InvalidName_NoTLD(t *testing.T) {
	svc := newTestService(&mockStorage{}, &mockDNS{}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "nodot")
	if err == nil {
		t.Fatal("expected error for domain name without TLD")
	}
}

func TestAddCustomDomain_GetByNameError(t *testing.T) {
	stor := &mockStorage{errGetDomainByName: errors.New("db error")}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err == nil {
		t.Fatal("expected error from storage")
	}
}

func TestAddCustomDomain_AlreadyExists(t *testing.T) {
	d := existingDomain("sub1")
	stor := &mockStorage{domainByName: d}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if !errors.Is(err, types.ErrDomainAlreadyExists) {
		t.Fatalf("expected ErrDomainAlreadyExists, got %v", err)
	}
}

func TestAddCustomDomain_NoDefaultSSHKey(t *testing.T) {
	stor := &mockStorage{errGetSSHKey: errors.New("no key")}
	svc := newTestService(stor, &mockDNS{token: "tok"}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err == nil {
		t.Fatal("expected error when SSH key missing")
	}
}

func TestAddCustomDomain_BYOS_NoHostIP(t *testing.T) {
	key := &shared_types.SSHKey{ID: uuid.New(), Host: nil}
	stor := &mockStorage{sshKey: key, provision: byosProvision()}
	svc := newTestService(stor, &mockDNS{token: "tok"}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err == nil {
		t.Fatal("expected error for BYOS with no IP")
	}
}

func TestAddCustomDomain_BYOS_CreateFails(t *testing.T) {
	stor := &mockStorage{
		sshKey:    validSSHKey("1.2.3.4"),
		provision: byosProvision(),
		errCreate: errors.New("db write error"),
	}
	dns := &mockDNS{token: "tok", provider: "cloudflare", dnsInstructionsBYOS: []types.DNSInstruction{{RecordType: "A"}}}
	svc := newTestService(stor, dns, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err == nil {
		t.Fatal("expected error from CreateCustomDomain")
	}
}

func TestAddCustomDomain_BYOS_Success(t *testing.T) {
	stor := &mockStorage{
		sshKey:    validSSHKey("1.2.3.4"),
		provision: byosProvision(),
	}
	instructions := []types.DNSInstruction{{RecordType: "A", Name: "app.example.com", Value: "1.2.3.4"}}
	dns := &mockDNS{token: "mytoken", provider: "cloudflare", dnsInstructionsBYOS: instructions}
	svc := newTestService(stor, dns, &mockQueue{})

	domain, got, provider, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.Status != "pending_dns" {
		t.Errorf("expected status pending_dns, got %s", domain.Status)
	}
	if len(got) == 0 {
		t.Error("expected DNS instructions")
	}
	if provider != "cloudflare" {
		t.Errorf("expected cloudflare, got %s", provider)
	}
	if *domain.TargetSubdomain != "1.2.3.4" {
		t.Errorf("expected target 1.2.3.4, got %s", *domain.TargetSubdomain)
	}
}

func TestAddCustomDomain_Managed_NoSubdomain(t *testing.T) {
	stor := &mockStorage{
		sshKey:    validSSHKey("1.2.3.4"),
		provision: &shared_types.UserProvisionDetails{Type: "managed"},
	}
	svc := newTestService(stor, &mockDNS{token: "tok"}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err == nil {
		t.Fatal("expected error when no subdomain configured")
	}
}

func TestAddCustomDomain_Managed_ProvisionLookupFails_NoSubdomain(t *testing.T) {
	stor := &mockStorage{
		sshKey:               validSSHKey("1.2.3.4"),
		errGetProvisionByKey: errors.New("no provision"),
	}
	svc := newTestService(stor, &mockDNS{token: "tok"}, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err == nil {
		t.Fatal("expected error: no subdomain because provision lookup failed")
	}
}

func TestAddCustomDomain_Managed_CreateFails(t *testing.T) {
	stor := &mockStorage{
		sshKey:    validSSHKey("1.2.3.4"),
		provision: managedProvision("org-sub"),
		errCreate: errors.New("write error"),
	}
	dns := &mockDNS{token: "tok", provider: "cloudflare", dnsInstructions: []types.DNSInstruction{{RecordType: "CNAME"}}}
	svc := newTestService(stor, dns, &mockQueue{})
	_, _, _, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err == nil {
		t.Fatal("expected error from CreateCustomDomain")
	}
}

func TestAddCustomDomain_Managed_Success(t *testing.T) {
	stor := &mockStorage{
		sshKey:    validSSHKey("1.2.3.4"),
		provision: managedProvision("org-sub"),
	}
	instructions := []types.DNSInstruction{{RecordType: "CNAME"}, {RecordType: "TXT"}}
	dns := &mockDNS{token: "tok", provider: "cloudflare", dnsInstructions: instructions}
	svc := newTestService(stor, dns, &mockQueue{})

	domain, got, provider, err := svc.AddCustomDomain(context.Background(), testUserID, testOrgID, "app.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *domain.TargetSubdomain != "org-sub" {
		t.Errorf("expected target org-sub, got %s", *domain.TargetSubdomain)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 instructions, got %d", len(got))
	}
	if provider != "cloudflare" {
		t.Errorf("expected cloudflare provider, got %s", provider)
	}
}

// ---------------------------------------------------------------------------
// VerifyCustomDomain tests
// ---------------------------------------------------------------------------

func TestVerifyCustomDomain_DomainNotFound(t *testing.T) {
	stor := &mockStorage{errGetDomainByID: errors.New("not found")}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	_, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyCustomDomain_BYOS_VerifyARecordError(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("1.2.3.4")}
	dns := &mockDNS{verifyARecordErr: errors.New("dns lookup error")}
	svc := newTestService(stor, dns, &mockQueue{})
	_, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error from VerifyARecord")
	}
}

func TestVerifyCustomDomain_BYOS_NotVerified(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("1.2.3.4")}
	dns := &mockDNS{verifyARecordResult: false}
	svc := newTestService(stor, dns, &mockQueue{})
	_, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if !errors.Is(err, types.ErrDNSNotVerified) {
		t.Fatalf("expected ErrDNSNotVerified, got %v", err)
	}
}

func TestVerifyCustomDomain_BYOS_UpdateVerificationFails(t *testing.T) {
	stor := &mockStorage{
		domainByID:            existingDomain("1.2.3.4"),
		errUpdateVerification: errors.New("db error"),
	}
	dns := &mockDNS{verifyARecordResult: true}
	svc := newTestService(stor, dns, &mockQueue{})
	_, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error from UpdateCustomDomainVerification")
	}
}

func TestVerifyCustomDomain_BYOS_Success_NoProvision(t *testing.T) {
	stor := &mockStorage{
		domainByID: existingDomain("1.2.3.4"),
	}
	dns := &mockDNS{verifyARecordResult: true}
	q := &mockQueue{}
	svc := newTestService(stor, dns, q)
	domain, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.Status != "dns_verified" {
		t.Errorf("expected dns_verified, got %s", domain.Status)
	}
	if !q.registerCalled {
		t.Error("expected EnqueueRegisterCustomDomain to be called")
	}
}

func TestVerifyCustomDomain_BYOS_EnqueueFails_NonFatal(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("1.2.3.4")}
	dns := &mockDNS{verifyARecordResult: true}
	q := &mockQueue{errEnqueueRegister: errors.New("queue unavailable")}
	svc := newTestService(stor, dns, q)
	// enqueue failure must NOT bubble up
	domain, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("enqueue failure should be non-fatal, got: %v", err)
	}
	if domain.Status != "dns_verified" {
		t.Errorf("expected dns_verified, got %s", domain.Status)
	}
}

func TestVerifyCustomDomain_Managed_VerifyDNSConfigError(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("org-sub")}
	dns := &mockDNS{verifyDNSConfigErr: errors.New("dns error")}
	svc := newTestService(stor, dns, &mockQueue{})
	_, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error from VerifyDNSConfig")
	}
}

func TestVerifyCustomDomain_Managed_NotVerified(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("org-sub")}
	dns := &mockDNS{verifyDNSConfigResult: false}
	svc := newTestService(stor, dns, &mockQueue{})
	_, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if !errors.Is(err, types.ErrDNSNotVerified) {
		t.Fatalf("expected ErrDNSNotVerified, got %v", err)
	}
}

func TestVerifyCustomDomain_Managed_WithProvisionDetails(t *testing.T) {
	provision := managedProvision("org-sub")
	stor := &mockStorage{
		domainByID: existingDomain("org-sub"),
		provision:  provision,
	}
	dns := &mockDNS{verifyDNSConfigResult: true}
	q := &mockQueue{}
	svc := newTestService(stor, dns, q)
	domain, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.Status != "dns_verified" {
		t.Errorf("expected dns_verified, got %s", domain.Status)
	}
}

func TestVerifyCustomDomain_Managed_NoTargetSubdomain(t *testing.T) {
	dom := existingDomain("")
	dom.TargetSubdomain = nil
	stor := &mockStorage{domainByID: dom}
	dns := &mockDNS{verifyDNSConfigResult: true}
	svc := newTestService(stor, dns, &mockQueue{})
	domain, err := svc.VerifyCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.Status != "dns_verified" {
		t.Errorf("expected dns_verified, got %s", domain.Status)
	}
}

// ---------------------------------------------------------------------------
// RemoveCustomDomain tests
// ---------------------------------------------------------------------------

func TestRemoveCustomDomain_DomainNotFound(t *testing.T) {
	stor := &mockStorage{errGetDomainByID: errors.New("not found")}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	err := svc.RemoveCustomDomain(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveCustomDomain_UpdateStatusFails(t *testing.T) {
	stor := &mockStorage{
		domainByID:      existingDomain("org-sub"),
		errUpdateStatus: errors.New("db error"),
	}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	err := svc.RemoveCustomDomain(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error from UpdateCustomDomainStatus")
	}
}

func TestRemoveCustomDomain_WithServerID(t *testing.T) {
	stor := &mockStorage{
		domainByID: existingDomain("org-sub"),
		provision:  managedProvision("org-sub"),
	}
	q := &mockQueue{}
	svc := newTestService(stor, &mockDNS{}, q)
	err := svc.RemoveCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !q.removeCalled {
		t.Error("expected EnqueueRemoveCustomDomain to be called")
	}
}

func TestRemoveCustomDomain_ProvisionLookupFails_StillEnqueues(t *testing.T) {
	stor := &mockStorage{
		domainByID:                 existingDomain("org-sub"),
		errGetProvisionBySubdomain: errors.New("lookup error"),
	}
	q := &mockQueue{}
	svc := newTestService(stor, &mockDNS{}, q)
	err := svc.RemoveCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !q.removeCalled {
		t.Error("expected EnqueueRemoveCustomDomain to be called even without server ID")
	}
}

func TestRemoveCustomDomain_NoTargetSubdomain(t *testing.T) {
	dom := existingDomain("")
	dom.TargetSubdomain = nil
	stor := &mockStorage{domainByID: dom}
	q := &mockQueue{}
	svc := newTestService(stor, &mockDNS{}, q)
	err := svc.RemoveCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveCustomDomain_EnqueueFails_NonFatal(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("org-sub")}
	q := &mockQueue{errEnqueueRemove: errors.New("queue error")}
	svc := newTestService(stor, &mockDNS{}, q)
	err := svc.RemoveCustomDomain(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("enqueue failure must be non-fatal, got: %v", err)
	}
}

func TestRemoveCustomDomain_DeleteFails(t *testing.T) {
	stor := &mockStorage{
		domainByID: existingDomain("org-sub"),
		errDelete:  errors.New("delete failed"),
	}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	err := svc.RemoveCustomDomain(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error from DeleteCustomDomain")
	}
}

// ---------------------------------------------------------------------------
// ListCustomDomains tests
// ---------------------------------------------------------------------------

func TestListCustomDomains_Success(t *testing.T) {
	domains := []shared_types.Domain{{ID: uuid.New(), Name: "app.example.com"}}
	stor := &mockStorage{domains: domains}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	got, err := svc.ListCustomDomains(context.Background(), testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 domain, got %d", len(got))
	}
}

func TestListCustomDomains_Error(t *testing.T) {
	stor := &mockStorage{errGetDomains: errors.New("db error")}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	_, err := svc.ListCustomDomains(context.Background(), testOrgID)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// CheckDNSStatus tests
// ---------------------------------------------------------------------------

func TestCheckDNSStatus_DomainNotFound(t *testing.T) {
	stor := &mockStorage{errGetDomainByID: errors.New("not found")}
	svc := newTestService(stor, &mockDNS{}, &mockQueue{})
	_, _, err := svc.CheckDNSStatus(context.Background(), testDomID, testOrgID)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckDNSStatus_BYOS_Verified(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("1.2.3.4")}
	dns := &mockDNS{verifyARecordResult: true}
	svc := newTestService(stor, dns, &mockQueue{})
	verified, status, err := svc.CheckDNSStatus(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !verified || status != "verified" {
		t.Errorf("expected verified=true/verified, got %v/%s", verified, status)
	}
}

func TestCheckDNSStatus_BYOS_Propagating(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("1.2.3.4")}
	dns := &mockDNS{verifyARecordResult: false, propagationBYOS: "propagating"}
	svc := newTestService(stor, dns, &mockQueue{})
	verified, status, err := svc.CheckDNSStatus(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified || status != "propagating" {
		t.Errorf("expected verified=false/propagating, got %v/%s", verified, status)
	}
}

func TestCheckDNSStatus_Managed_VerifyError(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("org-sub")}
	dns := &mockDNS{verifyDNSConfigErr: errors.New("dns failure")}
	svc := newTestService(stor, dns, &mockQueue{})
	verified, status, err := svc.CheckDNSStatus(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified || status != "not_configured" {
		t.Errorf("expected not_configured on dns error, got %v/%s", verified, status)
	}
}

func TestCheckDNSStatus_Managed_Verified(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("org-sub")}
	dns := &mockDNS{verifyDNSConfigResult: true}
	svc := newTestService(stor, dns, &mockQueue{})
	verified, status, err := svc.CheckDNSStatus(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !verified || status != "verified" {
		t.Errorf("expected verified=true/verified, got %v/%s", verified, status)
	}
}

func TestCheckDNSStatus_Managed_Propagating(t *testing.T) {
	stor := &mockStorage{domainByID: existingDomain("org-sub")}
	dns := &mockDNS{verifyDNSConfigResult: false, propagation: "propagating"}
	svc := newTestService(stor, dns, &mockQueue{})
	verified, status, err := svc.CheckDNSStatus(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified || status != "propagating" {
		t.Errorf("expected propagating, got %v/%s", verified, status)
	}
}

func TestCheckDNSStatus_NoTargetSubdomain(t *testing.T) {
	dom := existingDomain("")
	dom.TargetSubdomain = nil
	stor := &mockStorage{domainByID: dom}
	dns := &mockDNS{verifyDNSConfigResult: true}
	svc := newTestService(stor, dns, &mockQueue{})
	verified, status, err := svc.CheckDNSStatus(context.Background(), testDomID, testOrgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !verified || status != "verified" {
		t.Errorf("expected verified=true/verified, got %v/%s", verified, status)
	}
}
