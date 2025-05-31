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
Provide a reproducible, script-driven workflow to download 1000 Genomes reference panel data and insert ancestry/sample metadata into the DuckDB database. This enables downstream calculation of PRS reference statistics for polygenic risk normalization.

---

## Responsibilities
- Download the 1000 Genomes sample panel (`integrated_call_samples_v3.20130502.ALL.panel`).
- Download VCF files for selected chromosomes as needed.
- Extract sample IDs for the desired ancestry (e.g., EUR).
- Insert panel/sample metadata into the DuckDB database (`reference_panel` or similar table).
- Ensure idempotency: script should not re-download or re-insert existing data.
- Provide clear logging and error handling.

---

## Required Bash Script Steps
1. **Download Panel:**
   - Use `wget -N` to fetch the panel file.
2. **Extract Ancestry Samples:**
   - Use `awk` or similar to filter for desired ancestry (e.g., EUR).
   - Output: `eur_samples.txt` (or similar).
3. **Download VCF(s):**
   - Use `wget -N` for required chromosomes.
4. **Insert Panel Data into DuckDB:**
   - Use `duckdb` CLI or `duckdb` SQL script to create and populate a `reference_panel` table.
   - Table schema example:
     | Column     | Type   | Description                |
     |------------|--------|----------------------------|
     | sample_id  | TEXT   | 1000G sample ID            |
     | ancestry   | TEXT   | Ancestry group (e.g., EUR) |
     | sex        | TEXT   | Reported sex               |
   - Example SQL:
     ```sql
     CREATE TABLE IF NOT EXISTS reference_panel (
         sample_id TEXT,
         ancestry TEXT,
         sex TEXT
     );
     COPY reference_panel FROM 'eur_samples.txt' (DELIMITER '\t', HEADER FALSE);
     ```
5. **Log completion and any errors.**

---

## Inputs
- 1000G panel file URL (default: `https://ftp.1000genomes.ebi.ac.uk/vol1/ftp/release/20130502/integrated_call_samples_v3.20130502.ALL.panel`)
- VCF base URL (default: `https://ftp.1000genomes.ebi.ac.uk/vol1/ftp/release/20130502`)
- Target ancestry (e.g., EUR)
- DuckDB database path

## Outputs
- Downloaded panel and VCF files
- Populated `reference_panel` table in DuckDB
- Log files or console output

## Required Tests
- Panel and sample data downloaded and inserted only if missing
- Table schema matches requirements and is always present after DB rebuild
- Errors are logged and script exits with nonzero status on failure
- Database can be rebuilt from scratch using `build_db.sh` followed by data ingestion scripts

## Reference Panel Data Preparation Script

To support the polygenic-risk-score pipeline, a new script (e.g., `scripts/reference_script.sh`) must be created with the following minimal responsibilities:

### 1. Download the 1000 Genomes Sample Panel
- Download the panel file from:
  - `https://ftp.1000genomes.ebi.ac.uk/vol1/ftp/release/20130502/integrated_call_samples_v3.20130502.ALL.panel`
- Save to a local data directory (e.g., `data/`).

### 2. Filter for Target Ancestry
- Extract sample IDs for the desired ancestry (e.g., EUR) from the panel file.
- Output a text file with one sample ID per line (e.g., `eur_samples.txt`).

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
