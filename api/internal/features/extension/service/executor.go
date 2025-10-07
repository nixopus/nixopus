package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *ExtensionService) executeRun(exec *types.ExtensionExecution, spec types.ExtensionSpec, vars map[string]interface{}) {
	exec.StartedAt = time.Now()
	_ = s.storage.UpdateExecution(exec)

	replacer := buildReplacer(vars)
	sshClient := ssh.NewSSH()
	steps, _ := s.storage.ListExecutionSteps(exec.ID.String())

	if stop := s.processPhase(exec, &steps, "run", spec.Execution.Run, 0, sshClient, replacer); stop {
		return
	}
	if stop := s.processPhase(exec, &steps, "validate", spec.Execution.Validate, len(spec.Execution.Run), sshClient, replacer); stop {
		return
	}

	exec.Status = types.ExecutionStatusCompleted
	finished := time.Now()
	exec.CompletedAt = &finished
	_ = s.storage.UpdateExecution(exec)
}

func buildReplacer(vars map[string]interface{}) func(string) string {
	return func(in string) string {
		out := in
		for k, v := range vars {
			token := fmt.Sprintf("{{ %s }}", k)
			out = strings.ReplaceAll(out, token, fmt.Sprint(v))
		}
		return out
	}
}

func (s *ExtensionService) processPhase(
	exec *types.ExtensionExecution,
	steps *[]types.ExecutionStep,
	phase string,
	specSteps []types.SpecStep,
	offset int,
	sshClient *ssh.SSH,
	replacer func(string) string,
) bool {
	for idx := range specSteps {
		if s.shouldCancel(exec) {
			s.markCancelled(exec)
			return true
		}

		step := s.getStepByPhaseAndOrder(steps, phase, offset+idx+1)
		if step == nil {
			continue
		}
		s.beginStep(step)

		out, runErr := s.executeSpecStep(specSteps[idx], sshClient, replacer)
		step.Output = out
		if runErr != nil {
			if s.failStep(exec, step, out, runErr, specSteps[idx].IgnoreErrors) {
				return true
			}
			continue
		}

		s.completeStep(exec, step, out)
	}
	return false
}

func (s *ExtensionService) shouldCancel(exec *types.ExtensionExecution) bool {
	cur, err := s.storage.GetExecutionByID(exec.ID.String())
	return err == nil && cur.Status == types.ExecutionStatusCancelled
}

func (s *ExtensionService) markCancelled(exec *types.ExtensionExecution) {
	exec.Status = types.ExecutionStatusCancelled
	finished := time.Now()
	exec.CompletedAt = &finished
	_ = s.storage.UpdateExecution(exec)
}

func (s *ExtensionService) beginStep(step *types.ExecutionStep) {
	step.Status = types.ExecutionStatusRunning
	step.StartedAt = time.Now()
	_ = s.storage.UpdateExecutionStep(step)
}

func (s *ExtensionService) completeStep(exec *types.ExtensionExecution, step *types.ExecutionStep, out string) {
	step.Status = types.ExecutionStatusCompleted
	completed := time.Now()
	step.CompletedAt = &completed
	_ = s.storage.UpdateExecutionStep(step)
	exec.ExecutionLog = exec.ExecutionLog + "\n" + out
	_ = s.storage.UpdateExecution(exec)
}

func (s *ExtensionService) failStep(exec *types.ExtensionExecution, step *types.ExecutionStep, out string, runErr error, ignore bool) bool {
	step.Status = types.ExecutionStatusFailed
	completed := time.Now()
	step.CompletedAt = &completed
	_ = s.storage.UpdateExecutionStep(step)
	exec.ExecutionLog = exec.ExecutionLog + "\n" + out
	_ = s.storage.UpdateExecution(exec)
	if ignore {
		return false
	}
	exec.Status = types.ExecutionStatusFailed
	exec.ErrorMessage = runErr.Error()
	finished := time.Now()
	exec.CompletedAt = &finished
	_ = s.storage.UpdateExecution(exec)
	return true
}

func (s *ExtensionService) getStepByPhaseAndOrder(steps *[]types.ExecutionStep, phase string, order int) *types.ExecutionStep {
	for i := range *steps {
		st := &(*steps)[i]
		if st.Phase == phase && st.StepOrder == order {
			return st
		}
	}
	return nil
}

func (s *ExtensionService) executeSpecStep(spec types.SpecStep, sshClient *ssh.SSH, replacer func(string) string) (string, error) {
	switch spec.Type {
	case "command":
		cmd, _ := spec.Properties["cmd"].(string)
		cmd = replacer(cmd)
		prefix := s.timeoutPrefix(sshClient, spec.Timeout)
		return sshClient.RunCommand(prefix + cmd)
	case "file":
		return executeFileStep(sshClient, spec.Properties, replacer)
	case "service":
		name, _ := spec.Properties["name"].(string)
		action, _ := spec.Properties["action"].(string)
		name = replacer(name)
		// Prefer systemctl; if not available, fallback to service
		cmd := s.serviceCmd(sshClient, action, name, spec.Timeout)
		return sshClient.RunCommand(cmd)
	case "user":
		return s.executeUserStep(sshClient, spec.Properties, replacer, spec.Timeout)
	default:
		return "unsupported step type", nil
	}
}
