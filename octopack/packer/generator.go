package packer

import (
	"fmt"
	"sort"
	"strings"
)

// Generate converts a DockerfileSpec to a Dockerfile string.
func Generate(spec *DockerfileSpec) (string, error) {
	var buf strings.Builder

	// Determine escape character (default is backslash)
	escapeChar := "\\"
	if spec.Directives != nil && spec.Directives.Escape != "" {
		escapeChar = spec.Directives.Escape
	}

	// 1. Parser directives
	if spec.Directives != nil {
		generateParserDirectives(spec.Directives, &buf)
		buf.WriteString("\n")
	}

	// 2. Global ARG instructions
	if len(spec.GlobalArgs) > 0 {
		for _, arg := range spec.GlobalArgs {
			generateArgInstruction(&arg, &buf, escapeChar)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// 3. Stages
	for i, stage := range spec.Stages {
		if i > 0 {
			buf.WriteString("\n")
		}
		generateStage(&stage, &buf, escapeChar)
	}

	return strings.TrimRight(buf.String(), "\n"), nil
}

// generateParserDirectives generates parser directives.
func generateParserDirectives(directives *ParserDirectives, buf *strings.Builder) {
	if directives.Syntax != "" {
		buf.WriteString(fmt.Sprintf("# syntax=%s\n", directives.Syntax))
	}
	if directives.Escape != "" {
		buf.WriteString(fmt.Sprintf("# escape=%s\n", directives.Escape))
	}
	if directives.Check != nil {
		generateCheckDirective(directives.Check, buf)
	}
}

// generateCheckDirective generates check directive.
func generateCheckDirective(check *CheckDirective, buf *strings.Builder) {
	var parts []string

	if len(check.Skip) > 0 {
		skipValue := strings.Join(check.Skip, ",")
		parts = append(parts, fmt.Sprintf("skip=%s", skipValue))
	}

	if check.Error {
		parts = append(parts, "error=true")
	}

	if len(parts) > 0 {
		buf.WriteString(fmt.Sprintf("# check=%s\n", strings.Join(parts, ";")))
	}
}

// generateStage generates a single build stage.
func generateStage(stage *Stage, buf *strings.Builder, escapeChar string) {
	// FROM instruction
	generateFromInstruction(&stage.From, buf)
	buf.WriteString("\n")

	// Other instructions
	for _, inst := range stage.Instructions {
		generateInstruction(inst, buf, escapeChar)
		buf.WriteString("\n")
	}
}

// generateFromInstruction generates a FROM instruction.
func generateFromInstruction(from *FromInstruction, buf *strings.Builder) {
	buf.WriteString("FROM")

	// Platform
	if from.Platform != "" {
		buf.WriteString(fmt.Sprintf(" --platform=%s", from.Platform))
	}

	// Image
	buf.WriteString(" " + from.Image)

	// Tag or digest
	if from.Tag != "" {
		buf.WriteString(":" + from.Tag)
	} else if from.Digest != "" {
		buf.WriteString("@" + from.Digest)
	}

	// Stage name
	if from.As != "" {
		buf.WriteString(" AS " + from.As)
	}
}

// generateInstruction generates any instruction type.
func generateInstruction(inst Instruction, buf *strings.Builder, escapeChar string) {
	switch typedInst := inst.(type) {
	case RunInstruction:
		generateRunInstruction(&typedInst, buf, escapeChar)
	case CmdInstruction:
		generateCmdInstruction(&typedInst, buf, escapeChar)
	case EntrypointInstruction:
		generateEntrypointInstruction(&typedInst, buf, escapeChar)
	case CopyInstruction:
		generateCopyInstruction(&typedInst, buf, escapeChar)
	case AddInstruction:
		generateAddInstruction(&typedInst, buf, escapeChar)
	case EnvInstruction:
		generateEnvInstruction(&typedInst, buf, escapeChar)
	case ArgInstruction:
		generateArgInstruction(&typedInst, buf, escapeChar)
	case ExposeInstruction:
		generateExposeInstruction(&typedInst, buf)
	case VolumeInstruction:
		generateVolumeInstruction(&typedInst, buf, escapeChar)
	case UserInstruction:
		generateUserInstruction(&typedInst, buf)
	case WorkdirInstruction:
		generateWorkdirInstruction(&typedInst, buf)
	case LabelInstruction:
		generateLabelInstruction(&typedInst, buf, escapeChar)
	case StopsignalInstruction:
		generateStopsignalInstruction(&typedInst, buf)
	case HealthcheckInstruction:
		generateHealthcheckInstruction(&typedInst, buf, escapeChar)
	case ShellInstruction:
		generateShellInstruction(&typedInst, buf, escapeChar)
	case OnbuildInstruction:
		generateOnbuildInstruction(&typedInst, buf, escapeChar)
	case MaintainerInstruction:
		generateMaintainerInstruction(&typedInst, buf)
	}
}

// generateRunInstruction generates a RUN instruction.
func generateRunInstruction(run *RunInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("RUN")

	// Mount options
	for _, mount := range run.Mount {
		buf.WriteString(" --mount=" + formatMountOption(&mount))
	}

	// Network
	if run.Network != "" {
		buf.WriteString(fmt.Sprintf(" --network=%s", run.Network))
	}

	// Security
	if run.Security != "" {
		buf.WriteString(fmt.Sprintf(" --security=%s", run.Security))
	}

	// Device
	for _, device := range run.Device {
		buf.WriteString(fmt.Sprintf(" --device=%s", device))
	}

	// Command
	if run.Command != nil {
		buf.WriteString(" ")
		generateCommand(run.Command, buf, escapeChar)
	}
}

// formatMountOption formats a mount option.
func formatMountOption(mount *MountOption) string {
	var parts []string
	parts = append(parts, "type="+mount.Type)
	parts = append(parts, "target="+mount.Target)

	if mount.Source != "" {
		parts = append(parts, "source="+mount.Source)
	}
	if mount.From != "" {
		parts = append(parts, "from="+mount.From)
	}
	if mount.ID != "" {
		parts = append(parts, "id="+mount.ID)
	}
	if mount.Mode != "" {
		parts = append(parts, "mode="+mount.Mode)
	}
	if mount.UID != nil {
		parts = append(parts, fmt.Sprintf("uid=%d", *mount.UID))
	}
	if mount.GID != nil {
		parts = append(parts, fmt.Sprintf("gid=%d", *mount.GID))
	}
	if mount.Sharing != "" {
		parts = append(parts, "sharing="+mount.Sharing)
	}
	if mount.Readonly {
		parts = append(parts, "readonly=true")
	}
	if mount.Required {
		parts = append(parts, "required=true")
	}

	return strings.Join(parts, ",")
}

// generateCmdInstruction generates a CMD instruction.
func generateCmdInstruction(cmd *CmdInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("CMD")
	if cmd.Command != nil {
		buf.WriteString(" ")
		generateCommand(cmd.Command, buf, escapeChar)
	}
}

// generateEntrypointInstruction generates an ENTRYPOINT instruction.
func generateEntrypointInstruction(entrypoint *EntrypointInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("ENTRYPOINT")
	if entrypoint.Command != nil {
		buf.WriteString(" ")
		generateCommand(entrypoint.Command, buf, escapeChar)
	}
}

// generateCopyInstruction generates a COPY instruction.
func generateCopyInstruction(copy *CopyInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("COPY")

	// Options
	if copy.From != "" {
		buf.WriteString(fmt.Sprintf(" --from=%s", copy.From))
	}
	if copy.Chown != "" {
		buf.WriteString(fmt.Sprintf(" --chown=%s", copy.Chown))
	}
	if copy.Chmod != "" {
		buf.WriteString(fmt.Sprintf(" --chmod=%s", copy.Chmod))
	}
	if copy.Link {
		buf.WriteString(" --link")
	}
	if copy.Parents {
		buf.WriteString(" --parents")
	}

	// Sources
	for _, source := range copy.Sources {
		buf.WriteString(" " + escapeShellArg(source, escapeChar))
	}

	// Destination
	buf.WriteString(" " + escapeShellArg(copy.Destination, escapeChar))
}

// generateAddInstruction generates an ADD instruction.
func generateAddInstruction(add *AddInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("ADD")

	// Options
	if add.Checksum != "" {
		buf.WriteString(fmt.Sprintf(" --checksum=%s", add.Checksum))
	}
	if add.Chown != "" {
		buf.WriteString(fmt.Sprintf(" --chown=%s", add.Chown))
	}
	if add.KeepGitDir {
		buf.WriteString(" --keep-git-dir")
	}
	if add.Unpack {
		buf.WriteString(" --unpack")
	}

	// Sources
	for _, source := range add.Sources {
		buf.WriteString(" " + escapeShellArg(source, escapeChar))
	}

	// Destination
	buf.WriteString(" " + escapeShellArg(add.Destination, escapeChar))
}

// generateEnvInstruction generates an ENV instruction.
func generateEnvInstruction(env *EnvInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("ENV")

	// Sort keys for deterministic output
	keys := make([]string, 0, len(env.Variables))
	for k := range env.Variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := env.Variables[key]
		buf.WriteString(" " + key + "=" + escapeEnvValue(value, escapeChar))
	}
}

// generateArgInstruction generates an ARG instruction.
func generateArgInstruction(arg *ArgInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("ARG " + arg.Name)
	if arg.DefaultValue != "" {
		buf.WriteString("=" + escapeShellArg(arg.DefaultValue, escapeChar))
	}
}

// generateExposeInstruction generates an EXPOSE instruction.
func generateExposeInstruction(expose *ExposeInstruction, buf *strings.Builder) {
	buf.WriteString("EXPOSE")

	for _, port := range expose.Ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = "tcp" // default
		}
		buf.WriteString(fmt.Sprintf(" %d/%s", port.Port, protocol))
	}
}

// generateVolumeInstruction generates a VOLUME instruction.
func generateVolumeInstruction(volume *VolumeInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("VOLUME")

	// Use JSON form if paths contain spaces or special characters
	useJSONForm := false
	for _, path := range volume.Paths {
		if strings.Contains(path, " ") || strings.ContainsAny(path, "\\$`\"") {
			useJSONForm = true
			break
		}
	}

	if useJSONForm {
		// JSON form
		buf.WriteString(" [")
		for i, path := range volume.Paths {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(`"` + escapeJSONString(path) + `"`)
		}
		buf.WriteString("]")
	} else {
		// Shell form
		for _, path := range volume.Paths {
			buf.WriteString(" " + path)
		}
	}
}

// generateUserInstruction generates a USER instruction.
func generateUserInstruction(user *UserInstruction, buf *strings.Builder) {
	buf.WriteString("USER " + user.User)
	if user.Group != "" {
		buf.WriteString(":" + user.Group)
	}
}

// generateWorkdirInstruction generates a WORKDIR instruction.
func generateWorkdirInstruction(workdir *WorkdirInstruction, buf *strings.Builder) {
	buf.WriteString("WORKDIR " + workdir.Path)
}

// generateLabelInstruction generates a LABEL instruction.
func generateLabelInstruction(label *LabelInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("LABEL")

	// Sort keys for deterministic output
	keys := make([]string, 0, len(label.Labels))
	for k := range label.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := label.Labels[key]
		buf.WriteString(" " + key + "=" + escapeLabelValue(value, escapeChar))
	}
}

// generateStopsignalInstruction generates a STOPSIGNAL instruction.
func generateStopsignalInstruction(stopsignal *StopsignalInstruction, buf *strings.Builder) {
	buf.WriteString("STOPSIGNAL " + stopsignal.Signal)
}

// generateHealthcheckInstruction generates a HEALTHCHECK instruction.
func generateHealthcheckInstruction(healthcheck *HealthcheckInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("HEALTHCHECK")

	// Options
	if healthcheck.Interval != "" {
		buf.WriteString(fmt.Sprintf(" --interval=%s", healthcheck.Interval))
	}
	if healthcheck.Timeout != "" {
		buf.WriteString(fmt.Sprintf(" --timeout=%s", healthcheck.Timeout))
	}
	if healthcheck.StartPeriod != "" {
		buf.WriteString(fmt.Sprintf(" --start-period=%s", healthcheck.StartPeriod))
	}
	if healthcheck.StartInterval != "" {
		buf.WriteString(fmt.Sprintf(" --start-interval=%s", healthcheck.StartInterval))
	}
	if healthcheck.Retries != nil {
		buf.WriteString(fmt.Sprintf(" --retries=%d", *healthcheck.Retries))
	}

	// Command or NONE
	if healthcheck.None {
		buf.WriteString(" NONE")
	} else if healthcheck.Command != nil {
		buf.WriteString(" CMD ")
		generateCommand(healthcheck.Command, buf, escapeChar)
	}
}

// generateShellInstruction generates a SHELL instruction.
func generateShellInstruction(shell *ShellInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("SHELL [")
	for i, s := range shell.Shell {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(`"` + escapeJSONString(s) + `"`)
	}
	buf.WriteString("]")
}

// generateOnbuildInstruction generates an ONBUILD instruction.
func generateOnbuildInstruction(onbuild *OnbuildInstruction, buf *strings.Builder, escapeChar string) {
	buf.WriteString("ONBUILD ")
	generateInstruction(onbuild.Instruction, buf, escapeChar)
}

// generateMaintainerInstruction generates a MAINTAINER instruction.
func generateMaintainerInstruction(maintainer *MaintainerInstruction, buf *strings.Builder) {
	buf.WriteString("MAINTAINER " + maintainer.Name)
}

// generateCommand generates a command in either shell or exec form.
func generateCommand(cmd interface{}, buf *strings.Builder, escapeChar string) {
	switch v := cmd.(type) {
	case string:
		// Shell form
		buf.WriteString(escapeShellArg(v, escapeChar))
	case []interface{}:
		// Exec form (JSON array)
		buf.WriteString("[")
		for i, item := range v {
			if i > 0 {
				buf.WriteString(", ")
			}
			if str, ok := item.(string); ok {
				buf.WriteString(`"` + escapeJSONString(str) + `"`)
			} else {
				// Fallback for non-string items
				buf.WriteString(`"` + fmt.Sprintf("%v", item) + `"`)
			}
		}
		buf.WriteString("]")
	default:
		// Try to convert to string
		buf.WriteString(escapeShellArg(fmt.Sprintf("%v", v), escapeChar))
	}
}

// escapeShellArg escapes a shell argument.
func escapeShellArg(s string, escapeChar string) string {
	// Basic escaping - escape special characters
	s = strings.ReplaceAll(s, "\\", escapeChar+"\\")
	s = strings.ReplaceAll(s, "$", escapeChar+"$")
	s = strings.ReplaceAll(s, "`", escapeChar+"`")
	s = strings.ReplaceAll(s, "\"", escapeChar+"\"")
	s = strings.ReplaceAll(s, "\n", escapeChar+"n")
	s = strings.ReplaceAll(s, "\r", escapeChar+"r")
	s = strings.ReplaceAll(s, "\t", escapeChar+"t")

	// Quote if contains spaces
	if strings.Contains(s, " ") {
		return `"` + s + `"`
	}
	return s
}

// escapeJSONString escapes a string for JSON/exec form.
func escapeJSONString(s string) string {
	// Escape backslashes first
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Escape double quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")
	// Escape newlines
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// escapeEnvValue escapes an ENV value.
func escapeEnvValue(value string, escapeChar string) string {
	// Quote if contains spaces or special characters
	if strings.ContainsAny(value, " $`\"\\") {
		// Escape $ if literal
		value = strings.ReplaceAll(value, "$", escapeChar+"$")
		return `"` + value + `"`
	}
	return value
}

// escapeLabelValue escapes a LABEL value.
func escapeLabelValue(value string, escapeChar string) string {
	// Quote if contains spaces or special characters
	if strings.ContainsAny(value, " $`\"\\") {
		// Escape $ if literal
		value = strings.ReplaceAll(value, "$", escapeChar+"$")
		return `"` + value + `"`
	}
	return value
}
