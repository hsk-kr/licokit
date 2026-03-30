package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SelectProjectDir uses fzf to let the user pick a git project directory.
// Returns the selected path or an error.
func SelectProjectDir() (string, error) {
	if !ExistCommand("fzf") {
		return "", fmt.Errorf("fzf is required. Install it from the Tools menu")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Find git repos up to 4 levels deep, pipe to fzf
	findCmd := fmt.Sprintf("find %s -maxdepth 4 -type d -name .git 2>/dev/null | sed 's/\\/.git$//' | sort", home)
	cmd := exec.Command("bash", "-c", findCmd+" | fzf --prompt='Select project: '")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("project selection cancelled")
	}

	dir := strings.TrimSpace(string(output))
	if dir == "" {
		return "", fmt.Errorf("no directory selected")
	}

	// Validate .git exists
	gitPath := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", fmt.Errorf("selected directory does not contain a git repository")
	}

	return dir, nil
}

// AgentPane holds the info needed to launch one pane.
type AgentPane struct {
	Name   string
	Prompt string // full prompt text, empty for Free role
}

// LaunchAgentSession creates a tmux session with panes for each agent.
// Each pane runs claude with the given prompt in the project directory.
func LaunchAgentSession(projectDir string, agents []AgentPane) error {
	if !ExistCommand("tmux") {
		return fmt.Errorf("tmux is required. Install it from the Tools menu")
	}
	if !ExistCommand("claude") {
		return fmt.Errorf("claude is required. Install it from the Tools menu")
	}

	// Write prompts to temp files to avoid command-line length limits
	promptFiles, err := writePromptFiles(agents)
	if err != nil {
		return fmt.Errorf("failed to write prompt files: %w", err)
	}
	// Note: temp files are NOT cleaned up — each tmux pane needs them at launch time.
	// They live in os.TempDir() and will be cleaned by the OS.

	agentCount := len(agents)

	if os.Getenv("TMUX") != "" {
		// Inside tmux — create one new window with panes for each agent
		firstCmd := buildClaudeCommand(projectDir, promptFiles[0])
		err = ExecCommandQuiet("tmux", "new-window", "-n", "claude-agents", firstCmd)
		if err != nil {
			return fmt.Errorf("failed to create window: %w", err)
		}
		_ = ExecCommandQuiet("tmux", "set-option", "-p", "@agent-role", agents[0].Name)
		_ = ExecCommandQuiet("tmux", "set-option", "-p", "@agent-status", "working")

		for i := 1; i < agentCount; i++ {
			cmd := buildClaudeCommand(projectDir, promptFiles[i])
			err := ExecCommandQuiet("tmux", "split-window", cmd)
			if err != nil {
				return fmt.Errorf("failed to create pane for %s: %w", agents[i].Name, err)
			}
			_ = ExecCommandQuiet("tmux", "set-option", "-p", "@agent-role", agents[i].Name)
			_ = ExecCommandQuiet("tmux", "set-option", "-p", "@agent-status", "working")
			_ = ExecCommandQuiet("tmux", "select-layout", "tiled")
		}

		// Show role names in pane borders (per-pane user option, can't be overridden)
		_ = ExecCommandQuiet("tmux", "set-option", "-w", "pane-border-status", "top")
		_ = ExecCommandQuiet("tmux", "set-option", "-w", "pane-border-format", "#{?#{==:#{@agent-status},attention},#[fg=colour220] ● #{@agent-role} ,#[fg=colour40] ● #{@agent-role} }")
		// Select first pane
		_ = ExecCommandQuiet("tmux", "select-pane", "-t", "0")
		return nil
	}

	// Not inside tmux — create a new session
	sessionName := generateSessionName(projectDir)

	firstCmd := buildClaudeCommand(projectDir, promptFiles[0])
	err = ExecCommandQuiet("tmux", "new-session", "-d", "-s", sessionName, "-n", "claude-agents", firstCmd)
	if err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}
	_ = ExecCommandQuiet("tmux", "set-option", "-t", sessionName, "-p", "@agent-role", agents[0].Name)

	for i := 1; i < agentCount; i++ {
		cmd := buildClaudeCommand(projectDir, promptFiles[i])
		err := ExecCommandQuiet("tmux", "split-window", "-t", sessionName, cmd)
		if err != nil {
			return fmt.Errorf("failed to create pane for %s: %w", agents[i].Name, err)
		}
		_ = ExecCommandQuiet("tmux", "set-option", "-p", "@agent-role", agents[i].Name)
		_ = ExecCommandQuiet("tmux", "select-layout", "-t", sessionName, "tiled")
	}

	_ = ExecCommandQuiet("tmux", "set-option", "-t", sessionName, "pane-border-status", "top")
	_ = ExecCommandQuiet("tmux", "set-option", "-t", sessionName, "pane-border-format", "#{?#{==:#{@agent-status},attention},#[fg=colour220] ● #{@agent-role} ,#[fg=colour40] ● #{@agent-role} }")

	attachCmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr
	return attachCmd.Run()
}

// writePromptFiles writes each agent's prompt to a temp file and returns the file paths.
// Returns empty string for agents with no prompt (Free role).
func writePromptFiles(agents []AgentPane) ([]string, error) {
	paths := make([]string, len(agents))
	for i, agent := range agents {
		if agent.Prompt == "" {
			paths[i] = ""
			continue
		}
		safeName := strings.ReplaceAll(agent.Name, "/", "-")
		f, err := os.CreateTemp("", fmt.Sprintf("claude-agent-%s-*.md", safeName))
		if err != nil {
			return nil, err
		}
		if _, err := f.WriteString(agent.Prompt); err != nil {
			f.Close()
			return nil, err
		}
		f.Close()
		paths[i] = f.Name()
	}
	return paths, nil
}

func buildClaudeCommand(projectDir string, promptFile string) string {
	if promptFile == "" {
		return fmt.Sprintf("cd %s && claude --dangerously-skip-permissions", shellEscape(projectDir))
	}
	return fmt.Sprintf("cd %s && claude --dangerously-skip-permissions --append-system-prompt-file %s", shellEscape(projectDir), shellEscape(promptFile))
}

func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func generateSessionName(projectDir string) string {
	base := filepath.Base(projectDir) + "-claude-agents"

	// Check if session already exists, append number if so
	name := base
	for i := 2; ; i++ {
		err := ExecCommandQuiet("tmux", "has-session", "-t", name)
		if err != nil {
			// Session doesn't exist, use this name
			return name
		}
		name = fmt.Sprintf("%s-%d", base, i)
	}
}

