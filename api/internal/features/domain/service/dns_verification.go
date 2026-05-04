package service

import (
	"fmt"
	"strings"
)

// verifyDNSConfiguration checks whether domain's DNS points to targetSubdomain
// via CNAME, matching A records, or a TXT ownership record.
func verifyDNSConfiguration(resolver NetLookup, domain, targetSubdomain string) (bool, error) {
	expectedTarget := fmt.Sprintf("%s.nixopus.ai.", targetSubdomain)

	cname, err := resolver.LookupCNAME(domain)
	if err == nil && strings.EqualFold(cname, expectedTarget) {
		return true, nil
	}

	hosts, err := resolver.LookupHost(domain)
	if err == nil {
		targetHosts, lookupErr := resolver.LookupHost(fmt.Sprintf("%s.nixopus.ai", targetSubdomain))
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
	txtRecords, err := resolver.LookupTXT(fmt.Sprintf("_nixopus-verify.%s", domain))
	if err == nil {
		for _, txt := range txtRecords {
			if strings.EqualFold(strings.TrimSpace(txt), expectedTXT) {
				return true, nil
			}
		}
	}

	return false, nil
}

// VerifyDNSConfiguration is the exported entry point backed by defaultResolver.
func VerifyDNSConfiguration(domain, targetSubdomain string) (bool, error) {
	return verifyDNSConfiguration(defaultResolver, domain, targetSubdomain)
}

// checkDNSPropagation returns the propagation state of the domain's DNS records.
func checkDNSPropagation(resolver NetLookup, domain string) (string, error) {
	cname, err := resolver.LookupCNAME(domain)
	if err == nil && cname != "" && cname != domain+"." {
		if strings.Contains(strings.ToLower(cname), "nixopus.ai") {
			return "verified", nil
		}
	}

	expectedTXT := fmt.Sprintf("nixopus-domain-verify=%s", domain)
	txtRecords, err := resolver.LookupTXT(fmt.Sprintf("_nixopus-verify.%s", domain))
	if err == nil {
		for _, txt := range txtRecords {
			if strings.EqualFold(strings.TrimSpace(txt), expectedTXT) {
				return "verified", nil
			}
		}
	}

	_, err = resolver.LookupHost(domain)
	if err != nil {
		return "not_configured", nil
	}

	return "propagating", nil
}

// CheckDNSPropagation is the exported entry point backed by defaultResolver.
func CheckDNSPropagation(domain string) (string, error) {
	return checkDNSPropagation(defaultResolver, domain)
}

// verifyARecordMatchesMachineIP checks that the domain resolves to machineIP.
func verifyARecordMatchesMachineIP(resolver NetLookup, domain, machineIP string) (bool, error) {
	if machineIP == "" {
		return false, fmt.Errorf("machine IP is not configured")
	}
	hosts, err := resolver.LookupHost(domain)
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

// VerifyARecordMatchesMachineIP is the exported entry point backed by defaultResolver.
func VerifyARecordMatchesMachineIP(domain, machineIP string) (bool, error) {
	return verifyARecordMatchesMachineIP(defaultResolver, domain, machineIP)
}

// checkDNSPropagationBYOS returns the propagation state for a BYOS domain that
// should resolve to machineIP via an A record.
func checkDNSPropagationBYOS(resolver NetLookup, domain, machineIP string) (string, error) {
	hosts, err := resolver.LookupHost(domain)
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

// CheckDNSPropagationBYOS is the exported entry point backed by defaultResolver.
func CheckDNSPropagationBYOS(domain, machineIP string) (string, error) {
	return checkDNSPropagationBYOS(defaultResolver, domain, machineIP)
}
