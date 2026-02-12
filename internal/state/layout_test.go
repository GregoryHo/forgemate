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
