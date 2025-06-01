# Reference Stats Data Engineering Agent Brief

> **Location:** This brief and all GWAS data engineering is located in the shared, normalized GWAS database directory (`gwas/`), outside of any specific application module.

> **Integration with Rebuildable Database:**
> - All reference stats and panel tables (e.g., `reference_panel`) **must be defined in schema SQL files** (e.g., `sql/create_table_reference_panel.sql`).
> - The main DB build script (`build_db.sh`) must `.read` these schema files so that all required tables are created on every DB rebuild.
> - Data ingestion scripts (for panel/sample/reference data) should be run **after** the database and tables are created.
> - This ensures the database can be fully rebuilt from schema and raw data at any time, supporting reproducibility and local development.

---


**Project:** GWAS Database (gwas.duckdb) — Reference Panel and Stats Ingestion

---

## Objective
Provide a reproducible, script-driven workflow to download the HGDP + 1000 Genomes (gnomAD v3.1.2, GRCh38) reference panel and VCF data, and insert ancestry/sample metadata into the DuckDB database. This enables downstream calculation of PRS reference statistics for polygenic risk normalization using the GRCh38 genome build.

All scripts/tools produced from this brief must:
- Be runnable on a new computer with only base developer tools (bash, wget, python, etc.)
- Require no manual intervention (fully automated, idempotent)
- Be referenced and invoked by `build_db.sh` as the single entry point for database and data setup


---

## Responsibilities
- Download the HGDP + 1000 Genomes sample panel metadata (gnomAD v3.1.2, GRCh38).
- Download VCF files for all chromosomes from gnomAD v3.1.2 (GRCh38).
- Extract sample IDs and ancestry for the desired population (e.g., EUR) from the sample metadata.
- Insert panel/sample metadata into the DuckDB database (`reference_panel` or similar table).
- Ensure idempotency: scripts should not re-download or re-insert existing data.
- Provide clear logging and error handling.
- Ensure all new scripts/data ingestion steps are referenced in `build_db.sh` so that a single command sets up the complete database and data from scratch.


---

## Required Bash Script Steps

> **Note:** All scripts must be designed to run as part of `build_db.sh` with no manual steps. Any script or tool referenced here must be included/invoked from `build_db.sh`.
1. **Download Sample Panel Metadata (GRCh38):**
   - Use `wget -N` or `curl -O` to fetch the sample metadata file from:
     - `https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/sample_qc/gnomad.genomes.v3.1.2.sites.samples.meta.tsv.bgz`
   - Decompress with `bgzip -d` if necessary.
2. **Extract Ancestry Samples:**
   - Use `awk` or similar to filter for desired ancestry (e.g., EUR) from the metadata TSV.
   - Output: `eur_samples.tsv` (tab-separated: sample_id, ancestry, sex).
3. **Download VCF(s) (GRCh38):**
   - Use `wget -c` to download all chromosome VCFs from:
     - `https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/vcf/genomes/`
   - Download both `.vcf.bgz` and `.vcf.bgz.tbi` files for chromosomes 1–22.
4. **Insert Panel Data into DuckDB:**
   - Use the `duckdb` CLI or SQL script to create and populate the `reference_panel` table.
   - Table schema example:
     | Column     | Type   | Description                |
     |------------|--------|----------------------------|
     | sample_id  | TEXT   | gnomAD sample ID           |
     | ancestry   | TEXT   | Ancestry group (e.g., EUR) |
     | sex        | TEXT   | Reported sex               |
   - Example SQL:
     ```sql
     CREATE TABLE IF NOT EXISTS reference_panel (
         sample_id TEXT,
         ancestry TEXT,
         sex TEXT
     );
     COPY reference_panel FROM 'eur_samples.tsv' (DELIMITER '\t', HEADER FALSE);
     ```
5. **Log completion and any errors.**

---

## Inputs
- gnomAD v3.1.2 (GRCh38) sample panel metadata URL (default: `https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/sample_qc/gnomad.genomes.v3.1.2.sites.samples.meta.tsv.bgz`)
- VCF base URL (default: `https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/vcf/genomes/`)
- Target ancestry (e.g., EUR)
- DuckDB database path

## Outputs
- Downloaded sample panel metadata and VCF files (GRCh38)
- Populated `reference_panel` table in DuckDB
- Log files or console output

## Required Tests
- Panel and sample data downloaded and inserted only if missing
- Table schema matches requirements and is always present after DB rebuild
- Errors are logged and script exits with nonzero status on failure
- Database can be rebuilt from scratch using `build_db.sh` followed by data ingestion scripts

## Reference Panel Data Preparation Script

To support the polygenic-risk-score pipeline, a new script (e.g., `scripts/reference_script.sh`) must be created with the following minimal responsibilities:

### 1. Download the gnomAD v3.1.2 (GRCh38) Sample Panel Metadata
- Download the metadata file from:
  - `https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/sample_qc/gnomad.genomes.v3.1.2.sites.samples.meta.tsv.bgz`
- Save to a local data directory (e.g., `data/`).
- Decompress with `bgzip -d` if needed.

### 2. Filter for Target Ancestry
- Extract sample IDs, ancestry, and sex for the desired ancestry (e.g., EUR) from the metadata file.
- Output a TSV file with columns: sample_id, ancestry, sex (e.g., `eur_samples.tsv`).

### 3. Download VCFs (GRCh38)
- Download all chromosome VCFs and index files from:
  - `https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/vcf/genomes/`
- Save to a local data directory (e.g., `gnomad_grch38_vcf/`).

### 4. Insert Sample Metadata into DuckDB
- Insert the filtered sample IDs, ancestry, and sex into the `reference_panel` table in the DuckDB database.
- Use the schema:
  - `sample_id TEXT, ancestry TEXT, sex TEXT`
- Use the DuckDB CLI or SQL `COPY` command for bulk insert.

### 5. Log Actions and Errors
- All steps must log completion and any errors to stdout or a log file.

---

**No additional or extraneous steps are to be included. This script is solely for preparing the reference panel/sample metadata required by the polygenic-risk-score pipeline.**

---

## Data Schema Components

### 1. Reference Panel Table (`reference_panel`)
- **Purpose:** Stores the full, unfiltered sample metadata for all reference samples. Enables flexible ancestry/subset selection at query time.
- **Schema (example, all columns from gnomAD v3.1.2 sample panel):**
  - `sample_id TEXT`
  - `sex TEXT`
  - `pop TEXT` (population)
  - `super_pop TEXT` (super-population/ancestry)
  - `platform TEXT`
  - ... (all other columns present in the raw metadata TSV)
- **Source:** Full gnomAD v3.1.2 sample panel metadata (see [reference_stats_brief.md](./reference_stats_brief.md))
- **Note:** Filtering for ancestry or other subsets should be performed via SQL query, not at ingestion.

### 2. Reference VCFs
- **Purpose:** Provide variant site data for all reference samples, enabling downstream PRS and allele frequency calculations
- **Format:** bgzipped, tabix-indexed VCFs for chromosomes 1–22
- **Source:** gnomAD v3.1.2 (see [vcf_download_hgdp_gnomad_grch38.md](./vcf_download_hgdp_gnomad_grch38.md))
- **Note:** VCFs are stored on disk and not loaded into DuckDB; they are used for downstream analysis and stats calculation.

### 3. PRS Reference Statistics Table (`reference_stats`)
- **Purpose:** Stores per-SNP reference statistics (e.g., allele frequencies, mean PRS, SD PRS) stratified by ancestry, required for PRS normalization in downstream tools such as the Polygenic Risk Calculator.
- **Schema (example):**
  - `rsid TEXT`
  - `ancestry TEXT`
  - `allele_freq DOUBLE`
  - `mean_prs DOUBLE`
  - `sd_prs DOUBLE`
- **Integration:**
  - This table must be created in SQL and populated as part of the automated pipeline (invoked by `build_db.sh`).
  - The schema should be aligned with the canonical data model (see `.agent/data_model.md`).
  - This table is required for downstream normalization; tools will expect it to exist and be populated for relevant SNPs and ancestries.

---

## Example Workflow
1. **Rebuild the database schema and ingest reference data:**
   ```bash
   bash build_db.sh
   ```
2. **(Optional) Inspect tables:**
   ```bash
   duckdb gwas.duckdb ".schema"
   ```

---

## Reproducibility Checklist
- [ ] All schema (including reference stats tables) defined in SQL files and referenced by `build_db.sh`
- [ ] Database can be rebuilt from schema and raw data alone
- [ ] Data download and ingestion scripts are run automatically by `build_db.sh`
- [ ] Input files and script versions are tracked for provenance

## Notes
- This brief covers only panel/sample ingestion. PRS calculation and stats insertion are handled elsewhere.
- For full reproducibility, version input files and document script versions.
