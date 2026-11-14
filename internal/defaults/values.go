// Package defaults is the single source of truth for tuned application-wide
// default values: CLI defaults, server listen/shutdown settings, and config paths.
package defaults

import (
	"time"

	"github.com/kthcloud/go-deploy/models/mode"
)

// CLI and environment defaults.
const (
	// DefaultMode is the default application runtime mode.
	DefaultMode = mode.Dev
	// DefaultConfigFile is the fallback configuration file path when DEPLOY_CONFIG_FILE is unset.
	DefaultConfigFile = "config.local.yml"
)

// HTTP server defaults.
const (
	// DefaultListenHost is the bind host for the API HTTP server.
	DefaultListenHost = "0.0.0.0"
	// DefaultGinMode is the fallback Gin mode when GIN_MODE is unset.
	DefaultGinMode = "release"
	// HTTPShutdownTimeout bounds graceful HTTP server shutdown.
	HTTPShutdownTimeout = 5 * time.Second
)
