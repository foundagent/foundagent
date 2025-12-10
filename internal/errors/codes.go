package errors

// Error codes for Foundagent
const (
	// Configuration errors (E0xx)
	ErrCodeWorkspaceExists     = "E001" // Workspace already exists
	ErrCodeInvalidName         = "E002" // Invalid workspace name
	ErrCodePathTooLong         = "E003" // Path exceeds OS limits
	ErrCodeInvalidConfig       = "E004" // Invalid configuration file
	ErrCodeConfigNotFound      = "E005" // Configuration file not found
	ErrCodeInvalidInput        = "E006" // Invalid input provided

	// Filesystem errors (E1xx)
	ErrCodePermissionDenied    = "E101" // Permission denied
	ErrCodeDiskFull            = "E102" // Disk full
	ErrCodeFileNotFound        = "E103" // File not found
	ErrCodeDirectoryNotEmpty   = "E104" // Directory not empty

	// Git errors (E2xx)
	ErrCodeGitNotInstalled     = "E201" // Git not installed
	ErrCodeGitOperationFailed  = "E202" // Git operation failed
	ErrCodeInvalidRepository   = "E203" // Invalid git repository

	// Worktree errors (E3xx)
	ErrCodeWorktreeExists      = "E301" // Worktree already exists
	ErrCodeWorktreeNotFound    = "E302" // Worktree not found
	ErrCodeBranchExists        = "E303" // Branch already exists
	ErrCodeBranchNotFound      = "E304" // Branch not found
	ErrCodeInvalidOperation    = "E305" // Invalid operation (e.g., removing worktree you're in)

	// Network errors (E4xx)
	ErrCodeNetworkError        = "E401" // Network operation failed
	ErrCodeAuthenticationFailed = "E402" // Authentication failed

	// General errors (E9xx)
	ErrCodeUnknown             = "E999" // Unknown error
)
