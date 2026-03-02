package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
)

var nsProviderMap = map[string]string{
	"cloudflare.com":        "cloudflare",
	"ns.cloudflare.com":     "cloudflare",
	"awsdns":                "route53",
	"amazonaws.com":         "route53",
	"domaincontrol.com":     "godaddy",
	"registrar-servers.com": "namecheap",
	"googledomains.com":     "google",
	"google.com":            "google",
	"dns.he.net":            "hurricane_electric",
	"vercel-dns.com":        "vercel",
	"netlify.com":           "netlify",
	"digitalocean.com":      "digitalocean",
	"dnsmadeeasy.com":       "dnsmadeeasy",
	"linode.com":            "linode",
	"vultr.com":             "vultr",
	"hetzner.com":           "hetzner",
	"ovh.net":               "ovh",
	"gandi.net":             "gandi",
	"hostgator.com":         "hostgator",
	"bluehost.com":          "bluehost",
	"siteground.net":        "siteground",
	"dreamhost.com":         "dreamhost",
	"name-services.com":     "enom",
	"dynect.net":            "dyn",
	"azure-dns.com":         "azure",
	"azure-dns.net":         "azure",
	"azure-dns.org":         "azure",
	"azure-dns.info":        "azure",
	"nsone.net":             "ns1",
	"ultradns.com":          "ultradns",
	"ultradns.net":          "ultradns",
	"dnsimple.com":          "dnsimple",
	"hover.com":             "hover",
	"porkbun.com":           "porkbun",
	"bunny.net":             "bunnycdn",
}

func extractRootDomain(domain string) string {
	domain = strings.TrimSuffix(strings.TrimSpace(domain), ".")
	parts := strings.Split(domain, ".")
	if len(parts) <= 2 {
		return domain
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

func matchNSToProvider(nsRecords []*net.NS) string {
	for _, ns := range nsRecords {
		host := strings.ToLower(strings.TrimSuffix(ns.Host, "."))
		for pattern, provider := range nsProviderMap {
			if strings.Contains(host, pattern) {
				return provider
			}
		}
	}
	return ""
}

func DetectDNSProvider(domain string) (string, error) {
	rootDomain := extractRootDomain(domain)

	if nsRecords, err := net.LookupNS(rootDomain); err == nil && len(nsRecords) > 0 {
		if provider := matchNSToProvider(nsRecords); provider != "" {
			return provider, nil
		}
	}

	if domain != rootDomain {
		if nsRecords, err := net.LookupNS(domain); err == nil && len(nsRecords) > 0 {
			if provider := matchNSToProvider(nsRecords); provider != "" {
				return provider, nil
			}
		}
	}

	if soaRecords, err := net.LookupNS(rootDomain); err == nil {
		for _, ns := range soaRecords {
			host := strings.ToLower(strings.TrimSuffix(ns.Host, "."))
			for pattern, provider := range nsProviderMap {
				if strings.Contains(host, pattern) {
					return provider, nil
				}
			}
		}
	}

	return "other", nil
}

func GenerateDNSInstructions(domain, targetSubdomain, provider string) []types.DNSInstruction {
	providerDescriptions := map[string]string{
		"cloudflare":         "Go to your Cloudflare dashboard > DNS > Records > Add Record",
		"route53":            "Go to AWS Route 53 > Hosted Zones > select your domain > Create Record",
		"godaddy":            "Go to GoDaddy > My Products > DNS > Add Record",
		"namecheap":          "Go to Namecheap > Domain List > Manage > Advanced DNS > Add New Record",
		"google":             "Go to Google Domains > DNS > Custom Records > Manage",
		"hurricane_electric": "Go to dns.he.net > Edit Zone > Add Record",
		"vercel":             "Go to Vercel > Settings > Domains > Add Record",
		"netlify":            "Go to Netlify > Domain Settings > DNS Records > Add Record",
		"digitalocean":       "Go to DigitalOcean > Networking > Domains > your domain > Add Record",
		"azure":              "Go to Azure Portal > DNS Zones > your domain > + Record set",
		"vultr":              "Go to Vultr > DNS > your domain > Add Record",
		"hetzner":            "Go to Hetzner DNS Console > your domain > Add Record",
		"ovh":                "Go to OVH Manager > Domains > your domain > DNS Zone > Add Record",
		"gandi":              "Go to Gandi > Domains > your domain > DNS Records > Add Record",
		"porkbun":            "Go to Porkbun > Domain Management > your domain > DNS Records > Add Record",
		"ns1":                "Go to NS1 > Zones > your domain > Add Record",
		"dnsimple":           "Go to DNSimple > your domain > DNS > Add Record",
		"linode":             "Go to Linode Cloud Manager > Domains > your domain > Add Record",
		"other":              "Go to your DNS provider's dashboard and add the following records",
	}

	description, ok := providerDescriptions[provider]
	if !ok {
		description = providerDescriptions["other"]
	}

	cnameValue := fmt.Sprintf("%s.nixopus.ai", targetSubdomain)

	instructions := []types.DNSInstruction{
		{
			RecordType:  "CNAME",
			Name:        "*",
			Value:       cnameValue,
			Description: fmt.Sprintf("%s. Add a CNAME record pointing * to %s", description, cnameValue),
		},
		{
			RecordType:  "TXT",
			Name:        "_nixopus-verify",
			Value:       fmt.Sprintf("nixopus-domain-verify=%s", domain),
			Description: fmt.Sprintf("%s. Add a TXT record for domain verification", description),
		},
	}

	return instructions
}

func GenerateVerificationToken() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return hex.EncodeToString(make([]byte, 16))
	}
	return hex.EncodeToString(bytes)
}
