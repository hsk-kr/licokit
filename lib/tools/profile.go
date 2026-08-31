package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/hsk-kr/licokit/lib/config"
)

func ValidProfile(profile string) bool {
	return profile == "core" || profile == "personal" || profile == "all"
}

func InstallProfile(cfg *config.Config, profile string, dryRun bool) error {
	if !ValidProfile(profile) {
		return fmt.Errorf("unknown profile %q (use core, personal, or all)", profile)
	}
	RefreshHomebrewEnvironment()

	failures := make([]string, 0)
	for _, tool := range cfg.Tools {
		if !tool.EnabledForProfile(profile) {
			continue
		}
		installed, err := IsInstalled(tool)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s detection: %v", tool.Name, err))
			continue
		}
		if installed {
			fmt.Printf("✓ %-28s already installed\n", tool.Name)
			continue
		}
		if dryRun {
			fmt.Printf("→ %-28s %s\n", tool.Name, installDescription(tool))
			continue
		}

		if err := Install(tool); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", tool.Name, err))
			WarningMessage(fmt.Sprintf("%s failed: %v", tool.Name, err))
			if tool.DetectType == "xcode" || tool.DetectValue == "brew" {
				break
			}
			continue
		}
		RefreshHomebrewEnvironment()

		// xcode-select --install returns after opening Apple's installer. Do not
		// cascade into Homebrew failures while the installation is still pending.
		if tool.DetectType == "xcode" {
			ready, _ := IsInstalled(tool)
			if !ready {
				return fmt.Errorf("Command Line Tools installation is pending; finish the macOS dialog, then rerun licokit install")
			}
		}
	}

	if dryRun {
		fmt.Println("→ Dotfiles                     sync safely and run post-install setup")
		if cfg.Services.CPUKiller {
			fmt.Println("→ CPU Killer                   enable LaunchAgent")
		}
		return nil
	}

	if !ExistCommand("brew") {
		return combineFailures(failures, "Homebrew is unavailable; rerun after the prerequisite installation finishes")
	}

	if err := SetupDotfiles(cfg.Dotfiles); err != nil {
		failures = append(failures, "dotfiles: "+err.Error())
	}
	if cfg.Services.CPUKiller {
		if err := EnableCPUKiller(); err != nil {
			failures = append(failures, "CPU Killer: "+err.Error())
		}
	}
	return combineFailures(failures, "")
}

func UpdateProfile(cfg *config.Config, profile string, dryRun bool) error {
	if !ValidProfile(profile) {
		return fmt.Errorf("unknown profile %q (use core, personal, or all)", profile)
	}
	RefreshHomebrewEnvironment()
	failures := make([]string, 0)

	brewCommands := [][]string{
		{"update"},
		{"upgrade", "--formula"},
		{"upgrade", "--cask", "--greedy"},
	}
	for _, args := range brewCommands {
		if dryRun {
			fmt.Println("→ brew " + strings.Join(args, " "))
			continue
		}
		if !ExistCommand("brew") {
			failures = append(failures, "Homebrew is not installed")
			break
		}
		if err := ExecCommand("brew", args...); err != nil {
			failures = append(failures, "brew "+strings.Join(args, " ")+": "+err.Error())
		}
	}

	installedNow := make(map[string]bool)
	for _, tool := range cfg.Tools {
		if !tool.EnabledForProfile(profile) {
			continue
		}
		installed, err := IsInstalled(tool)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s detection: %v", tool.Name, err))
			continue
		}
		if installed {
			continue
		}
		if dryRun {
			fmt.Printf("→ repair %-24s %s\n", tool.Name, installDescription(tool))
			continue
		}
		fmt.Printf("Repairing %s...\n", tool.Name)
		if err := Install(tool); err != nil {
			failures = append(failures, tool.Name+": "+err.Error())
			continue
		}
		installedNow[tool.Name] = true
	}

	if profile != "core" && ExistCommand("mas") {
		if dryRun {
			fmt.Println("→ mas upgrade")
		} else if err := ExecCommand("mas", "upgrade"); err != nil {
			failures = append(failures, "mas upgrade: "+err.Error())
		}
	}

	for _, tool := range cfg.Tools {
		if !tool.EnabledForProfile(profile) || tool.UpdateCommand == "" || installedNow[tool.Name] {
			continue
		}
		installed, _ := IsInstalled(tool)
		if !installed {
			continue
		}
		if dryRun {
			fmt.Printf("→ update %s\n", tool.Name)
			continue
		}
		fmt.Printf("Updating %s...\n", tool.Name)
		if err := ExecCommand("bash", "-c", tool.UpdateCommand); err != nil {
			failures = append(failures, tool.Name+": "+err.Error())
		}
	}

	if dryRun {
		fmt.Println("→ update dotfiles and post-install tools")
		return nil
	}

	_ = os.Setenv("LICOKIT_UPDATE", "1")
	err := SetupDotfiles(cfg.Dotfiles)
	_ = os.Unsetenv("LICOKIT_UPDATE")
	if err != nil {
		failures = append(failures, "dotfiles: "+err.Error())
	}
	if cfg.Services.CPUKiller {
		if err := EnableCPUKiller(); err != nil {
			failures = append(failures, "CPU Killer: "+err.Error())
		}
	}
	return combineFailures(failures, "")
}

func installDescription(tool config.ToolConfig) string {
	switch tool.InstallType {
	case "brew":
		return "brew install " + tool.BrewPackage()
	case "cask":
		return "brew install --cask " + tool.BrewPackage()
	case "tap":
		return "brew tap " + tool.Package
	case "mas":
		return "mas install " + tool.Package
	case "script":
		return "run official installer"
	case "manual":
		return tool.ManualMessage
	default:
		return tool.InstallType
	}
}

func combineFailures(failures []string, final string) error {
	if final != "" {
		failures = append(failures, final)
	}
	if len(failures) == 0 {
		return nil
	}
	return fmt.Errorf("%d step(s) need attention:\n- %s", len(failures), strings.Join(failures, "\n- "))
}
