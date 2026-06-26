#!/usr/bin/env bash
set -euo pipefail

# Refresh service manager metadata when systemd is the active init system.
# Enabling and restarting services is intentionally left to the deployment
# process so upgrades can run health checks and migrations at the right time.
if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
  systemctl daemon-reload || true
fi
