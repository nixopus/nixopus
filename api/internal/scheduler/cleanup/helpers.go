package cleanup

// GetRetentionDaysOrDefault returns the retention days from settings or the default value
func GetRetentionDaysOrDefault(days *int, defaultDays int) int {
	if days != nil {
		return *days
	}
	return defaultDays
}

// IsEnabledOrDefault returns whether the feature is enabled, defaulting to true
func IsEnabledOrDefault(enabled *bool) bool {
	if enabled != nil {
		return *enabled
	}
	return true
}
