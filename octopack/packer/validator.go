package packer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	Path    string // JSON path to the problematic field
	Message string // Error message
	Warning bool   // If true, this is a warning rather than an error
}

func (e ValidationError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s", e.Path, e.Message)
	}
	return e.Message
}

// ValidationResult contains all validation errors and warnings.
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
}

// HasErrors returns true if there are any validation errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// AddError adds an error to the validation result.
func (r *ValidationResult) AddError(path, message string) {
	r.Errors = append(r.Errors, ValidationError{Path: path, Message: message})
}

// AddWarning adds a warning to the validation result.
func (r *ValidationResult) AddWarning(path, message string) {
	r.Warnings = append(r.Warnings, ValidationError{Path: path, Message: message, Warning: true})
}

// Validate performs comprehensive validation on a DockerfileSpec.
func Validate(spec *DockerfileSpec) *ValidationResult {
	result := &ValidationResult{}

	// Global validations
	validateGlobalRules(spec, result)

	// Collect stage names for cross-reference validation
	stageNames := make(map[string]bool)
	for i, stage := range spec.Stages {
		stagePath := fmt.Sprintf("stages[%d]", i)

		// Validate FROM instruction
		validateFromInstruction(&stage.From, stagePath+".from", result, i)

		// Track stage name
		if stage.From.As != "" {
			if stageNames[stage.From.As] {
				result.AddError(stagePath+".from.as", fmt.Sprintf("duplicate stage name: %s", stage.From.As))
			}
			stageNames[stage.From.As] = true
		}

		// Validate instructions in this stage
		validateStageInstructions(&stage, stagePath, result, stageNames)
	}

	return result
}

// validateGlobalRules validates global-level rules.
func validateGlobalRules(spec *DockerfileSpec, result *ValidationResult) {
	// At least one stage must exist
	if len(spec.Stages) == 0 {
		result.AddError("stages", "at least one stage must be defined")
		return
	}

	// Validate parser directives
	if spec.Directives != nil {
		validateParserDirectives(spec.Directives, "directives", result)
	}
}

// validateParserDirectives validates parser directive rules.
func validateParserDirectives(directives *ParserDirectives, path string, result *ValidationResult) {
	// Escape directive must be single character
	if directives.Escape != "" {
		if directives.Escape != "\\" && directives.Escape != "`" {
			result.AddError(path+".escape", "escape directive must be either \"\\\\\" (backslash) or \"`\" (backtick)")
		} else if len([]rune(directives.Escape)) != 1 {
			result.AddError(path+".escape", "escape directive must be a single character")
		}
	}
}

// validateFromInstruction validates a FROM instruction.
func validateFromInstruction(from *FromInstruction, path string, result *ValidationResult, stageIndex int) {
	// Image cannot be empty
	if from.Image == "" {
		result.AddError(path+".image", "FROM image cannot be empty")
	}

	// Cannot have both tag and digest
	if from.Tag != "" && from.Digest != "" {
		result.AddError(path, "FROM cannot have both tag and digest (they are mutually exclusive)")
	}

	// Validate stage name (AS)
	if from.As != "" {
		if !isValidStageName(from.As) {
			result.AddError(path+".as", "stage name must contain only alphanumeric characters, underscores, and hyphens")
		}
	}
}

// isValidStageName checks if a stage name follows Docker naming rules.
func isValidStageName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

// validateStageInstructions validates all instructions in a stage.
func validateStageInstructions(stage *Stage, stagePath string, result *ValidationResult, stageNames map[string]bool) {
	cmdCount := 0
	entrypointCount := 0

	for i, inst := range stage.Instructions {
		instPath := fmt.Sprintf("%s.instructions[%d]", stagePath, i)

		switch typedInst := inst.(type) {
		case RunInstruction:
			validateRunInstruction(&typedInst, instPath, result)
		case CmdInstruction:
			cmdCount++
			validateCmdInstruction(&typedInst, instPath, result)
		case EntrypointInstruction:
			entrypointCount++
			validateEntrypointInstruction(&typedInst, instPath, result)
		case CopyInstruction:
			validateCopyInstruction(&typedInst, instPath, result, stageNames)
		case AddInstruction:
			validateAddInstruction(&typedInst, instPath, result, stageNames)
		case EnvInstruction:
			validateEnvInstruction(&typedInst, instPath, result)
		case ArgInstruction:
			validateArgInstruction(&typedInst, instPath, result)
		case ExposeInstruction:
			validateExposeInstruction(&typedInst, instPath, result)
		case VolumeInstruction:
			validateVolumeInstruction(&typedInst, instPath, result)
		case UserInstruction:
			validateUserInstruction(&typedInst, instPath, result)
		case WorkdirInstruction:
			validateWorkdirInstruction(&typedInst, instPath, result)
		case LabelInstruction:
			validateLabelInstruction(&typedInst, instPath, result)
		case StopsignalInstruction:
			validateStopsignalInstruction(&typedInst, instPath, result)
		case HealthcheckInstruction:
			validateHealthcheckInstruction(&typedInst, instPath, result)
		case ShellInstruction:
			validateShellInstruction(&typedInst, instPath, result)
		case OnbuildInstruction:
			validateOnbuildInstruction(&typedInst, instPath, result)
		case MaintainerInstruction:
			// MAINTAINER is deprecated but still valid
			validateMaintainerInstruction(&typedInst, instPath, result)
		}
	}

	// Warn about multiple CMD/ENTRYPOINT (last one wins)
	if cmdCount > 1 {
		result.AddWarning(stagePath, fmt.Sprintf("multiple CMD instructions found (%d), only the last one will be used", cmdCount))
	}
	if entrypointCount > 1 {
		result.AddWarning(stagePath, fmt.Sprintf("multiple ENTRYPOINT instructions found (%d), only the last one will be used", entrypointCount))
	}
}

// validateRunInstruction validates a RUN instruction.
func validateRunInstruction(run *RunInstruction, path string, result *ValidationResult) {
	// Validate command form
	if run.Command != nil {
		if cmdArr, ok := run.Command.([]interface{}); ok {
			// Exec form
			if len(cmdArr) == 0 {
				result.AddError(path+".command", "RUN exec form command array cannot be empty")
			}
		}
		// Shell form (string) is always valid
	}

	// Validate mount options
	for i, mount := range run.Mount {
		mountPath := fmt.Sprintf("%s.mount[%d]", path, i)
		validateMountOption(&mount, mountPath, result)
	}

	// Validate network mode
	if run.Network != "" {
		validNetworks := map[string]bool{"default": true, "none": true, "host": true}
		if !validNetworks[run.Network] {
			result.AddError(path+".network", "network must be one of: default, none, host")
		}
		if run.Network == "host" {
			result.AddWarning(path+".network", "--network=host requires entitlements")
		}
	}

	// Validate security mode
	if run.Security != "" {
		validSecurity := map[string]bool{"sandbox": true, "insecure": true}
		if !validSecurity[run.Security] {
			result.AddError(path+".security", "security must be one of: sandbox, insecure")
		}
		if run.Security == "insecure" {
			result.AddWarning(path+".security", "--security=insecure requires entitlements")
		}
	}
}

// validateMountOption validates a mount option.
func validateMountOption(mount *MountOption, path string, result *ValidationResult) {
	validTypes := map[string]bool{"bind": true, "cache": true, "tmpfs": true, "secret": true, "ssh": true}
	if !validTypes[mount.Type] {
		result.AddError(path+".type", "mount type must be one of: bind, cache, tmpfs, secret, ssh")
		return
	}

	// Target is required for all types
	if mount.Target == "" {
		result.AddError(path+".target", "mount target is required")
	}

	// Type-specific validations
	switch mount.Type {
	case "bind":
		if mount.Source == "" && mount.From == "" {
			result.AddError(path, "bind mount requires either source or from")
		}
	case "cache":
		// Cache mounts don't require source/from
	case "tmpfs":
		// Tmpfs mounts don't require source/from
	case "secret", "ssh":
		// Secret/SSH mounts require id
		if mount.ID == "" {
			result.AddError(path+".id", fmt.Sprintf("%s mount requires id", mount.Type))
		}
	}
}

// validateCmdInstruction validates a CMD instruction.
func validateCmdInstruction(cmd *CmdInstruction, path string, result *ValidationResult) {
	if cmd.Command != nil {
		if cmdArr, ok := cmd.Command.([]interface{}); ok {
			// Exec form
			if len(cmdArr) == 0 {
				result.AddError(path+".command", "CMD exec form command array cannot be empty")
			}
		}
		// Shell form (string) is always valid
	}
}

// validateEntrypointInstruction validates an ENTRYPOINT instruction.
func validateEntrypointInstruction(entrypoint *EntrypointInstruction, path string, result *ValidationResult) {
	if entrypoint.Command != nil {
		if cmdArr, ok := entrypoint.Command.([]interface{}); ok {
			// Exec form
			if len(cmdArr) == 0 {
				result.AddError(path+".command", "ENTRYPOINT exec form command array cannot be empty")
			}
		}
		// Shell form (string) is always valid
	}
}

// validateCopyInstruction validates a COPY instruction.
func validateCopyInstruction(copy *CopyInstruction, path string, result *ValidationResult, stageNames map[string]bool) {
	// Sources must not be empty
	if len(copy.Sources) == 0 {
		result.AddError(path+".sources", "COPY sources cannot be empty")
	}

	// Destination must not be empty
	if copy.Destination == "" {
		result.AddError(path+".destination", "COPY destination cannot be empty")
	}

	// If multiple sources, destination should end with /
	if len(copy.Sources) > 1 && !strings.HasSuffix(copy.Destination, "/") {
		result.AddWarning(path+".destination", "when copying multiple sources, destination should end with '/'")
	}

	// Validate chmod
	if copy.Chmod != "" {
		if !isValidOctal(copy.Chmod) {
			result.AddError(path+".chmod", "chmod must be a valid octal string (3 or 4 digits)")
		}
	}

	// Validate chown
	if copy.Chown != "" {
		if !isValidChown(copy.Chown) {
			result.AddError(path+".chown", "chown must match pattern: user, user:group, uid, or uid:gid")
		}
	}

	// Validate --from reference
	if copy.From != "" {
		if !stageNames[copy.From] {
			// Could be an image name, so this is just a warning
			result.AddWarning(path+".from", fmt.Sprintf("--from references stage '%s' which may not exist", copy.From))
		}
	}
}

// validateAddInstruction validates an ADD instruction (extends COPY validation).
func validateAddInstruction(add *AddInstruction, path string, result *ValidationResult, stageNames map[string]bool) {
	// Use COPY validation
	copyInst := add.CopyInstruction
	validateCopyInstruction(&copyInst, path, result, stageNames)

	// Validate checksum format if provided
	if add.Checksum != "" {
		if !strings.HasPrefix(add.Checksum, "sha256:") {
			result.AddError(path+".checksum", "checksum must start with 'sha256:'")
		}
	}
}

// validateEnvInstruction validates an ENV instruction.
func validateEnvInstruction(env *EnvInstruction, path string, result *ValidationResult) {
	// At least one variable required
	if len(env.Variables) == 0 {
		result.AddError(path+".variables", "ENV must have at least one variable")
		return
	}

	// Validate variable names
	for key := range env.Variables {
		if !isValidEnvVarName(key) {
			result.AddError(path+".variables", fmt.Sprintf("invalid environment variable name: %s (must be alphanumeric with underscores)", key))
		}
	}
}

// isValidEnvVarName checks if an environment variable name is valid.
func isValidEnvVarName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	return matched
}

// validateArgInstruction validates an ARG instruction.
func validateArgInstruction(arg *ArgInstruction, path string, result *ValidationResult) {
	// Name must be valid identifier
	if arg.Name == "" {
		result.AddError(path+".name", "ARG name cannot be empty")
	} else if !isValidEnvVarName(arg.Name) {
		result.AddError(path+".name", fmt.Sprintf("invalid ARG name: %s (must be a valid identifier)", arg.Name))
	}
}

// validateExposeInstruction validates an EXPOSE instruction.
func validateExposeInstruction(expose *ExposeInstruction, path string, result *ValidationResult) {
	// At least one port required
	if len(expose.Ports) == 0 {
		result.AddError(path+".ports", "EXPOSE must have at least one port")
		return
	}

	for i, port := range expose.Ports {
		portPath := fmt.Sprintf("%s.ports[%d]", path, i)

		// Port must be 1-65535
		if port.Port < 1 || port.Port > 65535 {
			result.AddError(portPath+".port", fmt.Sprintf("port must be between 1 and 65535, got %d", port.Port))
		}

		// Protocol must be tcp or udp
		if port.Protocol != "" {
			if port.Protocol != "tcp" && port.Protocol != "udp" {
				result.AddError(portPath+".protocol", "protocol must be 'tcp' or 'udp'")
			}
		}
	}
}

// validateVolumeInstruction validates a VOLUME instruction.
func validateVolumeInstruction(volume *VolumeInstruction, path string, result *ValidationResult) {
	// At least one path required
	if len(volume.Paths) == 0 {
		result.AddError(path+".paths", "VOLUME must have at least one path")
		return
	}

	// Warn if paths are not absolute
	for i, volPath := range volume.Paths {
		if !strings.HasPrefix(volPath, "/") {
			result.AddWarning(fmt.Sprintf("%s.paths[%d]", path, i), "volume path should be absolute (starting with '/')")
		}
	}
}

// validateUserInstruction validates a USER instruction.
func validateUserInstruction(user *UserInstruction, path string, result *ValidationResult) {
	// User must not be empty
	if user.User == "" {
		result.AddError(path+".user", "USER cannot be empty")
	}
}

// validateWorkdirInstruction validates a WORKDIR instruction.
func validateWorkdirInstruction(workdir *WorkdirInstruction, path string, result *ValidationResult) {
	// Path must not be empty
	if workdir.Path == "" {
		result.AddError(path+".path", "WORKDIR path cannot be empty")
	}
}

// validateLabelInstruction validates a LABEL instruction.
func validateLabelInstruction(label *LabelInstruction, path string, result *ValidationResult) {
	// At least one label required
	if len(label.Labels) == 0 {
		result.AddError(path+".labels", "LABEL must have at least one label")
	}
}

// validateStopsignalInstruction validates a STOPSIGNAL instruction.
func validateStopsignalInstruction(stopsignal *StopsignalInstruction, path string, result *ValidationResult) {
	// Signal must not be empty
	if stopsignal.Signal == "" {
		result.AddError(path+".signal", "STOPSIGNAL cannot be empty")
	}
}

// validateHealthcheckInstruction validates a HEALTHCHECK instruction.
func validateHealthcheckInstruction(healthcheck *HealthcheckInstruction, path string, result *ValidationResult) {
	// Cannot have both none and command
	if healthcheck.None && healthcheck.Command != nil {
		result.AddError(path, "HEALTHCHECK cannot have both 'none: true' and 'command'")
		return
	}

	// If not none, must have command
	if !healthcheck.None {
		if healthcheck.Command == nil {
			result.AddError(path+".command", "HEALTHCHECK must have a command if 'none' is not true")
		}
	}

	// Validate duration strings
	if healthcheck.Interval != "" {
		if !isValidDuration(healthcheck.Interval) {
			result.AddError(path+".interval", fmt.Sprintf("invalid duration format: %s (expected format like '30s', '5m')", healthcheck.Interval))
		}
	}
	if healthcheck.Timeout != "" {
		if !isValidDuration(healthcheck.Timeout) {
			result.AddError(path+".timeout", fmt.Sprintf("invalid duration format: %s", healthcheck.Timeout))
		}
	}
	if healthcheck.StartPeriod != "" {
		if !isValidDuration(healthcheck.StartPeriod) {
			result.AddError(path+".start_period", fmt.Sprintf("invalid duration format: %s", healthcheck.StartPeriod))
		}
	}
	if healthcheck.StartInterval != "" {
		if !isValidDuration(healthcheck.StartInterval) {
			result.AddError(path+".start_interval", fmt.Sprintf("invalid duration format: %s", healthcheck.StartInterval))
		}
	}

	// Retries must be positive
	if healthcheck.Retries != nil && *healthcheck.Retries < 0 {
		result.AddError(path+".retries", "retries must be non-negative")
	}
}

// isValidDuration checks if a string is a valid duration.
func isValidDuration(s string) bool {
	_, err := time.ParseDuration(s)
	return err == nil
}

// validateShellInstruction validates a SHELL instruction.
func validateShellInstruction(shell *ShellInstruction, path string, result *ValidationResult) {
	// Must be exec form (array)
	if len(shell.Shell) == 0 {
		result.AddError(path+".shell", "SHELL array cannot be empty")
	}
}

// validateOnbuildInstruction validates an ONBUILD instruction.
func validateOnbuildInstruction(onbuild *OnbuildInstruction, path string, result *ValidationResult) {
	if onbuild.Instruction == nil {
		result.AddError(path+".instruction", "ONBUILD must have a nested instruction")
		return
	}

	instType := onbuild.Instruction.Type()

	// Cannot nest ONBUILD
	if instType == "onbuild" {
		result.AddError(path+".instruction", "ONBUILD cannot nest another ONBUILD")
	}

	// Cannot wrap FROM or MAINTAINER
	if instType == "from" {
		result.AddError(path+".instruction", "ONBUILD cannot wrap FROM")
	}
	if instType == "maintainer" {
		result.AddError(path+".instruction", "ONBUILD cannot wrap MAINTAINER")
	}
}

// validateMaintainerInstruction validates a MAINTAINER instruction.
func validateMaintainerInstruction(maintainer *MaintainerInstruction, path string, result *ValidationResult) {
	// Name must not be empty
	if maintainer.Name == "" {
		result.AddError(path+".name", "MAINTAINER name cannot be empty")
	}
	// Warn that MAINTAINER is deprecated
	result.AddWarning(path, "MAINTAINER is deprecated, use LABEL instead")
}

// isValidOctal checks if a string is a valid octal number (3 or 4 digits).
func isValidOctal(s string) bool {
	if len(s) != 3 && len(s) != 4 {
		return false
	}
	_, err := strconv.ParseUint(s, 8, 32)
	return err == nil
}

// isValidChown checks if a string matches chown patterns: user, user:group, uid, uid:gid.
func isValidChown(s string) bool {
	parts := strings.Split(s, ":")
	if len(parts) > 2 {
		return false
	}
	// Both parts should be non-empty
	for _, part := range parts {
		if part == "" {
			return false
		}
	}
	return true
}
