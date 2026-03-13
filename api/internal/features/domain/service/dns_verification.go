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
