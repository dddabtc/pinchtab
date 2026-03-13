#!/bin/sh
set -eu

home_dir="${HOME:-/data}"
xdg_config_home="${XDG_CONFIG_HOME:-$home_dir/.config}"
default_config_path="$xdg_config_home/pinchtab/config.json"

mkdir -p "$home_dir" "$xdg_config_home" "$(dirname "$default_config_path")"

# Generate a persisted config on first boot.
# The PINCHTAB_TOKEN env var can be used to set an auth token via Docker secrets
# or environment variables. Prefer Docker secrets for sensitive data:
#   docker run -e PINCHTAB_TOKEN_FILE=/run/secrets/pinchtab_token
if [ -z "${PINCHTAB_CONFIG:-}" ] && [ ! -f "$default_config_path" ]; then
  /usr/local/bin/pinchtab config init >/dev/null
  if [ -n "${PINCHTAB_TOKEN:-}" ]; then
    /usr/local/bin/pinchtab config set server.token "$PINCHTAB_TOKEN" >/dev/null
  fi
fi

# RUNTIME BIND OVERRIDE FOR DOCKER PORT PUBLISHING
#
# The persisted config stores bind: "127.0.0.1" (secure loopback default).
# But Docker port publishing requires the process to listen on 0.0.0.0 inside
# the container, so the host can forward traffic to it.
#
# Solution: override PINCHTAB_BIND at runtime only when using managed config.
# The persisted config remains secure unless the user explicitly changes it.
if [ -z "${PINCHTAB_CONFIG:-}" ] && [ -z "${PINCHTAB_BIND:-}" ]; then
  export PINCHTAB_BIND=0.0.0.0
fi

# CHROME SANDBOX DISABLED IN CONTAINERS
#
# Chrome requires --no-sandbox inside containers because:
# - Containers don't have user namespaces (sandboxing requires this)
# - Container security (cgroups, capabilities, seccomp) provides isolation
# - The Dockerfile already drops capabilities and uses read-only filesystem
#
# This is standard for headless Chrome in containerized environments.
# Backfill the flag into managed config if not already set.
if [ -z "${PINCHTAB_CONFIG:-}" ] && [ -f "$default_config_path" ]; then
  current_flags="$(/usr/local/bin/pinchtab config get browser.extraFlags 2>/dev/null || true)"
  if [ -z "$current_flags" ]; then
    /usr/local/bin/pinchtab config set browser.extraFlags "--no-sandbox --disable-gpu" >/dev/null
  fi
fi

exec "$@"
