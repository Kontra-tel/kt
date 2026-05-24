#!/usr/bin/env bash
set -euo pipefail
for service in "$@"; do
  systemctl stop "$service" || true
  systemctl disable "$service" || true
done
systemctl daemon-reload
