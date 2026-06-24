package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

const cpuKillerLabel = "com.lico.cpu-killer"

// cpuKillerPlist is the LaunchAgent definition. It runs the watchdog at login
// (RunAtLoad) and keeps it alive 24/7 (KeepAlive). launchd starts agents with
// an empty environment, so HOME is set explicitly — otherwise the script can't
// find the user config at ~/.config/cpu-killer/config. Placeholders: label,
// script, home.
const cpuKillerPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
      <string>/bin/bash</string>
      <string>%s</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
      <key>HOME</key>
      <string>%s</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/cpu-killer.out.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/cpu-killer.err.log</string>
  </dict>
</plist>
`

// cpuKillerScript resolves the watchdog script path, preferring the dotfiles
// symlink in ~/scripts and falling back to the repo copy under ~/licokit.
func cpuKillerScript(home string) (string, error) {
	candidates := []string{
		filepath.Join(home, "scripts", "cpu-killer.sh"),
		filepath.Join(home, "licokit", "dotfiles", "scripts", "cpu-killer.sh"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("cpu-killer.sh not found — run Dotfiles setup first (looked in %v)", candidates)
}

func cpuKillerPlistPath(home string) string {
	return filepath.Join(home, "Library", "LaunchAgents", cpuKillerLabel+".plist")
}

func cpuKillerDomain() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}

// EnableCPUKiller installs the LaunchAgent and starts the watchdog immediately.
// RunAtLoad makes it start now and on every login; KeepAlive restarts it if it
// ever dies — so it runs 24/7 without waiting for a reboot.
func EnableCPUKiller() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	script, err := cpuKillerScript(home)
	if err != nil {
		return err
	}
	if err := os.Chmod(script, 0o755); err != nil {
		return fmt.Errorf("make script executable: %w", err)
	}

	plistPath := cpuKillerPlistPath(home)
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return err
	}
	plist := fmt.Sprintf(cpuKillerPlist, cpuKillerLabel, script, home)
	if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("write LaunchAgent: %w", err)
	}

	// Reload cleanly: boot out any existing instance, then bootstrap. With
	// RunAtLoad the bootstrap also starts it right now.
	domain := cpuKillerDomain()
	_ = ExecCommandQuiet("launchctl", "bootout", domain, plistPath)
	if err := ExecCommandQuiet("launchctl", "bootstrap", domain, plistPath); err != nil {
		return fmt.Errorf("launchctl bootstrap failed: %w", err)
	}
	return nil
}

// DisableCPUKiller stops the watchdog and removes its LaunchAgent.
func DisableCPUKiller() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	plistPath := cpuKillerPlistPath(home)

	_ = ExecCommandQuiet("launchctl", "bootout", cpuKillerDomain(), plistPath)
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove LaunchAgent: %w", err)
	}
	return nil
}

// CPUKillerEnabled reports whether the LaunchAgent is currently loaded.
func CPUKillerEnabled() bool {
	return ExecCommandQuiet("launchctl", "print", cpuKillerDomain()+"/"+cpuKillerLabel) == nil
}
