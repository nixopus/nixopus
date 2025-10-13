package engine

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	deploydocker "github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type dockerModule struct{}

func (dockerModule) Type() string { return "docker" }

func (dockerModule) Execute(_ *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	action, _ := step.Properties["action"].(string)
	name, _ := step.Properties["name"].(string)
	image, _ := step.Properties["image"].(string)
	tag, _ := step.Properties["tag"].(string)
	ports, _ := step.Properties["ports"].(string)
	restart, _ := step.Properties["restart"].(string)
	cmdStr, _ := step.Properties["cmd"].(string)
	envAny, _ := step.Properties["env"]
	volumesAny, _ := step.Properties["volumes"]
	networksAny, _ := step.Properties["networks"]

	if image != "" {
		image = replaceVars(image, vars)
	}
	if tag != "" {
		tag = replaceVars(tag, vars)
	}
	if name != "" {
		name = replaceVars(name, vars)
	}
	if ports != "" {
		ports = replaceVars(ports, vars)
	}
	if cmdStr != "" {
		cmdStr = replaceVars(cmdStr, vars)
	}

	if action == "" {
		return "", nil, fmt.Errorf("docker: action is required (name=%q image=%q tag=%q)", name, image, tag)
	}

	svc := deploydocker.NewDockerService()

	type handler func() (string, func(), error)
	handlers := map[string]handler{
		"pull": func() (string, func(), error) { return dockerPull(svc, image, tag) },
		"run": func() (string, func(), error) {
			return dockerRun(svc, name, image, tag, ports, restart, cmdStr, envAny, volumesAny, networksAny, vars)
		},
		"stop":  func() (string, func(), error) { return dockerStop(svc, name) },
		"start": func() (string, func(), error) { return dockerStart(svc, name) },
		"rm":    func() (string, func(), error) { return dockerRm(svc, name) },
	}

	h, ok := handlers[action]
	if !ok {
		return "", nil, fmt.Errorf("docker: unsupported action %q (name=%q image=%q tag=%q)", action, name, image, tag)
	}
	out, comp, err := h()
	if err != nil {
		return "", nil, err
	}
	return strings.TrimSpace(out), comp, nil
}

func parsePortMappings(mappings string) (nat.PortMap, error) {
	pm := nat.PortMap{}
	for _, m := range strings.Split(mappings, ",") {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		parts := strings.Split(m, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port mapping: %s", m)
		}
		host := parts[0]
		p, err := nat.NewPort("tcp", parts[1])
		if err != nil {
			return nil, err
		}
		pm[p] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: host}}
	}
	return pm, nil
}

func exposedFromBindings(pm nat.PortMap) nat.PortSet {
	es := nat.PortSet{}
	for p := range pm {
		es[p] = struct{}{}
	}
	return es
}

func normalizeStringList(v interface{}, vars map[string]interface{}) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(replaceVars(t, vars))
		if s == "" {
			return nil, nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out, nil
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, it := range t {
			s, ok := it.(string)
			if !ok {
				return nil, fmt.Errorf("invalid list item type, expected string")
			}
			s = strings.TrimSpace(replaceVars(s, vars))
			if s != "" {
				out = append(out, s)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported list type")
	}
}

func normalizeEnv(v interface{}, vars map[string]interface{}) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case map[string]interface{}:
		out := make([]string, 0, len(t))
		for k, val := range t {
			vs := fmt.Sprint(val)
			out = append(out, fmt.Sprintf("%s=%s", k, replaceVars(vs, vars)))
		}
		return out, nil
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, it := range t {
			s, ok := it.(string)
			if !ok {
				return nil, fmt.Errorf("invalid env item type, expected string")
			}
			out = append(out, replaceVars(strings.TrimSpace(s), vars))
		}
		return out, nil
	case string:
		s := strings.TrimSpace(replaceVars(t, vars))
		if s == "" {
			return nil, nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported env type")
	}
}

func findContainerIDByName(svc *deploydocker.DockerService, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	list, err := svc.ListContainers(container.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, c := range list {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return c.ID, nil
			}
		}
	}
	return "", fmt.Errorf("container not found: %s", name)
}

func init() {
	RegisterModule(dockerModule{})
}

func dockerPull(svc *deploydocker.DockerService, image string, tag string) (string, func(), error) {
	if image == "" {
		return "", nil, fmt.Errorf("docker: image is required for pull")
	}
	ref := image
	if tag != "" {
		ref = fmt.Sprintf("%s:%s", image, tag)
	}
	r, err := svc.Cli.ImagePull(context.Background(), ref, imagetypes.PullOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("docker: image pull failed ref=%q: %w", ref, err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	return string(b), nil, nil
}

func dockerRun(
	svc *deploydocker.DockerService,
	name string,
	image string,
	tag string,
	ports string,
	restart string,
	cmdStr string,
	envAny interface{},
	volumesAny interface{},
	networksAny interface{},
	vars map[string]interface{},
) (string, func(), error) {
	if image == "" || name == "" {
		return "", nil, fmt.Errorf("docker: run requires image and name (name=%q image=%q tag=%q)", name, image, tag)
	}
	ref := image
	if tag != "" {
		ref = fmt.Sprintf("%s:%s", image, tag)
	}
	containerCfg := container.Config{Image: ref}
	hostCfg := container.HostConfig{}
	networkingCfg := network.NetworkingConfig{}
	if ports != "" {
		pm, err := parsePortMappings(ports)
		if err != nil {
			return "", nil, err
		}
		hostCfg.PortBindings = pm
		containerCfg.ExposedPorts = exposedFromBindings(pm)
	}
	if restart != "" {
		hostCfg.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(restart)}
	}
	// env
	if envList, err := normalizeEnv(envAny, vars); err != nil {
		return "", nil, err
	} else if len(envList) > 0 {
		containerCfg.Env = envList
	}

	// volumes (binds)
	if binds, err := normalizeStringList(volumesAny, vars); err != nil {
		return "", nil, err
	} else if len(binds) > 0 {
		hostCfg.Binds = binds
	}

	// networks
	if nets, err := normalizeStringList(networksAny, vars); err != nil {
		return "", nil, err
	} else if len(nets) > 0 {
		if networkingCfg.EndpointsConfig == nil {
			networkingCfg.EndpointsConfig = map[string]*network.EndpointSettings{}
		}
		for _, n := range nets {
			networkingCfg.EndpointsConfig[n] = &network.EndpointSettings{}
		}
	}

	// command
	if strings.TrimSpace(cmdStr) != "" {
		containerCfg.Cmd = strings.Fields(cmdStr)
	}

	resp, err := svc.CreateContainer(containerCfg, hostCfg, networkingCfg, name)
	if err != nil {
		return "", nil, fmt.Errorf("docker: create container failed name=%q image=%q: %w", name, ref, err)
	}
	if err := svc.StartContainer(resp.ID, container.StartOptions{}); err != nil {
		return "", nil, fmt.Errorf("docker: start container failed name=%q id=%q: %w", name, resp.ID, err)
	}
	compensate := func() { _ = svc.RemoveContainer(resp.ID, container.RemoveOptions{Force: true}) }
	return resp.ID, compensate, nil
}

func dockerStop(svc *deploydocker.DockerService, name string) (string, func(), error) {
	id, err := findContainerIDByName(svc, name)
	if err != nil {
		return "", nil, fmt.Errorf("docker: stop failed - %w", err)
	}
	if err := svc.StopContainer(id, container.StopOptions{}); err != nil {
		return "", nil, fmt.Errorf("docker: stop container failed name=%q id=%q: %w", name, id, err)
	}
	compensate := func() { _ = svc.StartContainer(id, container.StartOptions{}) }
	return id, compensate, nil
}

func dockerStart(svc *deploydocker.DockerService, name string) (string, func(), error) {
	id, err := findContainerIDByName(svc, name)
	if err != nil {
		return "", nil, fmt.Errorf("docker: start failed - %w", err)
	}
	if err := svc.StartContainer(id, container.StartOptions{}); err != nil {
		return "", nil, fmt.Errorf("docker: start container failed name=%q id=%q: %w", name, id, err)
	}
	compensate := func() { _ = svc.StopContainer(id, container.StopOptions{}) }
	return id, compensate, nil
}

func dockerRm(svc *deploydocker.DockerService, name string) (string, func(), error) {
	id, err := findContainerIDByName(svc, name)
	if err != nil {
		return "", nil, fmt.Errorf("docker: rm failed - %w", err)
	}
	if err := svc.RemoveContainer(id, container.RemoveOptions{Force: true}); err != nil {
		return "", nil, fmt.Errorf("docker: remove container failed name=%q id=%q: %w", name, id, err)
	}
	return id, nil, nil
}
