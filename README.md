<p align="center">
  <h1 align="center">LicoKit</h1>
  <p align="center">A reproducible Apple Silicon macOS development setup.</p>
</p>

LicoKit installs and updates the software, runtimes, global developer tools, editor plugins, dotfiles, and background services used on Lico's laptop. It is designed to rebuild a freshly erased Mac without reusing the source checkout as the installed binary location.

## Reset workflow

On the old Mac, update LicoKit and check reset safety:

```bash
cd ~/licokit
git pull --ff-only
go run . doctor --profile personal --reset
```

`doctor --reset` is read-only. It scans Git repositories below `~/dev` and `~/hobby`, plus the legacy `~/licokit` checkout, and fails if a branch has dirty files, no upstream, or commits absent from locally known remote refs. It never backs up, edits, commits, pushes, or deletes projects. Push any work you need to keep and rerun the check before erasing the Mac.

On the fresh Mac, use Safari to run:

```bash
curl -fsSL https://raw.githubusercontent.com/hsk-kr/licokit/main/install.sh | bash -s -- install --profile personal
```

The installer downloads the latest release assets from GitHub, verifies the SHA-256 checksum, and installs `licokit` to `~/.local/bin/licokit`. The first run may pause while macOS installs Command Line Tools; finish that system dialog and rerun the same command.

Afterward, verify the result:

```bash
licokit doctor --profile personal
```

## Commands

```text
licokit                                      interactive menu
licokit install --profile personal           install/repair the laptop setup
licokit install --profile personal --dry-run show the planned setup
licokit update --profile personal            update packages, tools, and dotfiles
licokit update --profile personal --dry-run  show the update plan
licokit doctor --profile personal            inspect setup coverage
licokit doctor --profile personal --reset    block on local-only project work
licokit dotfiles                             safely refresh dotfile links
licokit version
```

Profiles:

- `core`: development environment and core applications.
- `personal`: `core` plus the personal applications found on the audited laptop. This is the normal rebuild profile.
- `all`: every item declared in the config, including future optional groups.

## What is covered

The default config currently declares 58 installable items.

| Area | Coverage |
|---|---|
| macOS bootstrap | Command Line Tools, Rosetta 2, Homebrew |
| Homebrew taps | anomalyco, AeroSpace, Supabase, y3owk1n |
| Core CLI tools | Git, GitHub CLI, Neovim, tree-sitter, tmux, ripgrep, fzf, tree, jq, zsh-vi-mode, bison, Mercurial, p7zip, btop, terminal-notifier |
| Languages/media/data | Go, nvm, Ruby, Python 3.13, Zig, ffmpeg, ImageMagick, yt-dlp, Supabase CLI |
| Developer apps | Docker Desktop, Ghostty, AeroSpace, Homerow, Karabiner Elements, Snipaste, Chrome, Slack, ChatGPT |
| AI/dev CLIs | Claude Code, Codex CLI, OpenCode, Bun, uv |
| Personal apps | WhatsApp, CapCut, NordVPN, OBS, pgAdmin 4, Folx |
| Mac App Store | Xcode, KakaoTalk, MKPlayer, Quick Camera |
| Node | Node 24 and 26, default Node 24, npm, Corepack, pnpm 10.19, Yarn 1.22, Wrangler |
| Go globals | air, asmfmt, delve, errcheck, fillstruct, godef, goimports, golangci-lint, gomodifytags, gopls, gotags, iferr, impl, motion, revive, staticcheck |
| Editor/shell | lazy.nvim plugins, Mason tools, tmux plugin manager/plugins, Claude plugins, OpenCode plugin dependencies, managed zsh paths |
| Service | High-CPU watchdog LaunchAgent |

Mason manages the LSP/formatter/debugger set, including Bash, Biome, ESLint, Go, Harper, JSON, Lua, Markdown, Prettier, Python, Ruby, Tailwind, TypeScript, YAML, shellcheck, shfmt, stylua, and delve tooling.

No project backup or project restore logic is included. GitHub remains the source of truth for repositories.

## What macOS still requires

Apple does not permit a bootstrap tool to silently finish every system-owned step. Expect to confirm or restore:

- Apple/App Store, GitHub, SSH, Docker, VPN, ChatGPT, Slack, and other app sign-ins.
- Accessibility/Input Monitoring for Homerow, AeroSpace, and Karabiner Elements.
- Screen Recording, microphone/camera, notifications, Full Disk Access, VPN extensions, and other Privacy & Security permissions as used.
- Any non-repository secrets, SSH private keys, certificates, browser state, databases, or app-local data you want to retain.

LicoKit reports these as reminders rather than claiming they were automated.

## Safe dotfiles

The executable lives at `~/.local/bin/licokit`; its managed repository lives separately at `~/.local/share/licokit/repo`. This prevents the previous fresh-install collision where the executable occupied the intended clone directory.

Dotfiles are linked from the managed checkout. If a regular target already exists, LicoKit moves it under `~/.local/state/licokit/backups/<timestamp>/` before creating the symlink. It will not hard-reset a dirty LicoKit checkout; updates use `git pull --ff-only`.

Managed config links are AeroSpace, Karabiner Elements, Neovim, OpenCode, tmux, zsh, and Ghostty, plus `~/scripts`.

## Configuration

Defaults are embedded from `lib/config/default_config.yaml`. To override them, create `~/.config/licokit/config.yaml`.

```yaml
dotfiles:
  repo: "https://github.com/hsk-kr/licokit.git"
  repo_path: "~/.local/share/licokit/repo"
  config_links: [nvim, tmux, zsh]
  zsh_source: "~/.config/zsh/zshrc"

services:
  cpu_killer: true

tools:
  - name: Go
    install_type: brew       # brew | cask | tap | mas | script | manual
    package: go
    detect_type: command     # command | application | brew_package | brew_tap | package_receipt | xcode
    detect_value: go
    profiles: [core]
```

## Development

Requires Go 1.23 or newer:

```bash
git clone https://github.com/hsk-kr/licokit.git
cd licokit
go test ./...
go run . install --profile personal --dry-run
GOOS=darwin GOARCH=arm64 go build -o licokit-darwin-arm64 .
```

Release tags build `licokit-darwin-arm64` and `licokit-darwin-arm64.sha256`. The public installer refuses to install if checksum verification fails.

## CPU killer

The optional LaunchAgent checks only the current user's processes every 30 seconds. A process over 90% CPU for 20 consecutive checks is terminated and a notification is sent. Its script is linked to `~/scripts/cpu-killer.sh`; logs are in `/tmp/cpu-killer.log`, and overrides live in `~/.config/cpu-killer/config`.

## License

[MIT](LICENSE)
