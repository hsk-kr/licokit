package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectProjectReposFindsDirtyAndUnpushedWork(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	repo := filepath.Join(root, "project")
	runTestGit(t, "init", "--bare", remote)
	runTestGit(t, "clone", remote, repo)
	runTestGit(t, "-C", repo, "config", "user.name", "LicoKit Test")
	runTestGit(t, "-C", repo, "config", "user.email", "licokit@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("initial\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, "-C", repo, "add", "README.md")
	runTestGit(t, "-C", repo, "commit", "-m", "initial")
	runTestGit(t, "-C", repo, "push", "-u", "origin", "HEAD")

	risks, scanErrors := inspectProjectRepos([]string{root})
	if len(scanErrors) != 0 || len(risks) != 0 {
		t.Fatalf("clean pushed repo returned risks=%v errors=%v", risks, scanErrors)
	}

	if err := os.WriteFile(filepath.Join(repo, "local.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	risks, _ = inspectProjectRepos([]string{root})
	if !riskContains(risks, "uncommitted or untracked") {
		t.Fatalf("expected dirty work risk, got %#v", risks)
	}

	runTestGit(t, "-C", repo, "add", "local.txt")
	runTestGit(t, "-C", repo, "commit", "-m", "local only")
	risks, _ = inspectProjectRepos([]string{root})
	if !riskContains(risks, "unpushed commit") {
		t.Fatalf("expected unpushed commit risk, got %#v", risks)
	}
}

func TestHasSSHPrivateKey(t *testing.T) {
	dir := t.TempDir()
	if hasSSHPrivateKey(dir) {
		t.Fatal("empty directory should not contain a private key")
	}
	if err := os.WriteFile(filepath.Join(dir, "id_ed25519.pub"), []byte("public"), 0o644); err != nil {
		t.Fatal(err)
	}
	if hasSSHPrivateKey(dir) {
		t.Fatal("a public key alone is not a private key")
	}
	if err := os.WriteFile(filepath.Join(dir, "id_ed25519"), []byte("private"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !hasSSHPrivateKey(dir) {
		t.Fatal("expected private key detection")
	}
}

func TestHasTimeMachineDestination(t *testing.T) {
	if hasTimeMachineDestination("tmutil: No destinations configured.\n", nil) {
		t.Fatal("tmutil's successful no-destination response must not count as configured")
	}
	if !hasTimeMachineDestination("Name : Backup\nKind : Local\n", nil) {
		t.Fatal("expected a Time Machine destination")
	}
}

func riskContains(risks []repoRisk, text string) bool {
	for _, risk := range risks {
		if strings.Contains(risk.Reason, text) {
			return true
		}
	}
	return false
}

func runTestGit(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
