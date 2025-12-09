package errors

// Error codes for Foundagent
const (
	// Configuration errors (E0xx)
	ErrCodeWorkspaceExists     = "E001" // Workspace already exists
	ErrCodeInvalidName         = "E002" // Invalid workspace name
	ErrCodePathTooLong         = "E003" // Path exceeds OS limits
	ErrCodeInvalidConfig       = "E004" // Invalid configuration file
	ErrCodeConfigNotFound      = "E005" // Configuration file not found

	// Filesystem errors (E1xx)
	ErrCodePermissionDenied    = "E101" // Permission denied
	ErrCodeDiskFull            = "E102" // Disk full
	ErrCodeFileNotFound        = "E103" // File not found
	ErrCodeDirectoryNotEmpty   = "E104" // Directory not empty

	// Git errors (E2xx)
	ErrCodeGitNotInstalled     = "E201" // Git not installed
	ErrCodeGitOperationFailed  = "E202" // Git operation failed
	ErrCodeInvalidRepository   = "E203" // Invalid git repository

	// General errors (E9xx)
	ErrCodeUnknown             = "E999" // Unknown error
)
