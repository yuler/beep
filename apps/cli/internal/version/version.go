package version

var (
	// Version is overridden at build time using -ldflags
	Version   = "0.0.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)
