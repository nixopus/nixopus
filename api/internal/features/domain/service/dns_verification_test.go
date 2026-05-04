package service

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// verifyDNSConfiguration
// ---------------------------------------------------------------------------

func TestVerifyDNSConfiguration_CNAMEMatch(t *testing.T) {
	mock := &mockNetLookup{cname: "myorg.nixopus.ai."}
	ok, err := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified via CNAME match")
	}
}

func TestVerifyDNSConfiguration_CNAMECaseInsensitive(t *testing.T) {
	mock := &mockNetLookup{cname: "MYORG.NIXOPUS.AI."}
	ok, _ := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if !ok {
		t.Error("CNAME match should be case-insensitive")
	}
}

func TestVerifyDNSConfiguration_HostIPMatch(t *testing.T) {
	// CNAME doesn't match; A records of domain and target subdomain share an IP.
	mock := &mockNetLookup{
		cname:    "other.example.com.",
		cnameErr: nil,
		hostsByName: map[string][]string{
			"app.example.com":  {"1.2.3.4"},
			"myorg.nixopus.ai": {"1.2.3.4"},
		},
	}
	ok, err := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified via matching A records")
	}
}

func TestVerifyDNSConfiguration_HostIPNoMatch(t *testing.T) {
	mock := &mockNetLookup{
		cnameErr: errors.New("no cname"),
		hostsByName: map[string][]string{
			"app.example.com":  {"1.2.3.4"},
			"myorg.nixopus.ai": {"5.6.7.8"},
		},
	}
	ok, err := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected not verified when A records differ")
	}
}

func TestVerifyDNSConfiguration_TXTMatch(t *testing.T) {
	mock := &mockNetLookup{
		cnameErr:   errors.New("no cname"),
		hostErr:    errors.New("no host"),
		txtRecords: []string{"nixopus-domain-verify=app.example.com"},
	}
	ok, err := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified via TXT record")
	}
}

func TestVerifyDNSConfiguration_TXTMatchCaseInsensitive(t *testing.T) {
	mock := &mockNetLookup{
		cnameErr:   errors.New("no cname"),
		hostErr:    errors.New("no host"),
		txtRecords: []string{"  NIXOPUS-DOMAIN-VERIFY=app.example.com  "},
	}
	ok, _ := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if !ok {
		t.Error("TXT match should be case-insensitive and trim spaces")
	}
}

func TestVerifyDNSConfiguration_NothingMatches(t *testing.T) {
	mock := &mockNetLookup{
		cnameErr: errors.New("no cname"),
		hostErr:  errors.New("no host"),
		txtErr:   errors.New("no txt"),
	}
	ok, err := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected not verified when nothing resolves")
	}
}

func TestVerifyDNSConfiguration_TargetHostLookupFails(t *testing.T) {
	// Domain resolves but target subdomain lookup fails → no IP comparison possible.
	mock := &mockNetLookup{
		cnameErr: errors.New("no cname"),
		hostsByName: map[string][]string{
			"app.example.com": {"1.2.3.4"},
		},
		hostErrByName: map[string]error{
			"myorg.nixopus.ai": errors.New("no such host"),
		},
		txtErr: errors.New("no txt"),
	}
	ok, err := verifyDNSConfiguration(mock, "app.example.com", "myorg", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected not verified when target host lookup fails")
	}
}

// ---------------------------------------------------------------------------
// checkDNSPropagation
// ---------------------------------------------------------------------------

func TestCheckDNSPropagation_CNAMEVerified(t *testing.T) {
	mock := &mockNetLookup{cname: "myorg.nixopus.ai."}
	status, err := checkDNSPropagation(mock, "app.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "verified" {
		t.Errorf("expected verified, got %s", status)
	}
}

func TestCheckDNSPropagation_CNAMESelfPointing(t *testing.T) {
	// CNAME points to itself — should not count as verified.
	mock := &mockNetLookup{
		cname:  "app.example.com.",
		hosts:  []string{"1.2.3.4"},
		txtErr: errors.New("no txt"),
	}
	status, _ := checkDNSPropagation(mock, "app.example.com", nil)
	if status != "propagating" {
		t.Errorf("self-pointing CNAME should be propagating, got %s", status)
	}
}

func TestCheckDNSPropagation_TXTVerified(t *testing.T) {
	mock := &mockNetLookup{
		cnameErr:   errors.New("no cname"),
		txtRecords: []string{"nixopus-domain-verify=app.example.com"},
		hosts:      []string{"1.2.3.4"},
	}
	status, err := checkDNSPropagation(mock, "app.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "verified" {
		t.Errorf("expected verified via TXT, got %s", status)
	}
}

func TestCheckDNSPropagation_NotConfigured(t *testing.T) {
	mock := &mockNetLookup{
		cnameErr: errors.New("no cname"),
		txtErr:   errors.New("no txt"),
		hostErr:  errors.New("no host"),
	}
	status, err := checkDNSPropagation(mock, "app.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "not_configured" {
		t.Errorf("expected not_configured, got %s", status)
	}
}

func TestCheckDNSPropagation_Propagating(t *testing.T) {
	// Host resolves but no CNAME/TXT matching nixopus.
	mock := &mockNetLookup{
		cnameErr: errors.New("no cname"),
		txtErr:   errors.New("no txt"),
		hosts:    []string{"1.2.3.4"},
	}
	status, err := checkDNSPropagation(mock, "app.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "propagating" {
		t.Errorf("expected propagating, got %s", status)
	}
}

func TestCheckDNSPropagation_CNAMEDoesNotContainNixopus(t *testing.T) {
	mock := &mockNetLookup{
		cname:   "other.provider.com.",
		txtErr:  errors.New("no txt"),
		hostErr: errors.New("no host"),
	}
	status, _ := checkDNSPropagation(mock, "app.example.com", nil)
	if status != "not_configured" {
		t.Errorf("CNAME without nixopus.ai should not be verified, got %s", status)
	}
}

// ---------------------------------------------------------------------------
// verifyARecordMatchesMachineIP
// ---------------------------------------------------------------------------

func TestVerifyARecord_EmptyMachineIP(t *testing.T) {
	mock := &mockNetLookup{}
	ok, err := verifyARecordMatchesMachineIP(mock, "app.example.com", "", nil)
	if err == nil {
		t.Fatal("expected error for empty machineIP")
	}
	if ok {
		t.Error("expected false for empty machineIP")
	}
}

func TestVerifyARecord_LookupFails(t *testing.T) {
	mock := &mockNetLookup{hostErr: errors.New("nxdomain")}
	ok, err := verifyARecordMatchesMachineIP(mock, "app.example.com", "1.2.3.4", nil)
	if err != nil {
		t.Fatalf("lookup failure should not produce an error, got %v", err)
	}
	if ok {
		t.Error("expected false when lookup fails")
	}
}

func TestVerifyARecord_Match(t *testing.T) {
	mock := &mockNetLookup{hosts: []string{"1.2.3.4", "5.6.7.8"}}
	ok, err := verifyARecordMatchesMachineIP(mock, "app.example.com", "5.6.7.8", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified when IP is in resolved hosts")
	}
}

func TestVerifyARecord_NoMatch(t *testing.T) {
	mock := &mockNetLookup{hosts: []string{"1.2.3.4"}}
	ok, err := verifyARecordMatchesMachineIP(mock, "app.example.com", "9.9.9.9", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false when IP is not in resolved hosts")
	}
}

// ---------------------------------------------------------------------------
// checkDNSPropagationBYOS
// ---------------------------------------------------------------------------

func TestCheckDNSPropagationBYOS_LookupFails(t *testing.T) {
	mock := &mockNetLookup{hostErr: errors.New("nxdomain")}
	status, err := checkDNSPropagationBYOS(mock, "app.example.com", "1.2.3.4", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "not_configured" {
		t.Errorf("expected not_configured, got %s", status)
	}
}

func TestCheckDNSPropagationBYOS_Match(t *testing.T) {
	mock := &mockNetLookup{hosts: []string{"1.2.3.4"}}
	status, err := checkDNSPropagationBYOS(mock, "app.example.com", "1.2.3.4", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "verified" {
		t.Errorf("expected verified, got %s", status)
	}
}

func TestCheckDNSPropagationBYOS_NoMatch(t *testing.T) {
	mock := &mockNetLookup{hosts: []string{"5.5.5.5"}}
	status, err := checkDNSPropagationBYOS(mock, "app.example.com", "1.2.3.4", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "propagating" {
		t.Errorf("expected propagating, got %s", status)
	}
}

// ---------------------------------------------------------------------------
// Exported wrappers — each swaps defaultResolver so no real network calls
// ---------------------------------------------------------------------------

func TestVerifyDNSConfiguration_ExportedWrapper(t *testing.T) {
	old := defaultResolver
	defaultResolver = &mockNetLookup{cname: "myorg.nixopus.ai."}
	t.Cleanup(func() { defaultResolver = old })

	ok, err := VerifyDNSConfiguration("app.example.com", "myorg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified")
	}
}

func TestCheckDNSPropagation_ExportedWrapper(t *testing.T) {
	old := defaultResolver
	defaultResolver = &mockNetLookup{cname: "myorg.nixopus.ai."}
	t.Cleanup(func() { defaultResolver = old })

	status, err := CheckDNSPropagation("app.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "verified" {
		t.Errorf("expected verified, got %s", status)
	}
}

func TestVerifyARecordMatchesMachineIP_ExportedWrapper(t *testing.T) {
	old := defaultResolver
	defaultResolver = &mockNetLookup{hosts: []string{"1.2.3.4"}}
	t.Cleanup(func() { defaultResolver = old })

	ok, err := VerifyARecordMatchesMachineIP("app.example.com", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verified")
	}
}

func TestCheckDNSPropagationBYOS_ExportedWrapper(t *testing.T) {
	old := defaultResolver
	defaultResolver = &mockNetLookup{hosts: []string{"1.2.3.4"}}
	t.Cleanup(func() { defaultResolver = old })

	status, err := CheckDNSPropagationBYOS("app.example.com", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "verified" {
		t.Errorf("expected verified, got %s", status)
	}
}
