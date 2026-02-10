package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultCLIPath = "sigrok-cli"
	defaultTimeout = 30 * time.Second
)

// Config holds runtime configuration for the sigrok MCP server.
type Config struct {
	SigrokCLIPath string
	Timeout       time.Duration
	WorkingDir    string
}

// Load reads configuration from environment variables, falling back to defaults.
func Load() Config {
	cfg := Config{
		SigrokCLIPath: defaultCLIPath,
		Timeout:       defaultTimeout,
	}

	if v := os.Getenv("SIGROK_CLI_PATH"); v != "" {
		cfg.SigrokCLIPath = v
	}

	if v := os.Getenv("SIGROK_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Timeout = time.Duration(n) * time.Second
		}
	}

	if v := os.Getenv("SIGROK_WORKING_DIR"); v != "" {
		cfg.WorkingDir = v
	}

	return cfg
}

// FirmwareDirs returns the standard sigrok firmware search directories.
// These match the paths that libsigrok searches at runtime.
func FirmwareDirs() []string {
	dirs := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "share", "sigrok-firmware"))
	}
	dirs = append(dirs,
		"/usr/local/share/sigrok-firmware",
		"/usr/share/sigrok-firmware",
	)
	return dirs
}
