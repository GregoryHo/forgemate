package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Layout captures canonical state paths for one ForgeMate agent.
type Layout struct {
	RootDir          string
	ConfigDir        string
	ConfigFile       string
	AgentDir         string
	SessionsDir      string
	SessionStoreFile string
	MemoryDir        string
	MemoryDBFile     string
	RunDir           string
}

// ValidateAgentID rejects agent IDs that could escape the state directory.
func ValidateAgentID(agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID must not be empty")
	}
	if strings.ContainsAny(agentID, "/\\") || agentID == "." || agentID == ".." || strings.Contains(agentID, "..") {
		return fmt.Errorf("agent ID %q contains invalid path characters", agentID)
	}
	return nil
}

// ResolveLayout returns deterministic file-backed state paths.
// The caller must validate agentID with ValidateAgentID first.
func ResolveLayout(rootDir string, agentID string) Layout {
	agentDir := filepath.Join(rootDir, "agents", agentID)
	return Layout{
		RootDir:          rootDir,
		ConfigDir:        filepath.Join(rootDir, "config"),
		ConfigFile:       filepath.Join(rootDir, "config", "forgemate.json5"),
		AgentDir:         agentDir,
		SessionsDir:      filepath.Join(agentDir, "sessions"),
		SessionStoreFile: filepath.Join(agentDir, "sessions", "sessions.json"),
		MemoryDir:        filepath.Join(agentDir, "memory"),
		MemoryDBFile:     filepath.Join(agentDir, "memory", "memory.sqlite"),
		RunDir:           filepath.Join(rootDir, "run"),
	}
}

// EnsureLayout creates required directories and enforces secure permissions.
func EnsureLayout(layout Layout) error {
	dirs := []string{
		layout.RootDir,
		layout.ConfigDir,
		layout.AgentDir,
		layout.SessionsDir,
		layout.MemoryDir,
		layout.RunDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create state dir %s: %w", dir, err)
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return fmt.Errorf("chmod state dir %s: %w", dir, err)
		}
	}
	return nil
}
