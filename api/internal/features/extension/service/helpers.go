package service

import (
	"strconv"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

func (s *ExtensionService) hasCommand(sshClient *ssh.SSH, name string) bool {
	out, _ := sshClient.RunCommand("command -v " + name + " >/dev/null 2>&1 && echo yes || echo no")
	return strings.Contains(out, "yes")
}

func (s *ExtensionService) firstAvailable(sshClient *ssh.SSH, names ...string) string {
	for _, n := range names {
		if s.hasCommand(sshClient, n) {
			return n
		}
	}
	return ""
}

func (s *ExtensionService) timeoutPrefix(sshClient *ssh.SSH, seconds int) string {
	if seconds <= 0 {
		return ""
	}
	if s.hasCommand(sshClient, "timeout") {
		return "timeout " + strconv.Itoa(seconds) + "s "
	}
	return ""
}

type UserToolset struct {
	add      string
	del      string
	mod      string
	groupDel string
}

func (s *ExtensionService) userTools(sshClient *ssh.SSH) UserToolset {
	return UserToolset{
		add:      s.firstAvailable(sshClient, "useradd", "adduser"),
		del:      s.firstAvailable(sshClient, "userdel", "deluser"),
		mod:      s.firstAvailable(sshClient, "usermod", "chsh"),
		groupDel: s.firstAvailable(sshClient, "gpasswd", "deluser"),
	}
}

func (s *ExtensionService) serviceCmd(sshClient *ssh.SSH, action string, name string, timeout int) string {
	var base string
	if s.hasCommand(sshClient, "systemctl") {
		base = "sudo systemctl " + action + " " + name
	} else {
		base = "sudo service " + name + " " + action
	}
	return s.timeoutPrefix(sshClient, timeout) + base
}
