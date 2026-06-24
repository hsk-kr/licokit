package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrewPackage_WithVersion(t *testing.T) {
	tool := ToolConfig{
		Package: "go",
		Version: "1.23",
	}
	got := tool.BrewPackage()
	want := "go@1.23"
	if got != want {
		t.Errorf("BrewPackage() = %q, want %q", got, want)
	}
}

func TestBrewPackage_WithoutVersion(t *testing.T) {
	tool := ToolConfig{
		Package: "git",
	}
	got := tool.BrewPackage()
	want := "git"
	if got != want {
		t.Errorf("BrewPackage() = %q, want %q", got, want)
	}
}

func TestExpandPath_WithTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	got := ExpandPath("~/.config/nvim")
	want := filepath.Join(home, ".config/nvim")
	if got != want {
		t.Errorf("ExpandPath(\"~/.config/nvim\") = %q, want %q", got, want)
	}
}

func TestExpandPath_WithoutTilde(t *testing.T) {
	got := ExpandPath("/usr/local/bin")
	want := "/usr/local/bin"
	if got != want {
		t.Errorf("ExpandPath(\"/usr/local/bin\") = %q, want %q", got, want)
	}
}

func TestExpandPath_EmptyString(t *testing.T) {
	got := ExpandPath("")
	want := ""
	if got != want {
		t.Errorf("ExpandPath(\"\") = %q, want %q", got, want)
	}
}

func TestLoad_DefaultConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Dotfiles.ConfigLinks) == 0 {
		t.Error("Dotfiles.ConfigLinks should not be empty")
	}

	if len(cfg.Tools) == 0 {
		t.Error("Tools should not be empty")
	}

	// Verify all tools have required fields
	for _, tool := range cfg.Tools {
		if tool.Name == "" {
			t.Error("Tool name should not be empty")
		}
		if tool.InstallType == "" {
			t.Errorf("Tool %q: install_type should not be empty", tool.Name)
		}
		if tool.DetectType == "" {
			t.Errorf("Tool %q: detect_type should not be empty", tool.Name)
		}
		if tool.DetectValue == "" {
			t.Errorf("Tool %q: detect_value should not be empty", tool.Name)
		}
	}
}

func TestLoad_DefaultConfig_ToolCount(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Default config should have 20 tools
	if len(cfg.Tools) != 20 {
		t.Errorf("expected 20 tools, got %d", len(cfg.Tools))
	}
}

func TestLoad_DefaultConfig_DotfilesConfigLinks(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	expectedLinks := []string{"aerospace", "devdeck", "karabiner", "neru", "nvim", "tmux", "zsh", "alacritty", "ghostty"}
	if len(cfg.Dotfiles.ConfigLinks) != len(expectedLinks) {
		t.Errorf("expected %d config links, got %d", len(expectedLinks), len(cfg.Dotfiles.ConfigLinks))
	}

	for i, link := range expectedLinks {
		if i < len(cfg.Dotfiles.ConfigLinks) && cfg.Dotfiles.ConfigLinks[i] != link {
			t.Errorf("config link %d: got %q, want %q", i, cfg.Dotfiles.ConfigLinks[i], link)
		}
	}
}

func TestLoad_DefaultConfig_DoesNotLinkClaudeGlobalConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, link := range cfg.Dotfiles.ExtraLinks {
		if strings.HasPrefix(link.Source, "claude/") || strings.Contains(link.Target, ".claude") {
			t.Errorf("default dotfiles should not link Claude global config: source=%q target=%q", link.Source, link.Target)
		}
	}

	for _, script := range cfg.Dotfiles.PostScripts {
		if strings.HasPrefix(script, "claude/") {
			t.Errorf("default dotfiles should not run Claude setup scripts: %q", script)
		}
	}
}

func TestLoad_UserConfig(t *testing.T) {
	// Create a temporary user config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "licokit")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	userConfig := `
dotfiles:
  repo: "git@github.com:test-user/dotfiles.git"
  config_links:
    - nvim
tools:
  - name: TestTool
    install_type: brew
    package: test-tool
    detect_type: command
    detect_value: test-tool
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(userConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// We can't easily override UserHomeDir, so just test the YAML parsing directly
	var cfg Config
	if err := parseYAML([]byte(userConfig), &cfg); err != nil {
		t.Fatalf("parseYAML error: %v", err)
	}

	if cfg.Dotfiles.Repo != "git@github.com:test-user/dotfiles.git" {
		t.Errorf("repo = %q, want test-user repo", cfg.Dotfiles.Repo)
	}
	if len(cfg.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(cfg.Tools))
	}
	if cfg.Tools[0].Name != "TestTool" {
		t.Errorf("tool name = %q, want TestTool", cfg.Tools[0].Name)
	}
}

func TestLoad_InstallTypes(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	validInstallTypes := map[string]bool{"brew": true, "cask": true, "manual": true, "script": true}
	for _, tool := range cfg.Tools {
		if !validInstallTypes[tool.InstallType] {
			t.Errorf("Tool %q has invalid install_type: %q", tool.Name, tool.InstallType)
		}
	}
}

func TestLoad_DetectTypes(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	validDetectTypes := map[string]bool{"command": true, "application": true, "brew_package": true}
	for _, tool := range cfg.Tools {
		if !validDetectTypes[tool.DetectType] {
			t.Errorf("Tool %q has invalid detect_type: %q", tool.Name, tool.DetectType)
		}
	}
}

func TestLoad_ManualToolHasMessage(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, tool := range cfg.Tools {
		if tool.InstallType == "manual" && tool.ManualMessage == "" {
			t.Errorf("Tool %q has install_type=manual but no manual_message", tool.Name)
		}
	}
}

func TestLoad_ScriptToolHasCommand(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, tool := range cfg.Tools {
		if tool.InstallType == "script" && tool.InstallCommand == "" {
			t.Errorf("Tool %q has install_type=script but no install_command", tool.Name)
		}
	}
}

func TestLoad_BrewToolHasPackage(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, tool := range cfg.Tools {
		if (tool.InstallType == "brew" || tool.InstallType == "cask") && tool.Package == "" {
			t.Errorf("Tool %q has install_type=%s but no package", tool.Name, tool.InstallType)
		}
	}
}
