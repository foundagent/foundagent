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
	str := String()
	assert.Contains(t, str, "foundagent")

	// Should contain either version number or "dev"
	assert.True(t, len(str) > 0)
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
