package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hsk-kr/licokit/lib/config"
	"github.com/hsk-kr/licokit/lib/spinner"
)

const defaultDotfilesRepo = "https://github.com/hsk-kr/licokit.git"

func defaultDotfilesRepoPath(home string) string {
	return filepath.Join(home, ".local", "share", "licokit", "repo")
}

func SetupDotfiles(dotCfg config.DotfilesConfig) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	repo := dotCfg.Repo
	if repo == "" {
		repo = defaultDotfilesRepo
	}
	repoPath := config.ExpandPath(dotCfg.RepoPath)
	if repoPath == "" {
		repoPath = defaultDotfilesRepoPath(homePath)
	}

	if err := syncDotfilesRepo(repo, repoPath); err != nil {
		return err
	}

	dotfilesPath := filepath.Join(repoPath, "dotfiles")
	configDirPath := filepath.Join(homePath, ".config")
	if err := os.MkdirAll(configDirPath, 0o755); err != nil {
		return err
	}

	backupRoot := filepath.Join(
		homePath,
		".local",
		"state",
		"licokit",
		"backups",
		time.Now().Format("20060102-150405"),
	)
	backedUp := false

	for _, item := range dotCfg.ConfigLinks {
		source := filepath.Join(dotfilesPath, item)
		target := filepath.Join(configDirPath, item)
		moved, err := replaceWithSymlink(source, target, homePath, backupRoot)
		if err != nil {
			return fmt.Errorf("link %s: %w", item, err)
		}
		backedUp = backedUp || moved
	}

	for source, target := range dotCfg.HomeLinks {
		moved, err := replaceWithSymlink(
			filepath.Join(dotfilesPath, source),
			filepath.Join(homePath, target),
			homePath,
			backupRoot,
		)
		if err != nil {
			return fmt.Errorf("link %s: %w", target, err)
		}
		backedUp = backedUp || moved
	}

	for _, link := range dotCfg.ExtraLinks {
		moved, err := replaceWithSymlink(
			filepath.Join(dotfilesPath, link.Source),
			config.ExpandPath(link.Target),
			homePath,
			backupRoot,
		)
		if err != nil {
			return fmt.Errorf("link %s: %w", link.Target, err)
		}
		backedUp = backedUp || moved
	}

	if backedUp {
		SuccessMessage("Existing configuration was preserved at " + backupRoot)
	}

	postFailures := make([]string, 0)
	for _, script := range dotCfg.PostScripts {
		scriptPath := filepath.Join(dotfilesPath, script)
		if _, err := os.Stat(scriptPath); err != nil {
			WarningMessage(fmt.Sprintf("Post script %s is missing", script))
			postFailures = append(postFailures, script+": missing")
			continue
		}
		sp := spinner.New(fmt.Sprintf("Running %s...", script))
		sp.Start()
		err := ExecCommandQuiet("bash", scriptPath)
		sp.Stop()
		if err != nil {
			WarningMessage(fmt.Sprintf("Post script %s failed: %s", script, err.Error()))
			postFailures = append(postFailures, script+": "+err.Error())
		}
	}

	if dotCfg.ZshSource != "" {
		zshSource := dotCfg.ZshSource
		if strings.HasPrefix(zshSource, "~/") {
			zshSource = "$HOME/" + strings.TrimPrefix(zshSource, "~/")
		}
		if err := AddZshSource(fmt.Sprintf("source %s", zshSource)); err != nil {
			postFailures = append(postFailures, "zsh source: "+err.Error())
		}
	}

	if status, err := commandOutput("git", "-C", repoPath, "status", "--porcelain"); err != nil {
		postFailures = append(postFailures, "verify dotfiles checkout: "+err.Error())
	} else if strings.TrimSpace(status) != "" {
		postFailures = append(postFailures, "post-install setup modified the managed LicoKit checkout; review "+repoPath)
	}

	return combineFailures(postFailures, "")
}

func syncDotfilesRepo(repo, repoPath string) error {
	gitPath := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		if entries, readErr := os.ReadDir(repoPath); readErr == nil && len(entries) > 0 {
			return fmt.Errorf("dotfiles destination %s exists and is not an empty Git repository", repoPath)
		}
		if err := os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
			return err
		}
		sp := spinner.New("Cloning licokit dotfiles...")
		sp.Start()
		err := ExecCommandQuiet("git", "clone", repo, repoPath)
		sp.Stop()
		if err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
		return nil
	}

	status, err := commandOutput("git", "-C", repoPath, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("inspect dotfiles repository: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("dotfiles repository has local changes at %s; commit or discard them before updating", repoPath)
	}

	sp := spinner.New("Updating licokit dotfiles...")
	sp.Start()
	err = ExecCommandQuiet("git", "-C", repoPath, "pull", "--ff-only", "origin", "main")
	sp.Stop()
	if err != nil {
		WarningMessage(fmt.Sprintf("dotfiles update failed; using the existing checkout: %s", err.Error()))
	}
	return nil
}

func replaceWithSymlink(source, target, home, backupRoot string) (bool, error) {
	if _, err := os.Stat(source); err != nil {
		return false, fmt.Errorf("source does not exist: %s", source)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return false, err
	}

	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return false, os.Symlink(source, target)
	}
	if err != nil {
		return false, err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		current, readErr := os.Readlink(target)
		if readErr == nil && current == source {
			return false, nil
		}
		if err := os.Remove(target); err != nil {
			return false, err
		}
		return false, os.Symlink(source, target)
	}

	relative, err := filepath.Rel(home, target)
	if err != nil || strings.HasPrefix(relative, "..") {
		relative = filepath.Base(target)
	}
	backupPath := filepath.Join(backupRoot, relative)
	if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
		return false, err
	}
	if err := os.Rename(target, backupPath); err != nil {
		return false, fmt.Errorf("back up %s: %w", target, err)
	}
	if err := os.Symlink(source, target); err != nil {
		_ = os.Rename(backupPath, target)
		return false, err
	}
	return true, nil
}
