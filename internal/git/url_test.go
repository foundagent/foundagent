package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:     "SSH URL with .git",
			url:      "git@github.com:owner/repo.git",
			expected: "owner/repo.git",
		},
		{
			name:     "SSH URL without .git",
			url:      "git@github.com:owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS URL with .git",
			url:      "https://github.com/owner/repo.git",
			expected: "owner/repo.git",
		},
		{
			name:     "HTTPS URL without .git",
			url:      "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "HTTP URL",
			url:      "http://github.com/owner/repo.git",
			expected: "owner/repo.git",
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
		{
			name:        "invalid URL",
			url:         "not-a-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseURL(tt.url)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInferName(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:     "simple repo name with .git",
			url:      "git@github.com:owner/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "simple repo name without .git",
			url:      "git@github.com:owner/my-repo",
			expected: "my-repo",
		},
		{
			name:     "HTTPS URL",
			url:      "https://github.com/owner/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "nested path",
			url:      "git@github.com:org/team/my-repo.git",
			expected: "my-repo",
		},
		{
			name:        "invalid URL",
			url:         "not-a-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InferName(tt.url)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{
			name:  "valid SSH URL",
			url:   "git@github.com:owner/repo.git",
			valid: true,
		},
		{
			name:  "valid HTTPS URL",
			url:   "https://github.com/owner/repo.git",
			valid: true,
		},
		{
			name:  "invalid URL",
			url:   "not-a-url",
			valid: false,
		},
		{
			name:  "empty URL",
			url:   "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
