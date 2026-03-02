package tasks

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ParsedComposeService struct {
	ServiceName string
	Ports       []int
}

type composeFile struct {
	Services map[string]composeServiceDef `yaml:"services"`
}

type composeServiceDef struct {
	Ports  []interface{} `yaml:"ports"`
	Expose []interface{} `yaml:"expose"`
}

// ParseComposeFile reads a docker-compose YAML file and extracts service names
// and their published host ports.
func ParseComposeFile(filePath string) ([]ParsedComposeService, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	return ParseComposeYAML(data)
}

func ParseComposeYAML(data []byte) ([]ParsedComposeService, error) {
	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("failed to parse compose YAML: %w", err)
	}

	var result []ParsedComposeService
	for name, svc := range cf.Services {
		ports := extractPorts(svc.Ports)
		if len(ports) == 0 {
			ports = extractExposePorts(svc.Expose)
		}
		result = append(result, ParsedComposeService{
			ServiceName: name,
			Ports:       ports,
		})
	}

	return result, nil
}

// extractExposePorts handles the expose directive which lists container ports
// (e.g. expose: ["3000"] or expose: [3000]).
func extractExposePorts(raw []interface{}) []int {
	var ports []int
	seen := make(map[int]bool)

	for _, entry := range raw {
		var p int
		switch v := entry.(type) {
		case string:
			p, _ = strconv.Atoi(strings.TrimSpace(v))
		case int:
			p = v
		case float64:
			p = int(v)
		}
		if p > 0 && p <= 65535 && !seen[p] {
			seen[p] = true
			ports = append(ports, p)
		}
	}

	return ports
}

// extractPorts handles both short syntax ("8080:80", "3000") and long syntax
// (map with target/published keys) from docker-compose port definitions.
func extractPorts(raw []interface{}) []int {
	var ports []int
	seen := make(map[int]bool)

	for _, entry := range raw {
		var extracted []int

		switch v := entry.(type) {
		case string:
			extracted = parseShortPortSyntax(v)
		case int:
			extracted = []int{v}
		case float64:
			extracted = []int{int(v)}
		case map[string]interface{}:
			extracted = parseLongPortSyntax(v)
		}

		for _, p := range extracted {
			if p > 0 && p <= 65535 && !seen[p] {
				seen[p] = true
				ports = append(ports, p)
			}
		}
	}

	return ports
}

// parseShortPortSyntax handles formats like:
// "8080:80" -> published=8080, "3000" -> published=3000,
// "127.0.0.1:8080:80" -> published=8080, "8080-8090:80-90" -> published=8080
func parseShortPortSyntax(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Strip protocol suffix like "/tcp" or "/udp"
	if idx := strings.LastIndex(s, "/"); idx != -1 {
		s = s[:idx]
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		// Just a port number or range
		p := parsePortOrRangeFirst(parts[0])
		if p > 0 {
			return []int{p}
		}
	case 2:
		// host_port:container_port
		p := parsePortOrRangeFirst(parts[0])
		if p > 0 {
			return []int{p}
		}
	case 3:
		// ip:host_port:container_port
		p := parsePortOrRangeFirst(parts[1])
		if p > 0 {
			return []int{p}
		}
	}

	return nil
}

func parsePortOrRangeFirst(s string) int {
	if idx := strings.Index(s, "-"); idx != -1 {
		s = s[:idx]
	}
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return p
}

// parseLongPortSyntax handles the long form: {target: 80, published: 8080, protocol: tcp}
func parseLongPortSyntax(m map[string]interface{}) []int {
	if published, ok := m["published"]; ok {
		switch v := published.(type) {
		case int:
			return []int{v}
		case float64:
			return []int{int(v)}
		case string:
			p := parsePortOrRangeFirst(v)
			if p > 0 {
				return []int{p}
			}
		}
	}

	// Fall back to target port if no published port
	if target, ok := m["target"]; ok {
		switch v := target.(type) {
		case int:
			return []int{v}
		case float64:
			return []int{int(v)}
		}
	}

	return nil
}
