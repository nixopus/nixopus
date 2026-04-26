package service

import (
	"fmt"
	"net"
	"strings"
)

func VerifyDNSConfiguration(domain, targetSubdomain string) (bool, error) {
	expectedTarget := fmt.Sprintf("%s.nixopus.ai.", targetSubdomain)

	cname, err := net.LookupCNAME(domain)
	if err == nil && strings.EqualFold(cname, expectedTarget) {
		return true, nil
	}

	hosts, err := net.LookupHost(domain)
	if err == nil {
		targetHosts, lookupErr := net.LookupHost(fmt.Sprintf("%s.nixopus.ai", targetSubdomain))
		if lookupErr == nil {
			for _, h := range hosts {
				for _, th := range targetHosts {
					if h == th {
						return true, nil
					}
				}
			}
		}
	}

	expectedTXT := fmt.Sprintf("nixopus-domain-verify=%s", domain)
	txtRecords, err := net.LookupTXT(fmt.Sprintf("_nixopus-verify.%s", domain))
	if err == nil {
		for _, txt := range txtRecords {
			if strings.EqualFold(strings.TrimSpace(txt), expectedTXT) {
				return true, nil
			}
		}
	}

	return false, nil
}

func CheckDNSPropagation(domain string) (string, error) {
	cname, err := net.LookupCNAME(domain)
	if err == nil && cname != "" && cname != domain+"." {
		if strings.Contains(strings.ToLower(cname), "nixopus.ai") {
			return "verified", nil
		}
	}

	expectedTXT := fmt.Sprintf("nixopus-domain-verify=%s", domain)
	txtRecords, err := net.LookupTXT(fmt.Sprintf("_nixopus-verify.%s", domain))
	if err == nil {
		for _, txt := range txtRecords {
			if strings.EqualFold(strings.TrimSpace(txt), expectedTXT) {
				return "verified", nil
			}
		}
	}

	_, err = net.LookupHost(domain)
	if err != nil {
		return "not_configured", nil
	}

	return "propagating", nil
}

// VerifyDNSRecordMatchesMachineIP checks that the domain resolves to machineIP.
// Works for both IPv4 (A) and IPv6 (AAAA) records since LookupHost returns both.
// Used for BYOS machines where traffic goes directly to the user's server.
func VerifyDNSRecordMatchesMachineIP(domain, machineIP string) (bool, error) {
	if machineIP == "" {
		return false, fmt.Errorf("machine IP is not configured")
	}
	hosts, err := net.LookupHost(domain)
	if err != nil {
		return false, nil
	}
	for _, h := range hosts {
		if h == machineIP {
			return true, nil
		}
	}
	return false, nil
}

// CheckDNSPropagationBYOS checks propagation state for a BYOS domain
// that should resolve to machineIP via an A or AAAA record.
func CheckDNSPropagationBYOS(domain, machineIP string) (string, error) {
	hosts, err := net.LookupHost(domain)
	if err != nil {
		return "not_configured", nil
	}
	for _, h := range hosts {
		if h == machineIP {
			return "verified", nil
		}
	}
	return "propagating", nil
}
