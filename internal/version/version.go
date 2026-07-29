package version

import "fmt"

var (
	// Version is set via ldflags at build time, e.g.: -ldflags="-X github.com/mediacrunch/mediacrunch/internal/version.Version=v1.0.0"
	Version = "dev"
	// Commit is set via ldflags at build time
	Commit = "unknown"
	// BuildDate is set via ldflags at build time
	BuildDate = "unknown"
)

// String returns a formatted version string.
func String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
}
