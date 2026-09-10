package version

var (
	// Version is injected at build time using -ldflags.
	// Defaults to "dev" when run without build flags.
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)
