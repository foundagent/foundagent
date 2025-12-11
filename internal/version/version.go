package version

import (
	"fmt"
	"runtime"
)

// Build-time variables injected via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

// Info represents complete version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the full version information
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a simple version string
func String() string {
	if Version == "dev" {
		return fmt.Sprintf("foundagent %s (commit %s)", Version, Commit)
	}
	return fmt.Sprintf("foundagent v%s", Version)
}

// Full returns a detailed version string with all build info
func Full() string {
	info := Get()
	return fmt.Sprintf(`foundagent v%s
Commit:     %s
Build Date: %s
Go Version: %s
Platform:   %s/%s`,
		info.Version,
		info.Commit,
		info.BuildDate,
		info.GoVersion,
		info.OS,
		info.Arch,
	)
}
