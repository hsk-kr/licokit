#!/usr/bin/env bash
#
# cpu-killer.sh — kill any of your processes that sustains high CPU for too long.
#
# Default rule: if a process uses more than CPU_THRESHOLD% CPU for STRIKES
# consecutive samples taken every INTERVAL seconds, kill it.
#   90% x 30s x 20 strikes = 10 minutes of sustained high CPU.
#
# NOTE: macOS reports CPU as a percentage of a SINGLE core, so values above
# 100% are possible on multi-core machines. 90% ~= most of one core.
#
# Only the current user's processes are considered, so system/root daemons
# (kernel_task, WindowServer, mds, ...) are never touched.
#
# Written for /bin/bash 3.2 (the macOS system bash): no associative arrays,
# per-PID state is kept in files under STATE_DIR.

# Keep a predictable PATH even under launchd's minimal environment.
export PATH="/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin:/usr/local/bin:$PATH"

# ---- Configuration (override in ~/.config/cpu-killer/config) ----------------
CPU_THRESHOLD=90      # percent (per core) that counts as "high"
INTERVAL=30           # seconds between samples
STRIKES=20            # consecutive high samples before a kill (20 x 30s = 10 min)
KILL_GRACE=5          # seconds to wait after SIGTERM before SIGKILL
NOTIFY=true           # post a macOS notification when something is killed
LOG_FILE="/tmp/cpu-killer.log"
# Process names (basename of the command) that are never killed.
EXCLUDE="Finder loginwindow"

CONFIG_FILE="${CPU_KILLER_CONFIG:-$HOME/.config/cpu-killer/config}"
[ -f "$CONFIG_FILE" ] && . "$CONFIG_FILE"

STATE_DIR="/tmp/cpu-killer-state"
mkdir -p "$STATE_DIR"

SELF_PID=$$
# Resolve the user from the process credentials, NOT $USER — launchd starts
# this agent with an empty environment, so $USER/$HOME are unset there.
MYUSER="$(id -un)"

log() {
  printf '%s %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*" >> "$LOG_FILE"
}

# Post a desktop notification, preferring terminal-notifier, falling back to
# osascript (always present). Never fails the watchdog if notifications break.
notify() {
  title="$1"
  message="$2"
  tn="$(command -v terminal-notifier 2>/dev/null)"
  if [ -z "$tn" ]; then
    for cand in /opt/homebrew/bin/terminal-notifier /usr/local/bin/terminal-notifier; do
      [ -x "$cand" ] && { tn="$cand"; break; }
    done
  fi
  if [ -n "$tn" ]; then
    "$tn" -title "$title" -message "$message" >/dev/null 2>&1 && return
  fi
  /usr/bin/osascript -e "display notification \"$message\" with title \"$title\"" >/dev/null 2>&1
}

is_excluded() {
  name="$1"
  case " $EXCLUDE " in
    *" $name "*) return 0 ;;
  esac
  return 1
}

# Remove state files for PIDs that no longer exist.
prune_state() {
  for f in "$STATE_DIR"/*; do
    [ -e "$f" ] || continue
    pid="${f##*/}"
    kill -0 "$pid" 2>/dev/null || rm -f "$f"
  done
}

# Escalating kill: SIGTERM, wait KILL_GRACE seconds, then SIGKILL if needed.
kill_process() {
  pid="$1"; name="$2"; cpu="$3"; count="$4"
  log "KILL pid=$pid name=$name cpu=${cpu}% sustained $((count * INTERVAL))s >${CPU_THRESHOLD}%"
  kill -TERM "$pid" 2>/dev/null
  i=0
  while [ "$i" -lt "$KILL_GRACE" ]; do
    kill -0 "$pid" 2>/dev/null || break
    sleep 1
    i=$((i + 1))
  done
  if kill -0 "$pid" 2>/dev/null; then
    kill -KILL "$pid" 2>/dev/null
    log "force-killed pid=$pid name=$name (ignored SIGTERM)"
  else
    log "terminated pid=$pid name=$name"
  fi
  if [ "$NOTIFY" = "true" ]; then
    notify "CPU killer" "Killed $name (pid $pid) — ${cpu}% CPU for ~$((count * INTERVAL / 60)) min"
  fi
}

sample() {
  prune_state

  # Current user's processes, highest CPU first.
  ps -U "$MYUSER" -o pid=,pcpu=,comm= -r 2>/dev/null | while read -r pid cpu comm; do
    [ -n "$pid" ] || continue
    [ "$pid" = "$SELF_PID" ] && continue

    # Strip the directory with parameter expansion (handles names like "-zsh"
    # that would make basename treat the leading dash as an option).
    name="${comm##*/}"
    is_excluded "$name" && { rm -f "$STATE_DIR/$pid"; continue; }

    cpu_int="${cpu%%.*}"
    case "$cpu_int" in
      ''|*[!0-9]*) continue ;;
    esac

    state_file="$STATE_DIR/$pid"

    if [ "$cpu_int" -gt "$CPU_THRESHOLD" ]; then
      count=0
      if [ -f "$state_file" ]; then
        # Count FIRST so the command name (which may contain spaces, e.g.
        # "Google Chrome Helper (Renderer)") is the remainder read keeps intact.
        read -r saved_count saved_name < "$state_file" 2>/dev/null
        case "$saved_count" in ''|*[!0-9]*) saved_count=0 ;; esac
        # Same PID, same command? carry the strike count. Otherwise PID was
        # reused by a different process — start over.
        [ "$saved_name" = "$name" ] && count="$saved_count"
      fi
      count=$((count + 1))
      printf '%s %s\n' "$count" "$name" > "$state_file"

      if [ "$count" -ge "$STRIKES" ]; then
        kill_process "$pid" "$name" "$cpu" "$count"
        rm -f "$state_file"
      else
        log "high pid=$pid name=$name cpu=${cpu}% strike=$count/$STRIKES"
      fi
    else
      # Dropped below the threshold — reset its streak.
      rm -f "$state_file"
    fi
  done
}

log "cpu-killer started: threshold=${CPU_THRESHOLD}% interval=${INTERVAL}s strikes=${STRIKES} (kill after $((STRIKES * INTERVAL))s) notify=${NOTIFY}"

while true; do
  # Re-read the config each loop so edits apply within one interval.
  [ -f "$CONFIG_FILE" ] && . "$CONFIG_FILE"
  sample
  sleep "$INTERVAL"
done
