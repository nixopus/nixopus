package service

import (
	"fmt"
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// verifyDNSConfiguration checks whether domain's DNS points to targetSubdomain
// via CNAME, matching A records, or a TXT ownership record.
func verifyDNSConfiguration(resolver NetLookup, domain, targetSubdomain string, log *logger.Logger) (bool, error) {
	dnsLog(log, logger.Debug, "dns: verify config start", fmt.Sprintf("domain=%s target=%s", domain, targetSubdomain))
	expectedTarget := fmt.Sprintf("%s.nixopus.ai.", targetSubdomain)

	cname, err := resolver.LookupCNAME(domain)
	if err == nil && strings.EqualFold(cname, expectedTarget) {
		dnsLog(log, logger.Debug, "dns: verify config ok", fmt.Sprintf("domain=%s match=cname", domain))
		return true, nil
	}

	hosts, err := resolver.LookupHost(domain)
	if err == nil {
		targetHosts, lookupErr := resolver.LookupHost(fmt.Sprintf("%s.nixopus.ai", targetSubdomain))
		if lookupErr == nil {
			for _, h := range hosts {
				for _, th := range targetHosts {
					if h == th {
						dnsLog(log, logger.Debug, "dns: verify config ok", fmt.Sprintf("domain=%s match=a_same_ip", domain))
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
				dnsLog(log, logger.Debug, "dns: verify config ok", fmt.Sprintf("domain=%s match=txt", domain))
				return true, nil
			}
		}
	}

	dnsLog(log, logger.Debug, "dns: verify config no match", fmt.Sprintf("domain=%s target=%s", domain, targetSubdomain))
	return false, nil
}

// VerifyDNSConfiguration is the exported entry point backed by defaultResolver.
func VerifyDNSConfiguration(domain, targetSubdomain string) (bool, error) {
	return verifyDNSConfiguration(defaultResolver, domain, targetSubdomain, nil)
}

// checkDNSPropagation returns the propagation state of the domain's DNS records.
func checkDNSPropagation(resolver NetLookup, domain string, log *logger.Logger) (string, error) {
	dnsLog(log, logger.Debug, "dns: check propagation start", fmt.Sprintf("domain=%s", domain))
	cname, err := resolver.LookupCNAME(domain)
	if err == nil && cname != "" && cname != domain+"." {
		if strings.Contains(strings.ToLower(cname), "nixopus.ai") {
			dnsLog(log, logger.Debug, "dns: check propagation done", "domain="+domain+" status=verified via=cname")
			return "verified", nil
		}
	}

	expectedTXT := fmt.Sprintf("nixopus-domain-verify=%s", domain)
	txtRecords, err := resolver.LookupTXT(fmt.Sprintf("_nixopus-verify.%s", domain))
	if err == nil {
		for _, txt := range txtRecords {
			if strings.EqualFold(strings.TrimSpace(txt), expectedTXT) {
				dnsLog(log, logger.Debug, "dns: check propagation done", "domain="+domain+" status=verified via=txt")
				return "verified", nil
			}
		}
	}

	_, err = resolver.LookupHost(domain)
	if err != nil {
		dnsLog(log, logger.Debug, "dns: check propagation done", "domain="+domain+" status=not_configured")
		return "not_configured", nil
	}

	dnsLog(log, logger.Debug, "dns: check propagation done", "domain="+domain+" status=propagating")
	return "propagating", nil
}

// CheckDNSPropagation is the exported entry point backed by defaultResolver.
func CheckDNSPropagation(domain string) (string, error) {
	return checkDNSPropagation(defaultResolver, domain, nil)
}

// verifyARecordMatchesMachineIP checks that the domain resolves to machineIP.
func verifyARecordMatchesMachineIP(resolver NetLookup, domain, machineIP string, log *logger.Logger) (bool, error) {
	dnsLog(log, logger.Debug, "dns: verify A record start", fmt.Sprintf("domain=%s", domain))
	if machineIP == "" {
		dnsLog(log, logger.Error, "dns: verify A record", "machine IP is not configured")
		return false, fmt.Errorf("machine IP is not configured")
	}
	hosts, err := resolver.LookupHost(domain)
	if err != nil {
		dnsLog(log, logger.Debug, "dns: verify A record lookup failed", fmt.Sprintf("domain=%s err=%v", domain, err))
		return false, nil
	}
	for _, h := range hosts {
		if h == machineIP {
			dnsLog(log, logger.Debug, "dns: verify A record ok", fmt.Sprintf("domain=%s ip=%s", domain, machineIP))
			return true, nil
		}
	}
	dnsLog(log, logger.Debug, "dns: verify A record no match", fmt.Sprintf("domain=%s expected_ip=%s got=%v", domain, machineIP, hosts))
	return false, nil
}

// VerifyARecordMatchesMachineIP is the exported entry point backed by defaultResolver.
func VerifyARecordMatchesMachineIP(domain, machineIP string) (bool, error) {
	return verifyARecordMatchesMachineIP(defaultResolver, domain, machineIP, nil)
}

// checkDNSPropagationBYOS returns the propagation state for a BYOS domain that
// should resolve to machineIP via an A record.
func checkDNSPropagationBYOS(resolver NetLookup, domain, machineIP string, log *logger.Logger) (string, error) {
	dnsLog(log, logger.Debug, "dns: check propagation BYOS start", fmt.Sprintf("domain=%s", domain))
	hosts, err := resolver.LookupHost(domain)
	if err != nil {
		dnsLog(log, logger.Debug, "dns: check propagation BYOS done", fmt.Sprintf("domain=%s status=not_configured err=%v", domain, err))
		return "not_configured", nil
	}
	for _, h := range hosts {
		if h == machineIP {
			dnsLog(log, logger.Debug, "dns: check propagation BYOS done", fmt.Sprintf("domain=%s status=verified ip=%s", domain, machineIP))
			return "verified", nil
		}
	}
	dnsLog(log, logger.Debug, "dns: check propagation BYOS done", fmt.Sprintf("domain=%s status=propagating hosts=%v", domain, hosts))
	return "propagating", nil
}

// CheckDNSPropagationBYOS is the exported entry point backed by defaultResolver.
func CheckDNSPropagationBYOS(domain, machineIP string) (string, error) {
	return checkDNSPropagationBYOS(defaultResolver, domain, machineIP, nil)
}
