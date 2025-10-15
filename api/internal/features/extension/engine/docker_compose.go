package engine

import (
	"fmt"
	"strings"

	deploydocker "github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type dockerComposeModule struct{}

func (dockerComposeModule) Type() string { return "docker_compose" }

func (dockerComposeModule) Execute(_ *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	file, _ := step.Properties["file"].(string)
	action, _ := step.Properties["action"].(string) // up, down, pull, build, restart
	_, _ = step.Properties["project"].(string)
	_, _ = step.Properties["args"].(string)
	revertCmdRaw, _ := step.Properties["revert_cmd"].(string)
	_, _ = step.Properties["user"].(string)

	if action == "" {
		return "", nil, fmt.Errorf("docker_compose action is required")
	}

	svc := deploydocker.NewDockerService()

	type handler func() (string, func(), error)
	handlers := map[string]handler{
		"up":    func() (string, func(), error) { return composeUp(svc, file) },
		"down":  func() (string, func(), error) { return composeDown(svc, file) },
		"build": func() (string, func(), error) { return composeBuild(svc, file) },
	}

	h, ok := handlers[action]
	if !ok {
		return "", nil, fmt.Errorf("unsupported docker_compose action: %s", action)
	}
	out, comp, err := h()
	if err != nil {
		return "", nil, err
	}

	if revertCmdRaw != "" {
		// ignored by design in service backed module
	}
	return strings.TrimSpace(out), comp, nil
}

func composeUp(svc *deploydocker.DockerService, file string) (string, func(), error) {
	if err := svc.ComposeUp(file, map[string]string{}); err != nil {
		return "", nil, err
	}
	compensate := func() { _ = svc.ComposeDown(file) }
	return "compose up", compensate, nil
}

func composeDown(svc *deploydocker.DockerService, file string) (string, func(), error) {
	if err := svc.ComposeDown(file); err != nil {
		return "", nil, err
	}
	compensate := func() { _ = svc.ComposeUp(file, map[string]string{}) }
	return "compose down", compensate, nil
}

func composeBuild(svc *deploydocker.DockerService, file string) (string, func(), error) {
	if err := svc.ComposeBuild(file, map[string]string{}); err != nil {
		return "", nil, err
	}
	return "compose build", nil, nil
}

func init() {
	RegisterModule(dockerComposeModule{})
}
