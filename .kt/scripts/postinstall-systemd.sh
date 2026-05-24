#!/usr/bin/env bash
set -euo pipefail
systemctl daemon-reload
for service in "$@"; do
  systemctl enable "$service"
  systemctl restart "$service"
done
