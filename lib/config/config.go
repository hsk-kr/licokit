package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed default_config.yaml
var defaultConfig embed.FS

type ExtraLink struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type DotfilesConfig struct {
	Repo        string            `yaml:"repo"`
	RepoPath    string            `yaml:"repo_path"`
	ConfigLinks []string          `yaml:"config_links"`
	HomeLinks   map[string]string `yaml:"home_links"`
	ExtraLinks  []ExtraLink       `yaml:"extra_links"`
	ZshSource   string            `yaml:"zsh_source"`
	PostScripts []string          `yaml:"post_scripts"`
}

type ToolConfig struct {
	Name               string   `yaml:"name"`
	InstallType        string   `yaml:"install_type"`
	Package            string   `yaml:"package"`
	Version            string   `yaml:"version"`
	InstallCommand     string   `yaml:"install_command"`
	DetectType         string   `yaml:"detect_type"`
	DetectValue        string   `yaml:"detect_value"`
	ManualMessage      string   `yaml:"manual_message"`
	ZshSource          string   `yaml:"zsh_source"`
	PostInstallWarning string   `yaml:"post_install_warning"`
	PostInstallDirs    []string `yaml:"post_install_dirs"`
	PostInstallScripts []string `yaml:"post_install_scripts"`
	UpdateCommand      string   `yaml:"update_command"`
	Profiles           []string `yaml:"profiles"`
	Interactive        bool     `yaml:"interactive"`
}

// BrewPackage returns the brew package name with version pinning if specified.
func (t *ToolConfig) BrewPackage() string {
	if t.Version != "" {
		return fmt.Sprintf("%s@%s", t.Package, t.Version)
	}
	return t.Package
}

type Config struct {
	Dotfiles DotfilesConfig `yaml:"dotfiles"`
	Tools    []ToolConfig   `yaml:"tools"`
	Services ServicesConfig `yaml:"services"`
}

type ServicesConfig struct {
	CPUKiller bool `yaml:"cpu_killer"`
}

// EnabledForProfile reports whether a tool belongs to the requested install
// profile. The personal profile is a superset of core; all includes every
// declared tool. Tools without a profile remain enabled for compatibility with
// older user configs.
func (t ToolConfig) EnabledForProfile(profile string) bool {
	if profile == "all" || len(t.Profiles) == 0 {
		return true
	}

	for _, candidate := range t.Profiles {
		if candidate == profile || (profile == "personal" && candidate == "core") {
			return true
		}
	}
	return false
}

// Load reads the config from the user config path if it exists,
// otherwise falls back to the embedded default config.
func Load() (*Config, error) {
	// Check for user config at ~/.config/licokit/config.yaml
	homePath, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(homePath, ".config", "licokit", "config.yaml")
		if data, err := os.ReadFile(userConfigPath); err == nil {
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("failed to parse user config: %w", err)
			}
			return &cfg, nil
		}
	}

	// Fall back to embedded default
	data, err := defaultConfig.ReadFile("default_config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse default config: %w", err)
	}

	return &cfg, nil
}

func parseYAML(data []byte, cfg *Config) error {
	return yaml.Unmarshal(data, cfg)
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
