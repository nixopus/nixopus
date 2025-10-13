package engine

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type serviceModule struct{}

func (serviceModule) Type() string { return "service" }

func (serviceModule) Execute(sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	name, _ := step.Properties["name"].(string)
	action, _ := step.Properties["action"].(string)
	revertAction, _ := step.Properties["revert_action"].(string)
	runAsUser, _ := step.Properties["user"].(string)
	if name == "" {
		return "", nil, fmt.Errorf("service name is required for service step")
	}
	if action == "" {
		return "", nil, fmt.Errorf("service action is required for service step")
	}
	cmd := serviceCmd(sshClient, action, name, step.Timeout)
	if runAsUser != "" {
		cmd = fmt.Sprintf("sudo -u %s %s", runAsUser, cmd)
	}
	output, err := sshClient.RunCommand(cmd)
	if err != nil {
		return "", nil, fmt.Errorf("failed to execute service command '%s': %w (output: %s)", cmd, err, output)
	}
	if revertAction == "" {
		switch action {
		case "start":
			revertAction = "stop"
		case "stop":
			revertAction = "start"
		case "enable":
			revertAction = "disable"
		case "disable":
			revertAction = "enable"
		case "restart":
			revertAction = "restart"
		}
	}
	var compensate func()
	if revertAction != "" {
		rev := serviceCmd(sshClient, revertAction, name, step.Timeout)
		if runAsUser != "" {
			rev = fmt.Sprintf("sudo -u %s %s", runAsUser, rev)
		}
		compensate = func() { _, _ = sshClient.RunCommand(rev) }
	}
	return output, compensate, nil
}

func serviceCmd(sshClient *ssh.SSH, action string, name string, timeout int) string {
	var base string
	if hasCommand(sshClient, "systemctl") {
		base = "sudo systemctl " + action + " " + name
	} else {
		base = "sudo service " + name + " " + action
	}
	if timeout > 0 && hasCommand(sshClient, "timeout") {
		return fmt.Sprintf("timeout %ds %s", timeout, base)
	}
	return base
}

func init() {
	RegisterModule(serviceModule{})
}
