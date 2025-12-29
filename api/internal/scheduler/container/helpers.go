package container

// IsEnabledOrDefault returns whether the feature is enabled, defaulting to false
// Container operations are disabled by default for safety
func IsEnabledOrDefault(enabled *bool) bool {
	if enabled != nil {
		return *enabled
	}
	return false
}
