package doctor

// Status represents the status of a check
type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// Check represents a single diagnostic check
type Check interface {
	Name() string
	Run() CheckResult
}

// CheckResult represents the result of running a check
type CheckResult struct {
	Name        string `json:"name"`
	Status      Status `json:"status"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
	Fixable     bool   `json:"fixable"`
}

// IsSuccess returns true if the check passed
func (r CheckResult) IsSuccess() bool {
	return r.Status == StatusPass || r.Status == StatusWarn
}
