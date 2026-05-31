#!/usr/bin/env bash
set -euo pipefail
systemctl daemon-reload

# Enabling and restarting services is intentionally left to the deployment
# process so upgrades can run health checks and migrations at the right time.
