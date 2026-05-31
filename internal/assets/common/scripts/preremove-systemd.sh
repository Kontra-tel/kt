#!/usr/bin/env bash
set -euo pipefail

# Package managers invoke removal hooks during upgrades as well as removals.
# Leave service stop, disable, and restart policy to the deployment process.
true
