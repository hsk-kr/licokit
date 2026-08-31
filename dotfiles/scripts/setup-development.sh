#!/usr/bin/env bash

# Idempotent post-install setup for tools that Homebrew cannot fully configure.
# Set LICOKIT_UPDATE=1 to refresh global tools as well as install missing ones.
set -u

failures=0
updating="${LICOKIT_UPDATE:-0}"

run_step() {
  local label="$1"
  shift
  printf '  %s... ' "$label"
  if "$@" >/dev/null 2>&1; then
    echo "ok"
  else
    echo "failed"
    failures=$((failures + 1))
  fi
}

if [[ -x /opt/homebrew/bin/brew ]]; then
  eval "$(/opt/homebrew/bin/brew shellenv)"
elif [[ -x /usr/local/bin/brew ]]; then
  eval "$(/usr/local/bin/brew shellenv)"
fi

export PATH="$HOME/.local/bin:$HOME/.bun/bin:$PATH"

printf '%s\n' "Configuring language runtimes and editor tooling"

if command -v brew >/dev/null 2>&1 && [[ -s "$(brew --prefix nvm 2>/dev/null)/nvm.sh" ]]; then
  export NVM_DIR="$HOME/.nvm"
  mkdir -p "$NVM_DIR"
  # shellcheck disable=SC1090
  source "$(brew --prefix nvm)/nvm.sh"
  run_step "Node 24" nvm install 24
  run_step "Node 26" nvm install 26
  run_step "default Node 24" nvm alias default 24
  nvm use 24 >/dev/null 2>&1 || true
else
  echo "  nvm unavailable; skipping Node runtimes"
  failures=$((failures + 1))
fi

if command -v npm >/dev/null 2>&1; then
  if [[ "$updating" == "1" ]] || ! command -v wrangler >/dev/null 2>&1; then
    run_step "npm, Corepack, and Wrangler" npm install --global npm@latest corepack@latest wrangler@latest
  fi
  run_step "enable Corepack" corepack enable
  run_step "pnpm 10.19" corepack prepare pnpm@10.19.0 --activate
  run_step "Yarn 1.22" corepack prepare yarn@1.22.22 --activate
fi

if command -v go >/dev/null 2>&1; then
  go_bin="$(go env GOPATH)/bin"
  go_tools=(
    "air|github.com/air-verse/air@latest"
    "asmfmt|github.com/klauspost/asmfmt/cmd/asmfmt@latest"
    "dlv|github.com/go-delve/delve/cmd/dlv@latest"
    "errcheck|github.com/kisielk/errcheck@latest"
    "fillstruct|github.com/davidrjenni/reftools/cmd/fillstruct@latest"
    "godef|github.com/rogpeppe/godef@latest"
    "goimports|golang.org/x/tools/cmd/goimports@latest"
    "golangci-lint|github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    "gomodifytags|github.com/fatih/gomodifytags@latest"
    "gopls|golang.org/x/tools/gopls@latest"
    "gotags|github.com/jstemmer/gotags@latest"
    "iferr|github.com/koron/iferr@latest"
    "impl|github.com/josharian/impl@latest"
    "motion|github.com/fatih/motion@latest"
    "revive|github.com/mgechev/revive@latest"
    "staticcheck|honnef.co/go/tools/cmd/staticcheck@latest"
  )
  for entry in "${go_tools[@]}"; do
    name="${entry%%|*}"
    module="${entry#*|}"
    if [[ "$updating" == "1" || ! -x "$go_bin/$name" ]]; then
      run_step "Go tool $name" go install "$module"
    fi
  done
fi

tpm_dir="$HOME/.tmux/plugins/tpm"
if [[ ! -d "$tpm_dir/.git" ]] && command -v git >/dev/null 2>&1; then
  run_step "tmux plugin manager" git clone --depth 1 https://github.com/tmux-plugins/tpm "$tpm_dir"
elif [[ "$updating" == "1" && -d "$tpm_dir/.git" ]]; then
  run_step "update tmux plugin manager" git -C "$tpm_dir" pull --ff-only
fi
if [[ -x "$tpm_dir/bin/install_plugins" ]]; then
  run_step "tmux plugins" "$tpm_dir/bin/install_plugins"
fi

if command -v nvim >/dev/null 2>&1; then
  run_step "Neovim plugins" nvim --headless "+Lazy! restore" +qa
  if [[ "$updating" == "1" ]]; then
    run_step "Neovim Mason tools" nvim --headless "+MasonUpdate" "+MasonToolsUpdateSync" +qa
  else
    run_step "Neovim Mason tools" nvim --headless "+MasonToolsInstallSync" +qa
  fi
fi

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
if command -v claude >/dev/null 2>&1 && [[ -f "$repo_root/dotfiles/claude/setup-plugins.sh" ]]; then
  run_step "Claude Code plugins" bash "$repo_root/dotfiles/claude/setup-plugins.sh"
fi

if command -v bun >/dev/null 2>&1 && [[ -f "$HOME/.config/opencode/package.json" ]]; then
  if [[ "$updating" == "1" ]]; then
    run_step "OpenCode plugin dependencies" bun update --cwd "$HOME/.config/opencode"
  else
    run_step "OpenCode plugin dependencies" bun install --cwd "$HOME/.config/opencode"
  fi
fi

if ((failures > 0)); then
  echo "$failures post-install step(s) failed; run licokit doctor for details."
  exit 1
fi

echo "Development post-install setup complete."
