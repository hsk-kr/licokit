package tools

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hsk-kr/licokit/lib/styles"
)

func RenderItem(name string, disabled bool) {
	if disabled {
		fmt.Print(styles.ItemNameDisabled.Render(name))
		fmt.Print(styles.StatusInstalled.Render(" ✓ Installed"))
	} else {
		fmt.Print(styles.ItemName.Render(name))
		fmt.Print(styles.StatusNotInstalled.Render(" ✗ Not Installed"))
	}
}

func ExistCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func ExistBrewPackage(packageName string) bool {
	brew, err := exec.LookPath("brew")
	if err != nil {
		return false
	}
	_, err = exec.Command(brew, "list", packageName).Output()

	if err != nil {
		return false
	}

	return true
}

func ExistBrewTap(tap string) bool {
	brew, err := exec.LookPath("brew")
	if err != nil {
		return false
	}
	output, err := exec.Command(brew, "tap").Output()
	if err != nil {
		return false
	}
	for _, installed := range strings.Fields(string(output)) {
		if installed == tap {
			return true
		}
	}
	return false
}

func ExistApplication(appName string) bool {
	paths := []string{filepath.Join("/Applications", appName)}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, "Applications", appName))
	}
	for _, appPath := range paths {
		if _, err := os.Stat(appPath); err == nil {
			return true
		}
	}
	return false
}

// RefreshHomebrewEnvironment makes a newly installed Homebrew available to the
// running process. This avoids requiring a terminal restart in the middle of a
// full bootstrap.
func RefreshHomebrewEnvironment() {
	for _, candidate := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		prefix := filepath.Dir(filepath.Dir(candidate))
		path := os.Getenv("PATH")
		bin := filepath.Join(prefix, "bin")
		sbin := filepath.Join(prefix, "sbin")
		if !pathContains(path, bin) {
			path = bin + string(os.PathListSeparator) + path
		}
		if !pathContains(path, sbin) {
			path = sbin + string(os.PathListSeparator) + path
		}
		_ = os.Setenv("PATH", path)
		_ = os.Setenv("HOMEBREW_PREFIX", prefix)
		return
	}
}

func pathContains(path, candidate string) bool {
	for _, item := range filepath.SplitList(path) {
		if item == candidate {
			return true
		}
	}
	return false
}

func ExecCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", command, err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s failed: %w", command, err)
	}

	return nil
}

// ExecCommandQuiet runs a command without printing stdout/stderr.
// Used when a spinner is showing progress instead.
func ExecCommandQuiet(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", command, err)
	}

	if err := cmd.Wait(); err != nil {
		detail := strings.TrimSpace(output.String())
		if detail != "" {
			return fmt.Errorf("%s failed: %w: %s", command, err, detail)
		}
		return fmt.Errorf("%s failed: %w", command, err)
	}

	return nil
}

func commandOutput(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail != "" {
			return "", fmt.Errorf("%s failed: %w: %s", command, err, detail)
		}
		return "", fmt.Errorf("%s failed: %w", command, err)
	}
	return string(output), nil
}

/*
Add a source line to ~/.config/licokit/sources.zsh.

It will insert a small managed block into .zshrc if it is not set up.

If the source exists in the dev.zsh, it will be ignored.
*/
func AddZshSource(source string) error {
	homePath, err := os.UserHomeDir()

	if err != nil {
		WarningMessage(err.Error())
		return err
	}

	licokitConfigPath := filepath.Join(homePath, ".config", "licokit")
	licokitZshPath := filepath.Join(licokitConfigPath, "sources.zsh")
	zshrcPath := filepath.Join(homePath, ".zshrc")
	if err := os.MkdirAll(licokitConfigPath, 0o755); err != nil {
		return err
	}

	// add source into dev.zsh. if it's already setup, it does nothing
	addSourceToDevZsh := func() error {
		exist, err := existFile(licokitZshPath)

		if err != nil {
			return err
		}

		if exist {
			contains, err := containInFile(licokitZshPath, source)

			if err != nil {
				return err
			}

			if contains {
				return nil
			}
		}

		prefix := ""
		if exist {
			prefix = "\n"
		}
		err = appendFile(licokitZshPath, prefix+source+"\n")
		return err
	}

	// add source dev zsh into .zshrc. if it's already setup, it does nothing
	addDevZshToZshrc := func() error {
		exist, err := existFile(zshrcPath)

		if err != nil {
			return err
		}

		managedBlock := fmt.Sprintf("# >>> licokit >>>\n[ -f %q ] && source %q\n# <<< licokit <<<", licokitZshPath, licokitZshPath)
		if exist {
			contains, err := containInFile(zshrcPath, "# >>> licokit >>>")

			if err != nil {
				return err
			}

			if contains {
				return nil
			}
		}

		prefix := ""
		if exist {
			prefix = "\n"
		}
		err = appendFile(zshrcPath, prefix+managedBlock+"\n")
		return err
	}

	err = addSourceToDevZsh()
	if err != nil {
		WarningMessage(err.Error())
		return err
	}

	return addDevZshToZshrc()
}

func existFile(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

/*
Returns if the str exists in the content of the file or not
*/
func containInFile(path, str string) (bool, error) {
	bContent, err := os.ReadFile(path)

	if err != nil {
		return false, err
	}

	return strings.Contains(string(bContent), str), nil
}

func appendFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func WarningMessage(message string) {
	fmt.Println(styles.WarningBox.Render("⚠ " + message))
}

func SuccessMessage(message string) {
	fmt.Println(styles.SuccessBox.Render("✓ " + message))
}
