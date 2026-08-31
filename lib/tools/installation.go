package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hsk-kr/licokit/lib/config"
	"github.com/hsk-kr/licokit/lib/spinner"
)

func Install(tool config.ToolConfig) error {
	switch tool.InstallType {
	case "manual":
		SuccessMessage(tool.ManualMessage)
		return nil
	case "brew":
		sp := spinner.New(fmt.Sprintf("Installing %s...", tool.Name))
		sp.Start()
		err := ExecCommandQuiet("brew", "install", tool.BrewPackage())
		sp.Stop()
		if err != nil {
			return err
		}
	case "cask":
		sp := spinner.New(fmt.Sprintf("Installing %s...", tool.Name))
		sp.Start()
		err := ExecCommandQuiet("brew", "install", "--cask", tool.BrewPackage())
		sp.Stop()
		if err != nil {
			return err
		}
	case "tap":
		sp := spinner.New(fmt.Sprintf("Adding %s...", tool.Name))
		sp.Start()
		err := ExecCommandQuiet("brew", "tap", tool.Package)
		sp.Stop()
		if err != nil {
			return err
		}
	case "mas":
		sp := spinner.New(fmt.Sprintf("Installing %s...", tool.Name))
		sp.Start()
		err := ExecCommandQuiet("mas", "install", tool.Package)
		sp.Stop()
		if err != nil {
			return err
		}
	case "script":
		var err error
		if tool.Interactive {
			fmt.Printf("Installing %s...\n", tool.Name)
			err = ExecCommand("bash", "-c", tool.InstallCommand)
		} else {
			sp := spinner.New(fmt.Sprintf("Installing %s...", tool.Name))
			sp.Start()
			err = ExecCommandQuiet("bash", "-c", tool.InstallCommand)
			sp.Stop()
		}
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown install type: %s", tool.InstallType)
	}
	RefreshHomebrewEnvironment()

	// Create post-install directories
	for _, dir := range tool.PostInstallDirs {
		expanded := config.ExpandPath(dir)
		if err := os.MkdirAll(expanded, 0755); err != nil {
			return err
		}
	}

	// Add zsh source if specified
	if tool.ZshSource != "" {
		if err := AddZshSource(tool.ZshSource); err != nil {
			return err
		}
	}

	// Run post-install scripts
	for _, script := range tool.PostInstallScripts {
		expanded := config.ExpandPath(script)
		sp := spinner.New(fmt.Sprintf("Running %s...", filepath.Base(expanded)))
		sp.Start()
		err := ExecCommandQuiet("bash", expanded)
		sp.Stop()
		if err != nil {
			WarningMessage(fmt.Sprintf("Post script %s failed: %s", filepath.Base(expanded), err.Error()))
		}
	}

	// Show post-install warning if specified
	if tool.PostInstallWarning != "" {
		WarningMessage(tool.PostInstallWarning)
	}

	return nil
}

func IsInstalled(tool config.ToolConfig) (bool, error) {
	switch tool.DetectType {
	case "command":
		return ExistCommand(tool.DetectValue), nil
	case "application":
		return ExistApplication(tool.DetectValue), nil
	case "brew_package":
		return ExistBrewPackage(tool.DetectValue), nil
	case "brew_tap":
		return ExistBrewTap(tool.DetectValue), nil
	case "package_receipt":
		return exec.Command("pkgutil", "--pkg-info", tool.DetectValue).Run() == nil, nil
	case "xcode":
		return exec.Command("xcode-select", "-p").Run() == nil, nil
	default:
		return false, fmt.Errorf("unknown detect type: %s", tool.DetectType)
	}
}
