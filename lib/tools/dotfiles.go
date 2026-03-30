package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hsk-kr/licokit/lib/config"
	"github.com/hsk-kr/licokit/lib/spinner"
)

func SetupDotfiles(dotCfg config.DotfilesConfig) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		WarningMessage(err.Error())
		return err
	}

	repo := "https://github.com/hsk-kr/licokit.git"
	licokitHomePath := filepath.Join(homePath, "licokit")
	configDirPath := filepath.Join(homePath, ".config")
	dotfilesPath := filepath.Join(licokitHomePath, "dotfiles")

	// Clone or pull the licokit repo
	if _, err := os.Stat(filepath.Join(licokitHomePath, ".git")); os.IsNotExist(err) {
		sp := spinner.New("Cloning licokit...")
		sp.Start()
		err = ExecCommandQuiet("git", "clone", repo, licokitHomePath)
		sp.Stop()
		if err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	} else {
		sp := spinner.New("Updating licokit...")
		sp.Start()
		err = ExecCommandQuiet("git", "-C", licokitHomePath, "fetch", "origin")
		if err == nil {
			err = ExecCommandQuiet("git", "-C", licokitHomePath, "reset", "--hard", "origin/main")
		}
		sp.Stop()
		if err != nil {
			WarningMessage(fmt.Sprintf("git pull failed (offline?): %s", err.Error()))
		}
	}

	if err := ExecCommand("mkdir", "-p", configDirPath); err != nil {
		return err
	}

	// Symlink config directories
	for _, item := range dotCfg.ConfigLinks {
		target := filepath.Join(configDirPath, item)
		// Remove existing directory (not symlink) so ln doesn't create a link inside it
		if info, err := os.Lstat(target); err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			if err := os.RemoveAll(target); err != nil {
				return err
			}
		}
		if err := ExecCommand("ln", "-sfn", filepath.Join(dotfilesPath, item), target); err != nil {
			return err
		}
	}

	// Symlink home directories
	for source, target := range dotCfg.HomeLinks {
		if err := ExecCommand("ln", "-sfn", filepath.Join(dotfilesPath, source), filepath.Join(homePath, target)); err != nil {
			return err
		}
	}

	// Extra links (e.g., claude/skills -> ~/.claude/skills)
	for _, link := range dotCfg.ExtraLinks {
		targetPath := config.ExpandPath(link.Target)
		targetDir := filepath.Dir(targetPath)
		if err := ExecCommand("mkdir", "-p", targetDir); err != nil {
			return err
		}
		if err := ExecCommand("ln", "-sfn", filepath.Join(dotfilesPath, link.Source), targetPath); err != nil {
			return err
		}
	}

	// Run post-setup scripts
	for _, script := range dotCfg.PostScripts {
		scriptPath := filepath.Join(dotfilesPath, script)
		sp := spinner.New(fmt.Sprintf("Running %s...", script))
		sp.Start()
		err := ExecCommandQuiet("bash", scriptPath)
		sp.Stop()
		if err != nil {
			WarningMessage(fmt.Sprintf("Post script %s failed: %s", script, err.Error()))
		}
	}

	// Add zsh source
	if dotCfg.ZshSource != "" {
		zshSource := dotCfg.ZshSource
		// Replace ~ with $HOME so the path stays portable across users
		if len(zshSource) > 0 && zshSource[0] == '~' {
			zshSource = "$HOME" + zshSource[1:]
		}
		return AddZshSource(fmt.Sprintf("source %s", zshSource))
	}

	return nil
}
