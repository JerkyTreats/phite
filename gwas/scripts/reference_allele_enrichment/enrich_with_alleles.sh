#!/usr/bin/env bash

# Ensures required Python packages are installed, then runs the enrichment script.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REQS_FILE="${SCRIPT_DIR}/requirements.txt"
PY_SCRIPT="${SCRIPT_DIR}/enrich_with_alleles.py"

# Upgrade pip and install requirements quietly
python -m pip install --quiet --upgrade pip
python -m pip install --quiet -r "${REQS_FILE}"

# Forward all arguments to the Python script
python "${PY_SCRIPT}" "$@"
