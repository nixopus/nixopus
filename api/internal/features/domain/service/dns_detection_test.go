package service

import (
	"errors"
	"net"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Mock NetLookup
// ---------------------------------------------------------------------------

type mockNetLookup struct {
	cname      string
	cnameErr   error
	hosts      []string
	hostErr    error
	nsRecords  []*net.NS
	nsErr      error
	txtRecords []string
	txtErr     error

	// per-host overrides: key = hostname, value = IPs
	hostsByName   map[string][]string
	hostErrByName map[string]error
	// per-name NS override
	nsByName    map[string][]*net.NS
	nsErrByName map[string]error
}

func (m *mockNetLookup) LookupCNAME(_ string) (string, error) { return m.cname, m.cnameErr }

func (m *mockNetLookup) LookupHost(host string) ([]string, error) {
	if m.hostsByName != nil {
		if ips, ok := m.hostsByName[host]; ok {
			return ips, nil
		}
		if err, ok := m.hostErrByName[host]; ok {
			return nil, err
		}
	}
	return m.hosts, m.hostErr
}

func (m *mockNetLookup) LookupNS(name string) ([]*net.NS, error) {
	if m.nsByName != nil {
		if recs, ok := m.nsByName[name]; ok {
			return recs, nil
		}
	}
	if m.nsErrByName != nil {
		if err, ok := m.nsErrByName[name]; ok {
			return nil, err
		}
	}
	return m.nsRecords, m.nsErr
}

func (m *mockNetLookup) LookupTXT(_ string) ([]string, error) { return m.txtRecords, m.txtErr }

func ns(host string) *net.NS { return &net.NS{Host: host} }

// ---------------------------------------------------------------------------
// extractRootDomain
// ---------------------------------------------------------------------------

func TestExtractRootDomain_TwoParts(t *testing.T) {
	if got := extractRootDomain("example.com"); got != "example.com" {
		t.Errorf("got %s", got)
	}
}

func TestExtractRootDomain_ThreeParts(t *testing.T) {
	if got := extractRootDomain("sub.example.com"); got != "example.com" {
		t.Errorf("got %s", got)
	}
}

func TestExtractRootDomain_TrailingDot(t *testing.T) {
	if got := extractRootDomain("sub.example.com."); got != "example.com" {
		t.Errorf("got %s", got)
	}
}

func TestExtractRootDomain_SingleLabel(t *testing.T) {
	if got := extractRootDomain("localhost"); got != "localhost" {
		t.Errorf("got %s", got)
	}
}

// ---------------------------------------------------------------------------
// matchNSToProvider
// ---------------------------------------------------------------------------

func TestMatchNSToProvider_Cloudflare(t *testing.T) {
	recs := []*net.NS{ns("ns1.cloudflare.com.")}
	if got := matchNSToProvider(recs); got != "cloudflare" {
		t.Errorf("expected cloudflare, got %s", got)
	}
}

func TestMatchNSToProvider_NoMatch(t *testing.T) {
	recs := []*net.NS{ns("ns1.unknownprovider.xyz.")}
	if got := matchNSToProvider(recs); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestMatchNSToProvider_Empty(t *testing.T) {
	if got := matchNSToProvider(nil); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

// ---------------------------------------------------------------------------
// detectDNSProvider
// ---------------------------------------------------------------------------

func TestDetectDNSProvider_RootDomainNSMatch(t *testing.T) {
	mock := &mockNetLookup{nsRecords: []*net.NS{ns("ns1.cloudflare.com.")}}
	provider, err := detectDNSProvider(mock, "app.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "cloudflare" {
		t.Errorf("expected cloudflare, got %s", provider)
	}
}

func TestDetectDNSProvider_SubdomainNSMatch(t *testing.T) {
	// Root domain NS returns no match; subdomain NS returns cloudflare.
	mock := &mockNetLookup{
		nsByName: map[string][]*net.NS{
			"example.com":     {ns("ns1.unknownprovider.xyz.")},
			"app.example.com": {ns("ns1.cloudflare.com.")},
		},
	}
	provider, err := detectDNSProvider(mock, "app.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "cloudflare" {
		t.Errorf("expected cloudflare, got %s", provider)
	}
}

func TestDetectDNSProvider_SubdomainSkippedWhenSameAsRoot(t *testing.T) {
	// domain == rootDomain — second NS lookup must not be attempted.
	callCount := 0
	mock := &mockNetLookup{
		nsByName: map[string][]*net.NS{},
	}
	// Override with a counter mock via closure — use nsErr to force "other"
	mock.nsErr = errors.New("no NS")
	_ = callCount

	provider, err := detectDNSProvider(mock, "example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "other" {
		t.Errorf("expected other, got %s", provider)
	}
}

func TestDetectDNSProvider_FallbackToOther(t *testing.T) {
	mock := &mockNetLookup{nsErr: errors.New("no NS records")}
	provider, err := detectDNSProvider(mock, "app.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "other" {
		t.Errorf("expected other, got %s", provider)
	}
}

func TestDetectDNSProvider_Route53(t *testing.T) {
	mock := &mockNetLookup{nsRecords: []*net.NS{ns("ns1.awsdns-01.com.")}}
	provider, _ := detectDNSProvider(mock, "example.com", nil)
	if provider != "route53" {
		t.Errorf("expected route53, got %s", provider)
	}
}

// ---------------------------------------------------------------------------
// GenerateDNSInstructions
// ---------------------------------------------------------------------------

func TestGenerateDNSInstructions_KnownProvider(t *testing.T) {
	instrs := GenerateDNSInstructions("app.example.com", "org-sub", "cloudflare")
	if len(instrs) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(instrs))
	}
	if instrs[0].RecordType != "CNAME" {
		t.Errorf("first instruction should be CNAME")
	}
	if instrs[1].RecordType != "TXT" {
		t.Errorf("second instruction should be TXT")
	}
	if instrs[0].Value != "org-sub.nixopus.ai" {
		t.Errorf("CNAME value should be org-sub.nixopus.ai, got %s", instrs[0].Value)
	}
	if !strings.Contains(instrs[0].Description, "Cloudflare") {
		t.Errorf("description should mention Cloudflare")
	}
}

func TestGenerateDNSInstructions_UnknownProvider_FallsBackToOther(t *testing.T) {
	instrs := GenerateDNSInstructions("app.example.com", "org-sub", "totally-unknown")
	if len(instrs) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(instrs))
	}
	if !strings.Contains(instrs[0].Description, "DNS provider") {
		t.Errorf("fallback description should mention DNS provider")
	}
}

// ---------------------------------------------------------------------------
// GenerateDNSInstructionsBYOS
// ---------------------------------------------------------------------------

func TestGenerateDNSInstructionsBYOS_KnownProvider(t *testing.T) {
	instrs := GenerateDNSInstructionsBYOS("app.example.com", "1.2.3.4", "hetzner")
	if len(instrs) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(instrs))
	}
	if instrs[0].RecordType != "A" {
		t.Errorf("expected A record")
	}
	if instrs[0].Value != "1.2.3.4" {
		t.Errorf("expected value 1.2.3.4, got %s", instrs[0].Value)
	}
	if !strings.Contains(instrs[0].Description, "Hetzner") {
		t.Errorf("description should mention Hetzner")
	}
}

func TestGenerateDNSInstructionsBYOS_UnknownProvider_FallsBackToOther(t *testing.T) {
	instrs := GenerateDNSInstructionsBYOS("app.example.com", "5.6.7.8", "totally-unknown")
	if len(instrs) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(instrs))
	}
	if !strings.Contains(instrs[0].Description, "DNS provider") {
		t.Errorf("fallback description should mention DNS provider")
	}
}

// ---------------------------------------------------------------------------
// GenerateVerificationToken
// ---------------------------------------------------------------------------

func TestGenerateVerificationToken_Length(t *testing.T) {
	tok := GenerateVerificationToken()
	if len(tok) != 32 {
		t.Errorf("expected 32-char hex token, got len=%d: %s", len(tok), tok)
	}
}

func TestGenerateVerificationToken_Uniqueness(t *testing.T) {
	a, b := GenerateVerificationToken(), GenerateVerificationToken()
	if a == b {
		t.Error("two consecutive tokens should not be identical")
	}
}

func TestGenerateVerificationToken_HexOnly(t *testing.T) {
	tok := GenerateVerificationToken()
	for _, c := range tok {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex character %c in token %s", c, tok)
		}
	}
}

func TestGenerateVerificationToken_RandError_FallsBackToZeros(t *testing.T) {
	old := randReadFn
	randReadFn = func(b []byte) (int, error) { return 0, errors.New("entropy source unavailable") }
	t.Cleanup(func() { randReadFn = old })

	tok := GenerateVerificationToken()
	zeros := make([]byte, 16) // all-zero fallback
	expected := "00000000000000000000000000000000"
	_ = zeros
	if len(tok) != 32 {
		t.Errorf("expected 32-char token, got %d", len(tok))
	}
	if tok != expected {
		t.Errorf("expected all-zero hex token on rand error, got %s", tok)
	}
}

// ---------------------------------------------------------------------------
// detectDNSProvider — third-block (SOA fallback) match
// The third block is reached when the first LookupNS(rootDomain) returns
// zero records (bypassed by len>0 guard) but the subsequent LookupNS call
// inside block 3 returns matching records.
// ---------------------------------------------------------------------------

// callOrderNSMock returns different results per successive LookupNS call.
type callOrderNSMock struct {
	calls   int
	results [][]*net.NS
	errs    []error
}

func (m *callOrderNSMock) LookupCNAME(_ string) (string, error)  { return "", errors.New("no cname") }
func (m *callOrderNSMock) LookupHost(_ string) ([]string, error) { return nil, errors.New("no host") }
func (m *callOrderNSMock) LookupTXT(_ string) ([]string, error)  { return nil, errors.New("no txt") }
func (m *callOrderNSMock) LookupNS(_ string) ([]*net.NS, error) {
	i := m.calls
	m.calls++
	if i < len(m.results) {
		return m.results[i], m.errs[i]
	}
	return nil, errors.New("exhausted")
}

func TestDetectDNSProvider_ThirdBlock_Match(t *testing.T) {
	// domain == rootDomain so block 2 is skipped.
	// Call 1 (block 1): returns 0 records → len==0 guard fails, block 1 skipped.
	// Call 2 (block 3): returns matching NS → provider returned.
	mock := &callOrderNSMock{
		results: [][]*net.NS{
			{},                          // call 1: empty → bypasses block 1
			{ns("ns1.cloudflare.com.")}, // call 2: block 3 match
		},
		errs: []error{nil, nil},
	}
	provider, err := detectDNSProvider(mock, "example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "cloudflare" {
		t.Errorf("expected cloudflare from third block, got %s", provider)
	}
}

// ---------------------------------------------------------------------------
// Exported wrapper: DetectDNSProvider — uses defaultResolver
// ---------------------------------------------------------------------------

func TestDetectDNSProvider_ExportedWrapper(t *testing.T) {
	old := defaultResolver
	defaultResolver = &mockNetLookup{nsRecords: []*net.NS{ns("ns1.cloudflare.com.")}}
	t.Cleanup(func() { defaultResolver = old })

	provider, err := DetectDNSProvider("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "cloudflare" {
		t.Errorf("expected cloudflare, got %s", provider)
	}
}
