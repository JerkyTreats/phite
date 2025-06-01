#!/bin/bash
set -e

set -e

# Ensure gsutil is installed
if ! command -v gsutil &> /dev/null; then
  echo "gsutil not found. Installing via pip..."
  pip install --user gsutil
  export PATH="$HOME/.local/bin:$PATH"
fi

DATA_DIR="$(dirname "$0")/../data"
mkdir -p "$DATA_DIR"

# List candidate metadata files
echo "Listing candidate sample metadata files in GCS:"
gsutil ls gs://gcp-public-data--gnomad/release/3.1.2/vcf/genomes/ | grep -i 'meta\|sample\|tsv'

# Download the correct metadata file for v3.1.2
META_FILE="gnomad.genomes.v3.1.2.hgdp_1kg_subset_sample_meta.tsv.bgz"
if [ ! -f "$DATA_DIR/$META_FILE" ]; then
  echo "Downloading $META_FILE ..."
  gsutil cp "gs://gcp-public-data--gnomad/release/3.1.2/vcf/genomes/$META_FILE" "$DATA_DIR/"
else
  echo "$META_FILE already exists, skipping download."
fi

# Ensure bgzip is installed
if ! command -v bgzip &> /dev/null; then
  echo "bgzip not found. Installing htslib via Homebrew..."
  if command -v brew &> /dev/null; then
    brew install htslib
  else
    echo "Homebrew not found. Please install bgzip (htslib) manually. Aborting."
    exit 1
  fi
fi

# Decompress if needed
DECOMPRESSED_FILE="gnomad.genomes.v3.1.2.hgdp_1kg_subset_sample_meta.tsv"
if [ ! -f "$DATA_DIR/$DECOMPRESSED_FILE" ]; then
  echo "Decompressing $META_FILE ..."
  bgzip -d -c "$DATA_DIR/$META_FILE" > "$DATA_DIR/$DECOMPRESSED_FILE"
else
  echo "Decompressed TSV already exists, skipping."
fi

# (Optional) Ingest with Python as before
source .venv_reference_panel/bin/activate
pip install --upgrade pip
pip install -r "$(dirname "$0")/requirements.txt"
python "$(dirname "$0")/reference_panel_ingest.py"
