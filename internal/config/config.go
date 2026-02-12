package config

import (
	"os"
	"path/filepath"
)

// Config holds gateway runtime configuration resolved from environment.
type Config struct {
	HTTPAddr string
	StateDir string
	AgentID  string
	Sidecar  SidecarConfig
}

// SidecarConfig configures the Node sidecar connection and supervision behavior.
type SidecarConfig struct {
	SocketPath string
	Enabled    bool
}

// Load resolves runtime configuration from environment with deterministic defaults.
func Load() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	stateDir := envOrDefault("FORGEMATE_STATE_DIR", filepath.Join(home, ".forgemate"))
	agentID := envOrDefault("FORGEMATE_AGENT_ID", "main")

	return Config{
		HTTPAddr: envOrDefault("FORGEMATE_HTTP_ADDR", ":18789"),
		StateDir: stateDir,
		AgentID:  agentID,
		Sidecar: SidecarConfig{
			SocketPath: envOrDefault("FORGEMATE_SIDECAR_SOCKET", filepath.Join(stateDir, "run", "agent-runtime.sock")),
			Enabled:    envOrDefault("FORGEMATE_SIDECAR_ENABLED", "1") == "1",
		},
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
