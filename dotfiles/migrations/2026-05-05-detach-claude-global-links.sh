#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if ! CLAUDE_DOTFILES="$(cd "$SCRIPT_DIR/../claude" && pwd)"; then
  echo "No licokit Claude dotfiles directory found."
  exit 0
fi
CLAUDE_DIR="$HOME/.claude"
BACKUP_DIR=""
DETACHED=0

ensure_backup_dir() {
  if [[ -z "$BACKUP_DIR" ]]; then
    BACKUP_DIR="$CLAUDE_DIR/backups/licokit-detach-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"
  fi
}

backup_link_target() {
  local path="$1"
  local relative="${path#$CLAUDE_DIR/}"
  local destination="$BACKUP_DIR/$relative"

  mkdir -p "$(dirname "$destination")"
  cp -pR -L "$path" "$destination" 2>/dev/null || true
}

detach_if_licokit_link() {
  local path="$1"
  local target

  if [[ ! -L "$path" ]]; then
    return 0
  fi

  target="$(readlink "$path")"
  case "$target" in
    "$CLAUDE_DOTFILES"/*)
      ensure_backup_dir
      backup_link_target "$path"
      unlink "$path"
      DETACHED=$((DETACHED + 1))
      ;;
  esac
}

if [[ ! -d "$CLAUDE_DIR" ]]; then
  echo "No Claude directory found at $CLAUDE_DIR."
  exit 0
fi

for path in \
  "$CLAUDE_DIR/settings.json" \
  "$CLAUDE_DIR/CLAUDE.md" \
  "$CLAUDE_DIR/docs" \
  "$CLAUDE_DIR/statusline-command.sh" \
  "$CLAUDE_DIR/hooks" \
  "$CLAUDE_DIR/hooks/hooks" \
  "$CLAUDE_DIR/agents" \
  "$CLAUDE_DIR/commands" \
  "$CLAUDE_DIR/skills"
do
  detach_if_licokit_link "$path"
done

if [[ "$DETACHED" -gt 0 ]]; then
  echo "Detached $DETACHED licokit-managed Claude link(s). Backups: $BACKUP_DIR"
else
  echo "No licokit-managed Claude links found."
fi
