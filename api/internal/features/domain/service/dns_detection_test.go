package service

import (
	"testing"
)

func TestGenerateDNSInstructionsBYOS_IPv4(t *testing.T) {
	instructions := GenerateDNSInstructionsBYOS("example.com", "93.184.216.34", "cloudflare")
	if len(instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(instructions))
	}
	inst := instructions[0]
	if inst.RecordType != "A" {
		t.Errorf("expected record type A for IPv4, got %s", inst.RecordType)
	}
	if inst.Name != "example.com" {
		t.Errorf("expected name example.com, got %s", inst.Name)
	}
	if inst.Value != "93.184.216.34" {
		t.Errorf("expected value 93.184.216.34, got %s", inst.Value)
	}
}

func TestGenerateDNSInstructionsBYOS_IPv6(t *testing.T) {
	instructions := GenerateDNSInstructionsBYOS("example.com", "2001:db8::1", "cloudflare")
	if len(instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(instructions))
	}
	inst := instructions[0]
	if inst.RecordType != "AAAA" {
		t.Errorf("expected record type AAAA for IPv6, got %s", inst.RecordType)
	}
	if inst.Value != "2001:db8::1" {
		t.Errorf("expected value 2001:db8::1, got %s", inst.Value)
	}
}

func TestGenerateDNSInstructionsBYOS_UnknownProvider(t *testing.T) {
	instructions := GenerateDNSInstructionsBYOS("example.com", "10.0.0.1", "unknownprovider")
	if len(instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(instructions))
	}
	inst := instructions[0]
	if inst.RecordType != "A" {
		t.Errorf("expected record type A, got %s", inst.RecordType)
	}
	if inst.Description == "" {
		t.Error("expected non-empty description for unknown provider")
	}
}

func TestGenerateDNSInstructionsBYOS_EmptyIP(t *testing.T) {
	instructions := GenerateDNSInstructionsBYOS("example.com", "", "cloudflare")
	if len(instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(instructions))
	}
	if instructions[0].RecordType != "A" {
		t.Errorf("expected A for empty/unparseable IP, got %s", instructions[0].RecordType)
	}
}

func TestGenerateDNSInstructionsBYOS_IPv4MappedIPv6(t *testing.T) {
	// ::ffff:192.0.2.1 is IPv4-mapped IPv6; To4() returns non-nil so it should be A
	instructions := GenerateDNSInstructionsBYOS("example.com", "::ffff:192.0.2.1", "other")
	if len(instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(instructions))
	}
	if instructions[0].RecordType != "A" {
		t.Errorf("expected A for IPv4-mapped IPv6 address, got %s", instructions[0].RecordType)
	}
}
