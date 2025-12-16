package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	info := Get()

	// Version fields should be populated (even if "dev"/"unknown")
	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.Commit)
	assert.NotEmpty(t, info.BuildDate)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
}

func TestString(t *testing.T) {
	// Save original Version
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		name     string
		version  string
		contains []string
	}{
		{
			name:     "dev version",
			version:  "dev",
			contains: []string{"foundagent", "dev", "commit"},
		},
		{
			name:     "release version",
			version:  "1.2.3",
			contains: []string{"foundagent", "v1.2.3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			str := String()

			for _, substr := range tt.contains {
				assert.Contains(t, str, substr)
			}
		})
	}
}

func TestFull(t *testing.T) {
	full := Full()

	// Should contain all expected fields
	assert.Contains(t, full, "foundagent")
	assert.Contains(t, full, "Commit:")
	assert.Contains(t, full, "Build Date:")
	assert.Contains(t, full, "Go Version:")
	assert.Contains(t, full, "Platform:")
}
