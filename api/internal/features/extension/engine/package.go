package engine

import (
	"fmt"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type packageModule struct{}

func (packageModule) Type() string { return "package" }

func (packageModule) Execute(sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	name, _ := step.Properties["name"].(string)
	state, _ := step.Properties["state"].(string)
	if name == "" {
		return "", nil, fmt.Errorf("package name is required")
	}
	if state == "" {
		state = "present"
	}

	pm, err := detectPackageManager(sshClient)
	if err != nil {
		return "", nil, err
	}

	cmd, err := buildPackageCommand(pm, state, name)
	if err != nil {
		return "", nil, err
	}

	if step.Timeout > 0 && hasCommand(sshClient, "timeout") {
		cmd = fmt.Sprintf("timeout %ds %s", step.Timeout, cmd)
	}
	out, err := sshClient.RunCommand(cmd)
	if err != nil {
		return "", nil, fmt.Errorf("package operation failed: %w (output: %s)", err, out)
	}
	var compensate func()
	switch state {
	case "present", "latest":
		compensate = func() {
			if rollbackCmd, rbErr := buildPackageCommand(pm, "absent", name); rbErr == nil {
				_, _ = sshClient.RunCommand(rollbackCmd)
			}
		}
	case "absent":
		compensate = func() {
			if rollbackCmd, rbErr := buildPackageCommand(pm, "present", name); rbErr == nil {
				_, _ = sshClient.RunCommand(rollbackCmd)
			}
		}
	}
	return strings.TrimSpace(out), compensate, nil
}

type packageBuilder func(state, name string) (string, error)

func buildPackageCommand(pm string, state string, name string) (string, error) {
	builders := map[string]packageBuilder{
		"apt":    buildAptCommand,
		"dnf":    buildDnfCommand,
		"yum":    buildYumCommand,
		"apk":    buildApkCommand,
		"pacman": buildPacmanCommand,
	}
	b, ok := builders[pm]
	if !ok {
		return "", fmt.Errorf("unsupported package manager: %s", pm)
	}
	return b(state, name)
}

func buildAptCommand(state, name string) (string, error) {
	switch state {
	case "present":
		return fmt.Sprintf("sudo apt-get update -y && sudo DEBIAN_FRONTEND=noninteractive apt-get install -y %s", name), nil
	case "absent":
		return fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt-get remove -y %s", name), nil
	case "latest":
		return fmt.Sprintf("sudo apt-get update -y && sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --only-upgrade %s", name), nil
	default:
		return "", fmt.Errorf("unsupported state '%s' for apt", state)
	}
}

func buildDnfCommand(state, name string) (string, error) {
	switch state {
	case "present":
		return fmt.Sprintf("sudo dnf install -y %s", name), nil
	case "absent":
		return fmt.Sprintf("sudo dnf remove -y %s", name), nil
	case "latest":
		return fmt.Sprintf("sudo dnf upgrade -y %s", name), nil
	default:
		return "", fmt.Errorf("unsupported state '%s' for dnf", state)
	}
}

func buildYumCommand(state, name string) (string, error) {
	switch state {
	case "present":
		return fmt.Sprintf("sudo yum install -y %s", name), nil
	case "absent":
		return fmt.Sprintf("sudo yum remove -y %s", name), nil
	case "latest":
		return fmt.Sprintf("sudo yum update -y %s", name), nil
	default:
		return "", fmt.Errorf("unsupported state '%s' for yum", state)
	}
}

func buildApkCommand(state, name string) (string, error) {
	switch state {
	case "present":
		return fmt.Sprintf("sudo apk add --no-cache %s", name), nil
	case "absent":
		return fmt.Sprintf("sudo apk del %s", name), nil
	case "latest":
		return fmt.Sprintf("sudo apk add --upgrade %s", name), nil
	default:
		return "", fmt.Errorf("unsupported state '%s' for apk", state)
	}
}

func buildPacmanCommand(state, name string) (string, error) {
	switch state {
	case "present":
		return fmt.Sprintf("sudo pacman -S --noconfirm %s", name), nil
	case "absent":
		return fmt.Sprintf("sudo pacman -R --noconfirm %s", name), nil
	case "latest":
		return "sudo pacman -Syu --noconfirm", nil
	default:
		return "", fmt.Errorf("unsupported state '%s' for pacman", state)
	}
}

func detectPackageManager(sshClient *ssh.SSH) (string, error) {
	switch {
	case hasCommand(sshClient, "apt-get"):
		return "apt", nil
	case hasCommand(sshClient, "dnf"):
		return "dnf", nil
	case hasCommand(sshClient, "yum"):
		return "yum", nil
	case hasCommand(sshClient, "apk"):
		return "apk", nil
	case hasCommand(sshClient, "pacman"):
		return "pacman", nil
	default:
		return "", fmt.Errorf("no supported package manager found")
	}
}

func hasCommand(sshClient *ssh.SSH, name string) bool {
	out, _ := sshClient.RunCommand("command -v " + name + " >/dev/null 2>&1 && echo yes || echo no")
	return strings.Contains(out, "yes")
}

func init() {
	RegisterModule(packageModule{})
}
