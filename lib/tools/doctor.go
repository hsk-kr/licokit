package tools

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/hsk-kr/licokit/lib/config"
)

type repoRisk struct {
	Path   string
	Reason string
}

// Doctor checks whether LicoKit describes the machine and, with reset enabled,
// whether Git repositories contain local work that would be lost in an erase.
// It does not fetch, commit, push, or otherwise modify project repositories.
func Doctor(cfg *config.Config, profile string, reset bool) error {
	if !ValidProfile(profile) {
		return fmt.Errorf("unknown profile %q (use core, personal, or all)", profile)
	}
	RefreshHomebrewEnvironment()

	warnings := 0
	check := func(ok bool, success, failure string) {
		if ok {
			fmt.Println("✓ " + success)
			return
		}
		warnings++
		fmt.Println("! " + failure)
	}

	fmt.Println("LicoKit doctor")
	check(runtime.GOOS == "darwin" && runtime.GOARCH == "arm64",
		"Apple Silicon macOS", "LicoKit currently targets Apple Silicon macOS")
	check(exec.Command("xcode-select", "-p").Run() == nil,
		"Command Line Tools are installed", "Command Line Tools are not installed")
	check(ExistCommand("brew"), "Homebrew is available", "Homebrew is not available")

	missing := make([]string, 0)
	for _, tool := range cfg.Tools {
		if !tool.EnabledForProfile(profile) {
			continue
		}
		installed, err := IsInstalled(tool)
		if err != nil || !installed {
			missing = append(missing, tool.Name)
		}
	}
	if len(missing) == 0 {
		fmt.Printf("✓ All %s-profile software is installed\n", profile)
	} else {
		warnings++
		fmt.Printf("! %d %s-profile item(s) will be installed or repaired: %s\n",
			len(missing), profile, strings.Join(missing, ", "))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	doctorDotfiles(cfg.Dotfiles, home, check)
	doctorDevelopmentTools(home, check)
	check(hasSSHPrivateKey(filepath.Join(home, ".ssh")),
		"An SSH private key exists", "No SSH private key found; restore or recreate GitHub SSH credentials")
	check(ExecCommandQuiet("gh", "auth", "status") == nil,
		"GitHub CLI authentication works", "GitHub CLI needs authentication after the reset")

	risks, scanErrors := inspectProjectRepos([]string{
		filepath.Join(home, "dev"),
		filepath.Join(home, "hobby"),
		filepath.Join(home, "licokit"),
	})
	for _, scanErr := range scanErrors {
		warnings++
		fmt.Println("! " + scanErr.Error())
	}
	if len(risks) == 0 {
		fmt.Println("✓ Project repositories have no dirty or unpushed work (based on local tracking refs)")
	} else {
		warnings += len(risks)
		fmt.Println("! Project work is not fully represented by local upstream refs:")
		for _, risk := range risks {
			fmt.Printf("  - %s: %s\n", displayHomePath(risk.Path, home), risk.Reason)
		}
	}

	if runtime.GOOS == "darwin" {
		timeMachineOutput, timeMachineErr := commandOutput("tmutil", "destinationinfo")
		hasTimeMachine := hasTimeMachineDestination(timeMachineOutput, timeMachineErr)
		check(hasTimeMachine,
			"A Time Machine destination is configured", "No Time Machine destination detected (informational; projects are expected in GitHub)")
	}
	check(fileExists(filepath.Join(home, "Library", "Mobile Documents", "com~apple~CloudDocs")),
		"iCloud Drive storage is present", "iCloud Drive storage is not present on this Mac")

	fmt.Println("i After reinstall, macOS will still require sign-in/permission prompts for App Store apps, GitHub/SSH, Docker, VPN, Accessibility, Screen Recording, and Full Disk Access.")
	if warnings == 0 {
		fmt.Println("✓ No issues found")
	} else {
		fmt.Printf("Doctor completed with %d warning(s).\n", warnings)
	}

	if reset && (len(risks) > 0 || len(scanErrors) > 0) {
		return fmt.Errorf("reset check failed: resolve the repository warnings before erasing this Mac")
	}
	return nil
}

func doctorDotfiles(dotCfg config.DotfilesConfig, home string, check func(bool, string, string)) {
	repoPath := config.ExpandPath(dotCfg.RepoPath)
	if repoPath == "" {
		repoPath = defaultDotfilesRepoPath(home)
	}
	_, err := os.Stat(filepath.Join(repoPath, ".git"))
	check(err == nil, "LicoKit dotfiles checkout exists", "LicoKit dotfiles checkout will be created during install")
	if err == nil {
		status, statusErr := commandOutput("git", "-C", repoPath, "status", "--porcelain")
		check(statusErr == nil && strings.TrimSpace(status) == "",
			"LicoKit dotfiles checkout is clean", "LicoKit dotfiles checkout has local changes")
	}

	for _, item := range dotCfg.ConfigLinks {
		target := filepath.Join(home, ".config", item)
		check(isSymlink(target), "Managed link exists: ~/.config/"+item,
			"Managed link will be created: ~/.config/"+item)
	}
}

func doctorDevelopmentTools(home string, check func(bool, string, string)) {
	missing := make([]string, 0)
	require := func(ok bool, name string) {
		if !ok {
			missing = append(missing, name)
		}
	}

	for _, major := range []string{"24", "26"} {
		matches, _ := filepath.Glob(filepath.Join(home, ".nvm", "versions", "node", "v"+major+".*"))
		require(len(matches) > 0, "Node "+major)
	}
	defaultNode, _ := os.ReadFile(filepath.Join(home, ".nvm", "alias", "default"))
	require(strings.HasPrefix(strings.TrimSpace(string(defaultNode)), "24"), "default Node 24")
	for _, command := range []string{"npm", "corepack", "pnpm", "yarn", "wrangler"} {
		require(ExistCommand(command), command)
	}
	if pnpmVersion, err := commandOutput("pnpm", "--version"); err == nil {
		require(strings.TrimSpace(pnpmVersion) == "10.19.0", "pnpm 10.19.0")
	}

	goBin := ""
	if gopath, err := commandOutput("go", "env", "GOPATH"); err == nil {
		goBin = filepath.Join(strings.TrimSpace(gopath), "bin")
	}
	for _, tool := range []string{
		"air", "asmfmt", "dlv", "errcheck", "fillstruct", "godef", "goimports", "golangci-lint",
		"gomodifytags", "gopls", "gotags", "iferr", "impl", "motion", "revive", "staticcheck",
	} {
		require(goBin != "" && fileExists(filepath.Join(goBin, tool)), "Go tool "+tool)
	}

	masonPackages := filepath.Join(home, ".local", "share", "nvim", "mason", "packages")
	for _, tool := range []string{
		"bash-language-server", "css-lsp", "delve", "gopls", "harper-ls", "html-lsp", "json-lsp",
		"lua-language-server", "prettier", "prettierd", "pyright", "python-lsp-server", "ruby-lsp",
		"tailwindcss-language-server", "typescript-language-server", "yaml-language-server",
	} {
		require(fileExists(filepath.Join(masonPackages, tool)), "Mason "+tool)
	}

	require(fileExists(filepath.Join(home, ".tmux", "plugins", "tpm", ".git")), "tmux plugin manager")
	require(fileExists(filepath.Join(home, ".config", "opencode", "node_modules", "@opencode-ai", "plugin")), "OpenCode plugin dependencies")

	if output, err := commandOutput("claude", "plugin", "list"); err == nil {
		for _, plugin := range []string{
			"context7@claude-plugins-official",
			"gopls-lsp@claude-plugins-official",
			"claude-mem@thedotmack",
			"github@claude-plugins-official",
			"typescript-lsp@claude-plugins-official",
			"everything-claude-code@everything-claude-code",
		} {
			require(strings.Contains(output, plugin), "Claude plugin "+plugin)
		}
	} else {
		missing = append(missing, "Claude plugin state")
	}

	check(len(missing) == 0,
		"Language runtimes, global tools, editor tooling, and plugins are complete",
		fmt.Sprintf("%d development component(s) will be installed or repaired: %s", len(missing), strings.Join(missing, ", ")))
}

func inspectProjectRepos(roots []string) ([]repoRisk, []error) {
	repos := make([]string, 0)
	errorsFound := make([]error, 0)
	for _, root := range roots {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Name() == ".git" {
				repos = append(repos, filepath.Dir(path))
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !entry.IsDir() {
				return nil
			}
			if path != root && shouldSkipRepoScan(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			errorsFound = append(errorsFound, fmt.Errorf("scan %s: %w", root, err))
		}
	}

	sort.Strings(repos)
	risks := make([]repoRisk, 0)
	for _, repo := range repos {
		status, err := commandOutput("git", "-C", repo, "status", "--porcelain")
		if err != nil {
			risks = append(risks, repoRisk{repo, "could not inspect working tree"})
			continue
		}
		if strings.TrimSpace(status) != "" {
			risks = append(risks, repoRisk{repo, "uncommitted or untracked files"})
		}

		if _, err := commandOutput("git", "-C", repo, "rev-parse", "--verify", "HEAD"); err != nil {
			risks = append(risks, repoRisk{repo, "repository has no commit to restore"})
			continue
		}
		if _, err := commandOutput("git", "-C", repo, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err != nil {
			risks = append(risks, repoRisk{repo, "current branch has no upstream"})
			continue
		}
		aheadText, err := commandOutput("git", "-C", repo, "rev-list", "--count", "@{upstream}..HEAD")
		if err != nil {
			risks = append(risks, repoRisk{repo, "could not compare the current branch with its upstream"})
			continue
		}
		ahead, _ := strconv.Atoi(strings.TrimSpace(aheadText))
		if ahead > 0 {
			risks = append(risks, repoRisk{repo, fmt.Sprintf("%d unpushed commit(s) on the current branch", ahead)})
		}

		localOnlyText, err := commandOutput("git", "-C", repo, "rev-list", "--count", "--branches", "--not", "--remotes")
		if err == nil {
			localOnly, _ := strconv.Atoi(strings.TrimSpace(localOnlyText))
			if localOnly > ahead {
				risks = append(risks, repoRisk{repo, fmt.Sprintf("%d commit(s) across local branches are not in any locally known remote ref", localOnly)})
			}
		}
	}
	return risks, errorsFound
}

func shouldSkipRepoScan(name string) bool {
	switch name {
	case "node_modules", "vendor", ".cache", ".venv", "venv", "dist", "build":
		return true
	default:
		return false
	}
}

func hasSSHPrivateKey(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasSuffix(name, ".pub") || !strings.HasPrefix(name, "id_") {
			continue
		}
		return true
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasTimeMachineDestination(output string, err error) bool {
	return err == nil && !strings.Contains(output, "No destinations configured") && strings.TrimSpace(output) != ""
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

func displayHomePath(path, home string) string {
	if relative, err := filepath.Rel(home, path); err == nil && !strings.HasPrefix(relative, "..") {
		return "~/" + relative
	}
	return path
}
