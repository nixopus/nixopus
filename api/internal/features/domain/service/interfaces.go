package service

import (
	"context"
	"net"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/queue"
)

// dnsTimeout caps every realNetLookup call. Tests override it to a short value
// to avoid blocking on DNS timeouts for unreachable names.
var dnsTimeout = 5 * time.Second

// NetLookup abstracts the net.Lookup* family so DNS logic can be tested
// without real network calls.
type NetLookup interface {
	LookupCNAME(host string) (string, error)
	LookupHost(host string) ([]string, error)
	LookupNS(name string) ([]*net.NS, error)
	LookupTXT(name string) ([]string, error)
}

// realNetLookup delegates to the stdlib net package using context-based
// lookups bounded by dnsTimeout.
type realNetLookup struct{}

func (r *realNetLookup) LookupCNAME(host string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
	defer cancel()
	return net.DefaultResolver.LookupCNAME(ctx, host)
}

func (r *realNetLookup) LookupHost(host string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
	defer cancel()
	return net.DefaultResolver.LookupHost(ctx, host)
}

func (r *realNetLookup) LookupNS(name string) ([]*net.NS, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
	defer cancel()
	return net.DefaultResolver.LookupNS(ctx, name)
}

func (r *realNetLookup) LookupTXT(name string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
	defer cancel()
	return net.DefaultResolver.LookupTXT(ctx, name)
}

// defaultResolver is the NetLookup used by exported wrapper functions.
// Tests can swap it via restoreDefaultResolver to avoid real network calls.
var defaultResolver NetLookup = &realNetLookup{}

// DNSResolver abstracts all DNS network operations and instruction generation
// so that service logic can be tested without real network calls.
type DNSResolver interface {
	DetectProvider(domain string) (string, error)
	GenerateToken() string
	VerifyDNSConfig(domain, target string) (bool, error)
	VerifyARecord(domain, machineIP string) (bool, error)
	CheckPropagation(domain string) (string, error)
	CheckPropagationBYOS(domain, machineIP string) (string, error)
	GenerateDNSInstructions(domain, targetSubdomain, provider string) []types.DNSInstruction
	GenerateDNSInstructionsBYOS(domain, machineIP, provider string) []types.DNSInstruction
}

// QueueClient abstracts task queue operations so they can be stubbed in tests.
type QueueClient interface {
	EnqueueRegisterCustomDomain(ctx context.Context, payload queue.CustomDomainPayload) error
	EnqueueRemoveCustomDomain(ctx context.Context, payload queue.RemoveCustomDomainPayload) error
}

// RealDNSResolver delegates to the package-level DNS functions backed by net.Lookup*.
type RealDNSResolver struct {
	net NetLookup
}

// NewRealDNSResolver creates a RealDNSResolver backed by the stdlib net package.
func NewRealDNSResolver() *RealDNSResolver {
	return &RealDNSResolver{net: &realNetLookup{}}
}

func (r *RealDNSResolver) DetectProvider(domain string) (string, error) {
	return detectDNSProvider(r.net, domain)
}

func (r *RealDNSResolver) GenerateToken() string {
	return GenerateVerificationToken()
}

func (r *RealDNSResolver) VerifyDNSConfig(domain, target string) (bool, error) {
	return verifyDNSConfiguration(r.net, domain, target)
}

func (r *RealDNSResolver) VerifyARecord(domain, machineIP string) (bool, error) {
	return verifyARecordMatchesMachineIP(r.net, domain, machineIP)
}

func (r *RealDNSResolver) CheckPropagation(domain string) (string, error) {
	return checkDNSPropagation(r.net, domain)
}

func (r *RealDNSResolver) CheckPropagationBYOS(domain, machineIP string) (string, error) {
	return checkDNSPropagationBYOS(r.net, domain, machineIP)
}

func (r *RealDNSResolver) GenerateDNSInstructions(domain, targetSubdomain, provider string) []types.DNSInstruction {
	return GenerateDNSInstructions(domain, targetSubdomain, provider)
}

func (r *RealDNSResolver) GenerateDNSInstructionsBYOS(domain, machineIP, provider string) []types.DNSInstruction {
	return GenerateDNSInstructionsBYOS(domain, machineIP, provider)
}

// RealQueueClient delegates to the real taskq-backed queue functions.
type RealQueueClient struct{}

func (q *RealQueueClient) EnqueueRegisterCustomDomain(ctx context.Context, payload queue.CustomDomainPayload) error {
	return queue.EnqueueRegisterCustomDomain(ctx, payload)
}

func (q *RealQueueClient) EnqueueRemoveCustomDomain(ctx context.Context, payload queue.RemoveCustomDomainPayload) error {
	return queue.EnqueueRemoveCustomDomain(ctx, payload)
}
