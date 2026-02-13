package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveLayoutDeterministic(t *testing.T) {
	layout := ResolveLayout("/tmp/forgemate", "main")
	if layout.SessionStoreFile != "/tmp/forgemate/agents/main/sessions/sessions.json" {
		t.Fatalf("unexpected session store path: %s", layout.SessionStoreFile)
	}
	if layout.MemoryDBFile != "/tmp/forgemate/agents/main/memory/memory.sqlite" {
		t.Fatalf("unexpected memory db path: %s", layout.MemoryDBFile)
	}
}

func TestValidateAgentID_RejectsTraversal(t *testing.T) {
	bad := []string{"", "..", "../etc", "foo/bar", "a\\b", "..hidden"}
	for _, id := range bad {
		if err := ValidateAgentID(id); err == nil {
			t.Errorf("expected error for agent ID %q, got nil", id)
		}
	}
}

func TestValidateAgentID_AcceptsValid(t *testing.T) {
	good := []string{"main", "agent-1", "my_agent", "agent.v2"}
	for _, id := range good {
		if err := ValidateAgentID(id); err != nil {
			t.Errorf("unexpected error for agent ID %q: %v", id, err)
		}
	}
}

func TestEnsureLayoutCreatesDirs(t *testing.T) {
	root := t.TempDir()
	layout := ResolveLayout(filepath.Join(root, ".forgemate"), "agentA")
	if err := EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}
	for _, dir := range []string{layout.RootDir, layout.ConfigDir, layout.AgentDir, layout.SessionsDir, layout.MemoryDir, layout.RunDir} {
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("expected dir %s to exist: %v", dir, err)
		}
	}
}
