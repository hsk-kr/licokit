#!/usr/bin/env bash
set -euo pipefail

SESSION="${DEV_TMUX_SESSION:-dev}"

# Session already exists: just jump to it.
if tmux has-session -t "$SESSION" 2>/dev/null; then
  if [ -n "${TMUX:-}" ]; then
    exec tmux switch-client -t "$SESSION"
  else
    exec tmux attach-session -t "$SESSION"
  fi
fi

# Layout per dev window:
#   +----------------+----------------+
#   |      top       |                |
#   |     (left)     |     right      |
#   +-------+--------+  (full height) |
#   |  bl   |   br   |                |
#   +-------+--------+----------------+
add_dev_window() {
  local name="$1" dir="$2"
  local left bl
  left=$(tmux new-window -t "$SESSION" -n "$name" -c "$dir" -P -F '#{pane_id}')
  tmux split-window -h -t "$left" -c "$dir"
  bl=$(tmux split-window -v -l 30% -t "$left" -c "$dir" -P -F '#{pane_id}')
  tmux split-window -h -t "$bl" -c "$dir"
  tmux select-pane -t "$left"
}

tmux new-session -d -s "$SESSION" -n _tmp -c "$HOME"

add_dev_window helloscreen "$HOME/dev/upscope/helloscreen"
add_dev_window userview "$HOME/dev/upscope/userview"
add_dev_window livedocument "$HOME/dev/upscope/livedocument"
tmux new-window -t "$SESSION" -n home -c "$HOME"
add_dev_window licokit "$HOME/licokit"

tmux kill-window -t "$SESSION:_tmp"
tmux select-window -t "$SESSION:helloscreen"

if [ -n "${TMUX:-}" ]; then
  exec tmux switch-client -t "$SESSION"
else
  exec tmux attach-session -t "$SESSION"
fi
