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
	if err != nil {
		return false, nil
	}

	targetHosts, err := net.LookupHost(fmt.Sprintf("%s.nixopus.ai", targetSubdomain))
	if err != nil {
		return false, nil
	}

	for _, h := range hosts {
		for _, th := range targetHosts {
			if h == th {
				return true, nil
			}
		}
	}

	return false, nil
}

func CheckDNSPropagation(domain string) (string, error) {
	_, err := net.LookupHost(domain)
	if err != nil {
		return "not_configured", nil
	}

	cname, err := net.LookupCNAME(domain)
	if err != nil || cname == "" || cname == domain+"." {
		return "propagating", nil
	}

	if strings.Contains(strings.ToLower(cname), "nixopus.ai") {
		return "verified", nil
	}

	return "propagating", nil
}
