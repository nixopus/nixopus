package engine

import (
	"fmt"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type userModule struct{}

func (userModule) Type() string { return "user" }

func (userModule) Execute(sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	username, _ := step.Properties["username"].(string)
	action, _ := step.Properties["action"].(string)
	shell, _ := step.Properties["shell"].(string)
	home, _ := step.Properties["home"].(string)
	groups, _ := step.Properties["groups"].(string)
	revertAction, _ := step.Properties["revert_action"].(string)

	if username == "" {
		return "", nil, fmt.Errorf("username is required for user operations")
	}

	tools := userTools(sshClient)
	var cmd string
	switch action {
	case "ensure":
		var err error
		cmd, err = buildEnsureUserCmd(tools, username, shell, home, groups)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build ensure user command: %w", err)
		}
	case "delete":
		var err error
		cmd, err = buildDeleteUserCmd(tools, username)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build delete user command: %w", err)
		}
	case "modify":
		var err error
		cmd, err = buildModifyUserCmd(username, shell, home, groups)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build modify user command: %w", err)
		}
	case "add_groups":
		var err error
		cmd, err = buildAddGroupsCmd(username, groups)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build add groups command: %w", err)
		}
	case "remove_groups":
		var err error
		cmd, err = buildRemoveGroupsCmd(tools, username, groups)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build remove groups command: %w", err)
		}
	case "check":
		cmd = buildCheckUserCmd(username)
	default:
		return "", nil, fmt.Errorf("unsupported user action: %s", action)
	}

	if step.Timeout > 0 && hasCommand(sshClient, "timeout") {
		cmd = fmt.Sprintf("timeout %ds %s", step.Timeout, cmd)
	}

	output, err := sshClient.RunCommand(cmd)
	if err != nil {
		return "", nil, fmt.Errorf("failed to execute user command '%s': %w (output: %s)", cmd, err, output)
	}
	if revertAction == "" {
		switch action {
		case "ensure":
			revertAction = "delete"
		case "delete":
			revertAction = "ensure"
		case "add_groups":
			revertAction = "remove_groups"
		case "remove_groups":
			revertAction = "add_groups"
		case "modify":
			revertAction = "" // cannot reliably revert without pre-state
		case "check":
			revertAction = "" // no-op
		}
	}
	var compensate func()
	if revertAction != "" {
		tools := userTools(sshClient)
		var rev string
		var rbErr error
		switch revertAction {
		case "ensure":
			rev, rbErr = buildEnsureUserCmd(tools, username, shell, home, groups)
		case "delete":
			rev, rbErr = buildDeleteUserCmd(tools, username)
		case "add_groups":
			rev, rbErr = buildAddGroupsCmd(username, groups)
		case "remove_groups":
			rev, rbErr = buildRemoveGroupsCmd(tools, username, groups)
		}
		if rbErr == nil && rev != "" {
			if step.Timeout > 0 && hasCommand(sshClient, "timeout") {
				rev = fmt.Sprintf("timeout %ds %s", step.Timeout, rev)
			}
			compensate = func() { _, _ = sshClient.RunCommand(rev) }
		}
	}
	return output, compensate, nil
}

type userToolset struct {
	add      string
	del      string
	mod      string
	groupDel string
}

func userTools(sshClient *ssh.SSH) userToolset {
	return userToolset{
		add:      firstAvailable(sshClient, "/usr/sbin/useradd", "/usr/bin/useradd", "useradd", "/usr/sbin/adduser", "/usr/bin/adduser", "adduser"),
		del:      firstAvailable(sshClient, "/usr/sbin/userdel", "/usr/bin/userdel", "userdel", "/usr/sbin/deluser", "/usr/bin/deluser", "deluser"),
		mod:      firstAvailable(sshClient, "/usr/sbin/usermod", "/usr/bin/usermod", "usermod", "/usr/bin/chsh", "chsh"),
		groupDel: firstAvailable(sshClient, "/usr/bin/gpasswd", "gpasswd", "/usr/sbin/deluser", "/usr/bin/deluser", "deluser"),
	}
}

func firstAvailable(sshClient *ssh.SSH, names ...string) string {
	for _, n := range names {
		if hasCommand(sshClient, n) {
			return n
		}
	}
	return ""
}

func buildEnsureUserCmd(tools userToolset, username string, shell string, home string, groups string) (string, error) {
	var cmd string
	if strings.Contains(tools.add, "useradd") {
		cmd = fmt.Sprintf("id -u %s >/dev/null 2>&1 || sudo useradd %s", username, username)
	} else if strings.Contains(tools.add, "adduser") {
		cmd = fmt.Sprintf("id -u %s >/dev/null 2>&1 || sudo adduser -D %s", username, username)
	} else {
		return "", fmt.Errorf("user management tool not found")
	}
	if shell != "" {
		if strings.Contains(tools.mod, "usermod") {
			cmd += fmt.Sprintf(" && sudo usermod -s %s %s", shell, username)
		} else if strings.Contains(tools.mod, "chsh") {
			cmd += fmt.Sprintf(" && sudo chsh -s %s %s", shell, username)
		}
	}
	if home != "" {
		cmd += fmt.Sprintf(" && sudo usermod -d %s %s", home, username)
	}
	if groups != "" {
		cmd += fmt.Sprintf(" && sudo usermod -aG %s %s", groups, username)
	}
	return cmd, nil
}

func buildDeleteUserCmd(tools userToolset, username string) (string, error) {
	if strings.Contains(tools.del, "userdel") {
		return fmt.Sprintf("id -u %s >/dev/null 2>&1 && sudo userdel -r %s || true", username, username), nil
	}
	if strings.Contains(tools.del, "deluser") {
		return fmt.Sprintf("id -u %s >/dev/null 2>&1 && sudo deluser --remove-home %s || true", username, username), nil
	}
	return "", fmt.Errorf("user management tool not found")
}

func buildModifyUserCmd(username string, shell string, home string, groups string) (string, error) {
	parts := []string{}
	if shell != "" {
		parts = append(parts, fmt.Sprintf("-s %s", shell))
	}
	if home != "" {
		parts = append(parts, fmt.Sprintf("-d %s", home))
	}
	if groups != "" {
		parts = append(parts, fmt.Sprintf("-aG %s", groups))
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("no changes")
	}
	return fmt.Sprintf("sudo usermod %s %s", strings.Join(parts, " "), username), nil
}

func buildAddGroupsCmd(username string, groups string) (string, error) {
	if groups == "" {
		return "", fmt.Errorf("no groups provided")
	}
	cmds := []string{}
	for _, g := range strings.Split(groups, ",") {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		cmds = append(cmds, fmt.Sprintf("sudo usermod -aG %s %s", g, username))
	}
	if len(cmds) == 0 {
		return "", fmt.Errorf("no groups provided")
	}
	return strings.Join(cmds, " && "), nil
}

func buildRemoveGroupsCmd(tools userToolset, username string, groups string) (string, error) {
	if groups == "" {
		return "", fmt.Errorf("no groups provided")
	}
	cmds := []string{}
	for _, g := range strings.Split(groups, ",") {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		if strings.Contains(tools.groupDel, "gpasswd") {
			cmds = append(cmds, fmt.Sprintf("sudo gpasswd -d %s %s || true", username, g))
		} else if strings.Contains(tools.groupDel, "deluser") {
			cmds = append(cmds, fmt.Sprintf("sudo deluser %s %s || true", username, g))
		}
	}
	if len(cmds) == 0 {
		return "", fmt.Errorf("no groups provided")
	}
	return strings.Join(cmds, " && "), nil
}

func buildCheckUserCmd(username string) string {
	return fmt.Sprintf("id -u %s >/dev/null 2>&1 && echo exists || echo missing", username)
}

func init() {
	RegisterModule(userModule{})
}
