#!/usr/bin/env bash

set -euo pipefail

if ! command -v claude &>/dev/null; then
  echo "claude CLI not found, skipping plugin update."
  exit 0
fi

echo "=== Claude Code Plugin Update ==="

plugins=$(claude plugin list 2>/dev/null | awk '/^[[:space:]]*❯ / { print $2 }')

if [[ -z "$plugins" ]]; then
  echo "No installed plugins found."
  exit 0
fi

while IFS= read -r plugin; do
  [[ -z "$plugin" ]] && continue
  echo ""
  echo "--- Updating: $plugin ---"
  claude plugin update "$plugin" || echo "  (update failed, skipping)"
done <<< "$plugins"

echo ""
echo "=== Done ==="
echo "Restart Claude Code to apply updates."
