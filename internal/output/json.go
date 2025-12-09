package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/errors"
)

// SuccessResponse represents a successful operation output
type SuccessResponse struct {
	Status  string                 `json:"status"`
	Data    map[string]interface{} `json:"data"`
}

// ErrorResponse represents an error output
type ErrorResponse struct {
	Status      string `json:"status"`
	Error       string `json:"error"`
	Code        string `json:"code"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
}

// PrintJSON prints a JSON response to stdout
func PrintJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// PrintSuccess prints a success response
func PrintSuccess(data map[string]interface{}) error {
	return PrintJSON(SuccessResponse{
		Status: "success",
		Data:   data,
	})
}

// PrintError prints an error response
func PrintError(err error) error {
	if faErr, ok := err.(*errors.Error); ok {
		return PrintJSON(ErrorResponse{
			Status:      "error",
			Error:       err.Error(),
			Code:        faErr.Code,
			Message:     faErr.Message,
			Remediation: faErr.Remediation,
		})
	}

	return PrintJSON(ErrorResponse{
		Status:  "error",
		Error:   err.Error(),
		Code:    errors.ErrCodeUnknown,
		Message: err.Error(),
	})
}

// PrintMessage prints a simple text message (non-JSON mode)
func PrintMessage(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// PrintErrorMessage prints an error message to stderr (non-JSON mode)
func PrintErrorMessage(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
