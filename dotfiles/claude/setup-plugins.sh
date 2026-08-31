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

BACKUP_ROOT="$HOME/.local/state/licokit/backups/claude-$(date +%Y%m%d-%H%M%S)"
BACKED_UP=0

preserve_path() {
  local source="$1"
  local relative="$2"
  local destination="$BACKUP_ROOT/$relative"
  mkdir -p "$(dirname "$destination")"
  mv "$source" "$destination"
  BACKED_UP=1
}

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

PLUGIN_FAILURES=0
plugin_list="$(claude plugin list 2>/dev/null || true)"
while IFS= read -r plugin; do
  [[ -z "$plugin" ]] && continue
  if ! grep -Fq "$plugin" <<< "$plugin_list"; then
    echo "Plugin verification failed: $plugin"
    PLUGIN_FAILURES=$((PLUGIN_FAILURES + 1))
  fi
done < <(jq -r '.enabledPlugins // {} | to_entries[] | select(.value == true) | .key' "$SETTINGS_FILE")

# Reconcile ECC artifacts. Anything displaced is preserved in LicoKit's
# recoverable config backup area rather than deleted.
printf "\n--- Cleaning ECC artifacts ---\n"
CLAUDE_DIR="$HOME/.claude"

# Preserve stale symlinks before traversing those paths. This avoids mutating a
# symlink target owned by another checkout or tool.
for link in "$CLAUDE_DIR/agents" "$CLAUDE_DIR/commands" "$CLAUDE_DIR/skills"; do
  if [[ -L "$link" ]]; then
    echo "Preserving stale symlink: $link"
    preserve_path "$link" "links/$(basename "$link")"
  fi
done

# Preserve language-specific rules (ECC installs all languages; we only want common)
if compgen -G "$CLAUDE_DIR/rules/*/" > /dev/null 2>&1; then
  for dir in "$CLAUDE_DIR"/rules/*/; do
    dirname="$(basename "$dir")"
    if [[ "$dirname" != "common" ]]; then
      echo "Preserving rules/$dirname (ECC artifact)"
      preserve_path "$dir" "rules/$dirname"
    fi
  done
fi

# Ensure common rules exist (copy from ECC plugin source if missing)
ECC_PLUGIN_DIR="$CLAUDE_DIR/plugins/marketplaces/everything-claude-code"
if [[ -d "$ECC_PLUGIN_DIR/rules/common" ]] && [[ ! -d "$CLAUDE_DIR/rules/common" ]]; then
  echo "Restoring rules/common from ECC plugin"
  mkdir -p "$CLAUDE_DIR/rules"
  cp -r "$ECC_PLUGIN_DIR/rules/common" "$CLAUDE_DIR/rules/common"
fi

# Rename README.md in rules so it's not loaded as a rule
if [[ -f "$CLAUDE_DIR/rules/README.md" ]]; then
  mv "$CLAUDE_DIR/rules/README.md" "$CLAUDE_DIR/rules/README"
fi

# Preserve user skills (ECC plugin provides them — no need for duplicates)
if compgen -G "$CLAUDE_DIR/skills/*/" > /dev/null 2>&1; then
  for dir in "$CLAUDE_DIR"/skills/*/; do
    dirname="$(basename "$dir")"
    if [[ "$dirname" != "learned" ]]; then
      preserve_path "$dir" "skills/$dirname"
    fi
  done
fi

if [[ "$BACKED_UP" == "1" ]]; then
  echo "Displaced Claude configuration was preserved at $BACKUP_ROOT"
fi

# Update all installed plugins to latest versions
if [[ -x "$SCRIPT_DIR/update-plugins.sh" ]]; then
  printf "\n"
  "$SCRIPT_DIR/update-plugins.sh"
fi

if [[ "$PLUGIN_FAILURES" -gt 0 ]]; then
  echo "$PLUGIN_FAILURES Claude plugin(s) are still missing."
  exit 1
fi

printf "\n=== Done ===\n"
