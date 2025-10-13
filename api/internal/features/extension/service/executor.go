package service

import (
	"fmt"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/extension/engine"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type StepOutcome struct {
	Step   *types.ExecutionStep
	Output string
	Err    error
	Ignore bool
}

func (s *ExtensionService) executeRun(ctx *RunContext) {
	ctx.Exec.StartedAt = time.Now()
	_ = s.storage.UpdateExecution(ctx.Exec)

	s.logger.Log(logger.Info, fmt.Sprintf("Starting extension execution: %s", ctx.Exec.ID.String()), "")
	s.appendLog(ctx.Exec.ID, nil, "info", "execution_started", map[string]interface{}{})

	steps := ctx.Steps

	if stop := s.processPhase(ctx, &steps, "run", ctx.Spec.Execution.Run, 0); stop {
		s.logger.Log(logger.Error, fmt.Sprintf("Extension execution failed during run phase: %s", ctx.Exec.ID.String()), "")
		return
	}
	if stop := s.processPhase(ctx, &steps, "validate", ctx.Spec.Execution.Validate, len(ctx.Spec.Execution.Run)); stop {
		s.logger.Log(logger.Error, fmt.Sprintf("Extension execution failed during validate phase: %s", ctx.Exec.ID.String()), "")
		return
	}

	ctx.Exec.Status = types.ExecutionStatusCompleted
	finished := time.Now()
	ctx.Exec.CompletedAt = &finished
	_ = s.storage.UpdateExecution(ctx.Exec)

	s.logger.Log(logger.Info, fmt.Sprintf("Extension execution completed successfully: %s", ctx.Exec.ID.String()), "")
	s.appendLog(ctx.Exec.ID, nil, "info", "execution_completed", map[string]interface{}{"status": ctx.Exec.Status})
}

func (s *ExtensionService) processPhase(
	ctx *RunContext,
	steps *[]types.ExecutionStep,
	phase string,
	specSteps []types.SpecStep,
	offset int,
) bool {
	s.logger.Log(logger.Info, fmt.Sprintf("Processing phase: %s with %d spec steps, offset: %d", phase, len(specSteps), offset), "")
	s.logger.Log(logger.Info, fmt.Sprintf("Available steps: %d", len(*steps)), "")

	for idx := range specSteps {
		if s.shouldCancel(ctx) {
			s.markCancelled(ctx)
			return true
		}

		stepOrder := offset + idx + 1
		step := s.getStepByPhaseAndOrder(steps, phase, stepOrder)
		if step == nil {
			s.logger.Log(logger.Error, fmt.Sprintf("Step not found for phase: %s, order: %d", phase, stepOrder), "")
			continue
		}
		s.beginStep(step)

		out, runErr := s.executeSpecStep(ctx, specSteps[idx])
		step.Output = out
		outcome := StepOutcome{Step: step, Output: out, Err: runErr, Ignore: specSteps[idx].IgnoreErrors}
		if s.finalizeStep(ctx, outcome) {
			return true
		}
	}
	return false
}

func (s *ExtensionService) shouldCancel(ctx *RunContext) bool {
	cur, err := s.storage.GetExecutionByID(ctx.Exec.ID.String())
	return err == nil && cur.Status == types.ExecutionStatusCancelled
}

func (s *ExtensionService) markCancelled(ctx *RunContext) {
	ctx.Exec.Status = types.ExecutionStatusCancelled
	finished := time.Now()
	ctx.Exec.CompletedAt = &finished
	_ = s.storage.UpdateExecution(ctx.Exec)
}

func (s *ExtensionService) beginStep(step *types.ExecutionStep) {
	step.Status = types.ExecutionStatusRunning
	step.StartedAt = time.Now()

	// Log to console
	s.logger.Log(logger.Info, fmt.Sprintf("STEP STARTED: %s", step.StepName), "")
	sid := step.ID
	s.appendLog(step.ExecutionID, &sid, "info", "step_started", map[string]interface{}{"step_name": step.StepName, "phase": step.Phase, "order": step.StepOrder})

	_ = s.storage.UpdateExecutionStep(step)
}

func (s *ExtensionService) finalizeStep(ctx *RunContext, outcome StepOutcome) bool {
	if outcome.Err == nil {
		outcome.Step.Status = types.ExecutionStatusCompleted
		completed := time.Now()
		outcome.Step.CompletedAt = &completed
		successLog := fmt.Sprintf("[%s] STEP COMPLETED: %s\nOutput: %s",
			completed.Format("2006-01-02 15:04:05"),
			outcome.Step.StepName,
			outcome.Output)
		s.logger.Log(logger.Info, fmt.Sprintf("STEP COMPLETED: %s - %s", outcome.Step.StepName, outcome.Output), "")
		sid := outcome.Step.ID
		s.appendLog(ctx.Exec.ID, &sid, "info", "step_completed", map[string]interface{}{"step_name": outcome.Step.StepName, "output": outcome.Output})
		ctx.Exec.ExecutionLog = ctx.Exec.ExecutionLog + "\n" + successLog
		_ = s.storage.UpdateExecutionStep(outcome.Step)
		_ = s.storage.UpdateExecution(ctx.Exec)
		return false
	}

	outcome.Step.Status = types.ExecutionStatusFailed
	completed := time.Now()
	outcome.Step.CompletedAt = &completed
	errorLog := fmt.Sprintf("[%s] STEP FAILED: %s\nError: %v\nOutput: %s",
		completed.Format("2006-01-02 15:04:05"),
		outcome.Step.StepName,
		outcome.Err,
		outcome.Output)
	s.logger.Log(logger.Error, fmt.Sprintf("STEP FAILED: %s - Error: %v, Output: %s", outcome.Step.StepName, outcome.Err, outcome.Output), "")
	sid := outcome.Step.ID
	s.appendLog(ctx.Exec.ID, &sid, "error", "step_failed", map[string]interface{}{"step_name": outcome.Step.StepName, "error": outcome.Err.Error(), "output": outcome.Output})
	ctx.Exec.ExecutionLog = ctx.Exec.ExecutionLog + "\n" + errorLog
	_ = s.storage.UpdateExecutionStep(outcome.Step)
	_ = s.storage.UpdateExecution(ctx.Exec)

	if outcome.Ignore {
		return false
	}
	// perform rollback for prior successful steps
	ctx.rollback()
	ctx.Exec.Status = types.ExecutionStatusFailed
	ctx.Exec.ErrorMessage = outcome.Err.Error()
	finished := time.Now()
	ctx.Exec.CompletedAt = &finished
	_ = s.storage.UpdateExecution(ctx.Exec)
	return true
}

func (s *ExtensionService) getStepByPhaseAndOrder(steps *[]types.ExecutionStep, phase string, order int) *types.ExecutionStep {
	s.logger.Log(logger.Info, fmt.Sprintf("Looking for step: phase=%s, order=%d", phase, order), "")
	for i := range *steps {
		st := &(*steps)[i]
		s.logger.Log(logger.Info, fmt.Sprintf("Checking step: phase=%s, order=%d, name=%s", st.Phase, st.StepOrder, st.StepName), "")
		if st.Phase == phase && st.StepOrder == order {
			s.logger.Log(logger.Info, fmt.Sprintf("Found step: %s", st.StepName), "")
			return st
		}
	}
	s.logger.Log(logger.Error, fmt.Sprintf("Step not found: phase=%s, order=%d", phase, order), "")
	return nil
}

func (s *ExtensionService) executeSpecStep(ctx *RunContext, spec types.SpecStep) (string, error) {
	if m := engine.GetModule(spec.Type); m != nil {
		out, comp, err := m.Execute(ctx.SSH, spec, ctx.Vars)
		if err == nil && comp != nil {
			ctx.pushCompensation(comp)
		}
		return out, err
	}
	return "", fmt.Errorf("unsupported step type: %s", spec.Type)
}
