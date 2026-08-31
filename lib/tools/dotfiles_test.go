package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceWithSymlinkPreservesExistingTarget(t *testing.T) {
	home := t.TempDir()
	source := filepath.Join(home, "repo", "dotfiles", "nvim")
	target := filepath.Join(home, ".config", "nvim")
	backupRoot := filepath.Join(home, ".local", "state", "licokit", "backups", "test")

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "init.lua"), []byte("user config"), 0o644); err != nil {
		t.Fatal(err)
	}

	moved, err := replaceWithSymlink(source, target, home, backupRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !moved || !isSymlink(target) {
		t.Fatal("expected the existing directory to be preserved and replaced by a symlink")
	}
	backup := filepath.Join(backupRoot, ".config", "nvim", "init.lua")
	data, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(data) != "user config" {
		t.Fatalf("backup content = %q", data)
	}
}

func TestReplaceWithSymlinkIsIdempotent(t *testing.T) {
	home := t.TempDir()
	source := filepath.Join(home, "repo", "dotfiles", "tmux")
	target := filepath.Join(home, ".config", "tmux")
	backupRoot := filepath.Join(home, "backups", "test")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}

	moved, err := replaceWithSymlink(source, target, home, backupRoot)
	if err != nil {
		t.Fatal(err)
	}
	if moved {
		t.Fatal("an existing correct symlink should not create a backup")
	}
	if _, err := os.Stat(backupRoot); !os.IsNotExist(err) {
		t.Fatalf("unexpected backup directory: %v", err)
	}
}

func TestReplaceWithSymlinkRejectsMissingSource(t *testing.T) {
	home := t.TempDir()
	_, err := replaceWithSymlink(
		filepath.Join(home, "missing"),
		filepath.Join(home, ".config", "missing"),
		home,
		filepath.Join(home, "backups"),
	)
	if err == nil {
		t.Fatal("expected an error for a missing source")
	}
}
