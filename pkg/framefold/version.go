package framefold

// Version information
var (
	Version    = "dev"     // Will be set during build from git tag
	Name       = "Framefold"
	CommitHash = "unknown" // Will be set during build
)

// VersionInfo returns a formatted version string
func VersionInfo() string {
	return Name + " version " + Version + " (" + CommitHash + ")"
}
