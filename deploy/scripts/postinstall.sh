#!/usr/bin/env bash
set -euo pipefail

# kt is a CLI tool, so there is no service to enable or restart.
# Keep this script intentionally small and idempotent.
if [ -x /usr/local/bin/kt ]; then
  echo "kt installed to /usr/local/bin/kt"
elif [ -x /usr/bin/kt ]; then
  echo "kt installed to /usr/bin/kt"
else
  echo "kt installed"
fi
