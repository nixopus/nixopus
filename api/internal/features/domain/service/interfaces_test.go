package service

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/queue"
)

// ---------------------------------------------------------------------------
// NewDomainsService
// ---------------------------------------------------------------------------

func TestNewDomainsService_NonNil(t *testing.T) {
	svc := NewDomainsService(context.Background(), logger.NewLogger(), &mockStorage{})
	if svc == nil {
		t.Fatal("expected non-nil DomainsService")
	}
}

// ---------------------------------------------------------------------------
// NewRealDNSResolver
// ---------------------------------------------------------------------------

func TestNewRealDNSResolver_NonNil(t *testing.T) {
	r := NewRealDNSResolver(nil)
	if r == nil {
		t.Fatal("expected non-nil RealDNSResolver")
	}
}

// ---------------------------------------------------------------------------
// RealDNSResolver — all methods delegate to lowercase functions via r.net
// ---------------------------------------------------------------------------

func TestRealDNSResolver_DetectProvider(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{nsRecords: []*net.NS{ns("ns1.cloudflare.com.")}}}
	provider, err := r.DetectProvider("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "cloudflare" {
		t.Errorf("expected cloudflare, got %s", provider)
	}
}

func TestRealDNSResolver_GenerateToken(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{}}
	tok := r.GenerateToken()
	if len(tok) != 32 {
		t.Errorf("expected 32-char token, got %d", len(tok))
	}
}

func TestRealDNSResolver_VerifyDNSConfig(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{cname: "myorg.nixopus.ai."}}
	ok, err := r.VerifyDNSConfig("app.example.com", "myorg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified")
	}
}

func TestRealDNSResolver_VerifyARecord(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{hosts: []string{"1.2.3.4"}}}
	ok, err := r.VerifyARecord("app.example.com", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified")
	}
}

func TestRealDNSResolver_CheckPropagation(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{cname: "myorg.nixopus.ai."}}
	status, err := r.CheckPropagation("app.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "verified" {
		t.Errorf("expected verified, got %s", status)
	}
}

func TestRealDNSResolver_CheckPropagationBYOS(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{hosts: []string{"1.2.3.4"}}}
	status, err := r.CheckPropagationBYOS("app.example.com", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "verified" {
		t.Errorf("expected verified, got %s", status)
	}
}

func TestRealDNSResolver_GenerateDNSInstructions(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{}}
	instrs := r.GenerateDNSInstructions("app.example.com", "org-sub", "cloudflare")
	if len(instrs) != 2 {
		t.Errorf("expected 2 instructions, got %d", len(instrs))
	}
}

func TestRealDNSResolver_GenerateDNSInstructionsBYOS(t *testing.T) {
	r := &RealDNSResolver{net: &mockNetLookup{}}
	instrs := r.GenerateDNSInstructionsBYOS("app.example.com", "1.2.3.4", "cloudflare")
	if len(instrs) != 1 {
		t.Errorf("expected 1 instruction, got %d", len(instrs))
	}
}

// ---------------------------------------------------------------------------
// RealQueueClient — both methods return "not initialized" error when queue
// has not been set up, which is the case in unit tests.
// ---------------------------------------------------------------------------

func TestRealQueueClient_EnqueueRegister_NotInitialized(t *testing.T) {
	q := &RealQueueClient{}
	err := q.EnqueueRegisterCustomDomain(context.Background(), queue.CustomDomainPayload{})
	if err == nil {
		t.Fatal("expected error when queue not initialized")
	}
}

func TestRealQueueClient_EnqueueRemove_NotInitialized(t *testing.T) {
	q := &RealQueueClient{}
	err := q.EnqueueRemoveCustomDomain(context.Background(), queue.RemoveCustomDomainPayload{})
	if err == nil {
		t.Fatal("expected error when queue not initialized")
	}
}

// ---------------------------------------------------------------------------
// realNetLookup — thin stdlib wrappers; exercise each method to mark as covered.
// dnsTimeout is set to 200ms so DNS timeouts don't stall the test suite.
// ---------------------------------------------------------------------------

func TestRealNetLookup_Methods(t *testing.T) {
	old := dnsTimeout
	dnsTimeout = 200 * time.Millisecond
	t.Cleanup(func() { dnsTimeout = old })

	r := &realNetLookup{}
	// We only need each statement to execute for coverage; results are not asserted
	// because they vary by platform and DNS resolver configuration.
	_, _ = r.LookupCNAME("localhost")
	_, _ = r.LookupHost("localhost")
	_, _ = r.LookupNS("localhost")
	_, _ = r.LookupTXT("localhost")
}

// Ensure types import is used (GenerateDNSInstructions returns []types.DNSInstruction).
var _ []types.DNSInstruction
