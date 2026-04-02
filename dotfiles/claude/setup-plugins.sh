#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SETTINGS_FILE="$SCRIPT_DIR/settings.json"

if ! command -v claude &>/dev/null; then
  echo "claude CLI not found, skipping plugin setup."
  exit 0
fi

if ! command -v jq &>/dev/null; then
  echo "Error: jq not found. Install with: brew install jq"
  exit 1
fi

if [[ ! -f "$SETTINGS_FILE" ]]; then
  echo "settings.json not found at $SETTINGS_FILE, skipping plugin setup."
  exit 0
fi

echo "=== Claude Code Plugin Setup ==="

# Add extra marketplaces
printf "\n--- Adding marketplaces ---\n"
jq -r '.extraKnownMarketplaces // {} | to_entries[] | .value.source.repo' "$SETTINGS_FILE" | while read -r repo; do
  echo "Adding marketplace: $repo"
  claude plugin marketplace add "$repo" 2>/dev/null || echo "  (already added or failed)"
done

# Install enabled plugins
printf "\n--- Installing plugins ---\n"
jq -r '.enabledPlugins // {} | to_entries[] | select(.value == true) | .key' "$SETTINGS_FILE" | while read -r plugin; do
  plugin_name="${plugin%%@*}"
  echo "Installing: $plugin_name"
  claude plugin install "$plugin" 2>/dev/null || echo "  (already installed or failed)"
done

# Clean up ECC artifacts — keep only common rules and no user skills (ECC plugin provides them)
printf "\n--- Cleaning ECC artifacts ---\n"
CLAUDE_DIR="$HOME/.claude"

# Remove language-specific rules (ECC installs all languages; we only want common)
for dir in "$CLAUDE_DIR"/rules/*/; do
  dirname="$(basename "$dir")"
  if [[ "$dirname" != "common" ]]; then
    echo "Removing rules/$dirname (ECC artifact)"
    rm -r "$dir"
  fi
done

# Rename README.md in rules so it's not loaded as a rule
if [[ -f "$CLAUDE_DIR/rules/README.md" ]]; then
  mv "$CLAUDE_DIR/rules/README.md" "$CLAUDE_DIR/rules/README"
fi

# Remove user skills (ECC plugin provides them — no need for duplicates)
for dir in "$CLAUDE_DIR"/skills/*/; do
  dirname="$(basename "$dir")"
  if [[ "$dirname" != "learned" ]]; then
    rm -r "$dir"
  fi
done

# Remove stale symlinks to dotfiles (ECC plugin provides agents/commands/skills)
for link in "$CLAUDE_DIR/agents" "$CLAUDE_DIR/commands" "$CLAUDE_DIR/skills"; do
  if [[ -L "$link" ]]; then
    echo "Removing stale symlink: $link"
    rm "$link"
  fi
done

printf "\n=== Done ===\n"
