package shared

import (
	"github.com/hashicorp/go-plugin"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion: 1,

	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "SCHEMAHERO_PLUGIN",
	MagicCookieValue: "schemahero-database-plugin",
}

// PluginMap is the map of plugins we can dispense.
// This will be populated by the plugin package to avoid circular dependencies.
var PluginMap = map[string]plugin.Plugin{}

// Interface name constant for database plugins
const DatabaseInterfaceName = "database"
