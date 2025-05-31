#!/bin/bash
# reference_script.sh: Prepare reference panel/sample metadata for polygenic-risk-score pipeline
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DATA_DIR="$PROJECT_ROOT/data"
PANEL_URL="https://ftp.1000genomes.ebi.ac.uk/vol1/ftp/release/20130502/integrated_call_samples_v3.20130502.ALL.panel"
DB_PATH="$PROJECT_ROOT/gwas.duckdb"
ANCESTRY="EUR"
PANEL_FILE="$DATA_DIR/integrated_call_samples_v3.20130502.ALL.panel"
SAMPLES_FILE="$DATA_DIR/$(echo "$ANCESTRY" | tr '[:upper:]' '[:lower:]')_samples.txt"

mkdir -p "$DATA_DIR"

# 1. Download the 1000 Genomes Sample Panel
if [ ! -f "$PANEL_FILE" ]; then
  echo "Downloading 1000G panel: $PANEL_URL"
  curl -L -o "$PANEL_FILE" "$PANEL_URL"
else
  echo "Panel file already exists: $PANEL_FILE"
fi

# 2. Filter for target ancestry and extract sample_id, ancestry, sex
# Always regenerate the ancestry file and ensure only valid lines are written
echo "Filtering panel for ancestry: $ANCESTRY"
# Write header, then append filtered rows
{
  echo -e "sample_id\tancestry\tsex"
  awk -v anc="$ANCESTRY" 'NR > 1 && $3 == anc {print $1 "\t" $3 "\t" $4}' "$PANEL_FILE"
} > "$SAMPLES_FILE"

# 4. Insert sample metadata into DuckDB
if [ -s "$SAMPLES_FILE" ]; then
  echo "Preview of $SAMPLES_FILE just before import:"
  head "$SAMPLES_FILE"
  wc -l "$SAMPLES_FILE"
  awk -F'\t' 'NF!=3 {print NR, $0}' "$SAMPLES_FILE"
  echo "Inserting sample metadata into DuckDB: $DB_PATH"
  duckdb "$DB_PATH" \
    "CREATE TABLE IF NOT EXISTS reference_panel (sample_id TEXT, ancestry TEXT, sex TEXT);"
  duckdb "$DB_PATH" \
    "DELETE FROM reference_panel WHERE ancestry = '$ANCESTRY';"
  duckdb "$DB_PATH" <<EOF
.mode tabs
.import $SAMPLES_FILE reference_panel
EOF
  echo "Sample metadata inserted for ancestry: $ANCESTRY"
else
  echo "No sample metadata to insert for ancestry: $ANCESTRY" >&2
  exit 1
fi

echo "Reference panel data preparation complete."
