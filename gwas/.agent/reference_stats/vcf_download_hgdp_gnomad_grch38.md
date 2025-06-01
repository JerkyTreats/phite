
# 📦 Downloading HGDP + 1000 Genomes VCF Files (GRCh38)

This section defines the commands and structure for downloading the full set of GRCh38-aligned genotype data from the **HGDP + 1000 Genomes** reference panel, published via **gnomAD v3**. This dataset includes diverse global populations and is fully aligned to GRCh38. 

**Note:** For a complete reference panel pipeline, also download the corresponding sample panel metadata (see `reference_stats_brief.md`) and ingest sample/ancestry information into DuckDB as described there.

## 🌍 Source

The data is hosted on Google Cloud (public bucket):

```
gs://gcp-public-data--gnomad/release/3.1.2/ht/genomes/
```

Download via the [gsutil CLI](https://cloud.google.com/storage/docs/gsutil_install) or mirror via [Broad Institute Terra](https://gnomad.broadinstitute.org/downloads#v3-hgdp).

Ensure any CLI or setup steps are included in any produced script. See .agent/README for expanded setup/build requirements.

---

## 📁 Dataset Format

Files are in **Hail Table (.ht)** and **MatrixTable (.mt)** formats, partitioned by genome build.

For PRS usage, you can use the **VCF export subset**, specifically:

```
gs://gcp-public-data--gnomad/release/3.1.2/vcf/genomes/
```

Example files:

```
gnomad.genomes.v3.1.2.sites.chr1.vcf.bgz
gnomad.genomes.v3.1.2.sites.chr1.vcf.bgz.tbi
```

---

## 🖥️ Bash Script: Download All Chromosomes

```bash
#!/usr/bin/env bash
set -e

BASE_URL="https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/vcf/genomes"
DEST_DIR="./gnomad_grch38_vcf"
mkdir -p "$DEST_DIR"

for CHR in {1..22}; do
  for EXT in vcf.bgz vcf.bgz.tbi; do
    FILE="gnomad.genomes.v3.1.2.sites.chr${CHR}.${EXT}"
    echo "Downloading $FILE..."
    wget -c "$BASE_URL/$FILE" -P "$DEST_DIR"
  done
done
```

---

## 📎 Notes

- The `-c` flag resumes partial downloads if interrupted.
- Files are **bgzipped and tabix-indexed** for fast region-based access.
- These VCFs include **variant site data only**, not full genotypes.
- Use with additional `Hail` processing or subset VCFs if individual-level data is needed.

---

## 🧪 Validation: Check File Count

```bash
ls ./gnomad_grch38_vcf/*.vcf.bgz | wc -l
# Expect 22 VCFs + 22 .tbi index files
```

---

## Integrate with Build Script

Once complete, this will be a step added in the appropriate location in `build_db.sh`.

---

## Sample Panel Metadata (Required for Ancestry Ingestion)

To fully support ancestry-aware PRS and reference panel workflows, you must also download the gnomAD v3.1.2 (GRCh38) sample panel metadata:

```
https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/sample_qc/gnomad.genomes.v3.1.2.sites.samples.meta.tsv.bgz
```

After download and decompression, filter for your target ancestry and ingest into DuckDB as described in `reference_stats_brief.md`. This ensures your VCFs and sample metadata are always in sync and GRCh38-aligned.
