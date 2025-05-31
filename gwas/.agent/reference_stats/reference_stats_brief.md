# Reference Stats Data Engineering Agent Brief

> **Location:** This brief and all GWAS data engineering is located in the shared, normalized GWAS database directory (`gwas/`), outside of any specific application module.


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
- 1000G panel file URL
- VCF base URL
- Target ancestry (e.g., EUR)
- DuckDB database path

## Outputs
- Downloaded panel and VCF files
- Populated `reference_panel` table in DuckDB
- Log files or console output

## Required Tests
- Panel and sample data downloaded and inserted only if missing
- Table schema matches requirements
- Errors are logged and script exits with nonzero status on failure

## Example Invocation
```bash
bash scripts/run_reference_script.sh --db gwas.duckdb --ancestry EUR --chr 1
```

---

## Notes
- This brief covers only panel/sample ingestion. PRS calculation and stats insertion are handled elsewhere.
- For full reproducibility, version input files and document script versions.
