package version

// NOTE: these variables are injected at build time

var (
	version, gitSHA, buildTime string
	managerImage               string // optional override for manager image
	pluginRegistry             string // optional override for plugin registry
)
