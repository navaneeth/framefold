package framefold

// Version information
var (
	Version    = "0.1.2"
	Name       = "Framefold"
	CommitHash = "unknown" // Will be set during build
)

// VersionInfo returns a formatted version string
func VersionInfo() string {
	return Name + " version " + Version + " (" + CommitHash + ")"
}
