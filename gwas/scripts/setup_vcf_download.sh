#!/bin/bash
set -e

set -e

# Ensure gsutil is installed
if ! command -v gsutil &> /dev/null; then
  echo "gsutil not found. Installing via pip..."
  pip install --user gsutil
  export PATH="$HOME/.local/bin:$PATH"
fi

VCF_DIR="$(dirname "$0")/../gnomad_grch38_vcf"
mkdir -p "$VCF_DIR"

# Download all chr1-22 VCFs and .tbi files
for CHR in {1..22}; do
  for EXT in vcf.bgz vcf.bgz.tbi; do
    FILE="gnomad.genomes.v3.1.2.sites.chr${CHR}.${EXT}"
    if [ ! -f "$VCF_DIR/$FILE" ]; then
      echo "Downloading $FILE ..."
      gsutil cp "gs://gcp-public-data--gnomad/release/3.1.2/vcf/genomes/$FILE" "$VCF_DIR/"
    else
      echo "$FILE already exists, skipping."
    fi
  done
done
