package engine

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type commandModule struct{}

func (commandModule) Type() string { return "command" }

func (commandModule) Execute(sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	raw, _ := step.Properties["cmd"].(string)
	if raw == "" {
		return "", nil, fmt.Errorf("command: 'cmd' is required")
	}
	revertRaw, _ := step.Properties["revert_cmd"].(string)
	user, _ := step.Properties["user"].(string)

	cmd := replaceVars(raw, vars)
	if user != "" {
		cmd = fmt.Sprintf("sudo -u %s %s", user, cmd)
	}
	if step.Timeout > 0 && hasCommand(sshClient, "timeout") {
		cmd = fmt.Sprintf("timeout %ds %s", step.Timeout, cmd)
	}
	output, err := sshClient.RunCommand(cmd)
	if err != nil {
		return "", nil, fmt.Errorf("command: execution failed cmd=%q user=%q timeout=%ds: %w (output: %s)", cmd, user, step.Timeout, err, output)
	}

	var compensate func()
	if revertRaw != "" {
		rev := replaceVars(revertRaw, vars)
		if user != "" {
			rev = fmt.Sprintf("sudo -u %s %s", user, rev)
		}
		if step.Timeout > 0 && hasCommand(sshClient, "timeout") {
			rev = fmt.Sprintf("timeout %ds %s", step.Timeout, rev)
		}
		compensate = func() { _, _ = sshClient.RunCommand(rev) }
	}
	return output, compensate, nil
}

func init() {
	RegisterModule(commandModule{})
}
