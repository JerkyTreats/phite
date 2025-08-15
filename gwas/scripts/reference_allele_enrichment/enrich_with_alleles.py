import argparse
import logging
import os
import gc
from pathlib import Path
from typing import List, Optional

import duckdb
import pandas as pd
from tqdm import tqdm
import time

# Optional BigQuery imports.  We import lazily so users who already have local
# Parquet files can run the script without Google Cloud dependencies installed.
try:
    from google.cloud import bigquery
except ImportError:  # pragma: no cover
    bigquery = None  # type: ignore


LOGGER = logging.getLogger("gwas.enrich_with_alleles")
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")

# Chromosomes to pull from gnomAD (v3 genomes tables are chr1-22, X, Y)
CHROMS: List[str] = [str(c) for c in range(1, 23)] + ["X", "Y"]

# BigQuery public dataset template
BQ_TABLE_TMPL = "bigquery-public-data.gnomAD.v3_genomes__chr{chrom}"

# DuckDB DDL for the allele reference table
CREATE_SNP_ALLELES_SQL = """
CREATE TABLE IF NOT EXISTS snp_alleles (
    rsid        VARCHAR PRIMARY KEY,
    chr         VARCHAR,
    chr_pos     BIGINT,
    ref_allele  VARCHAR,
    alt_allele  VARCHAR
);
"""

# Convenience view combining associations_clean with allele data
CREATE_VIEW_SQL = """
CREATE OR REPLACE VIEW associations_with_alleles AS
SELECT a.*, s.ref_allele, s.alt_allele
FROM associations_clean a
LEFT JOIN snp_alleles s USING (rsid);
"""

# Lightweight import state to support resuming by chromosome
CREATE_IMPORT_STATE_SQL = """
CREATE TABLE IF NOT EXISTS allele_import_state (
    chrom        VARCHAR PRIMARY KEY,
    parquet_file VARCHAR,
    imported_at  TIMESTAMP DEFAULT now()
);
"""

def run_query_to_parquet(chrom: str, client: "bigquery.Client", out_dir: Path) -> Path:
    """Executes a BigQuery query for a single chromosome and writes the result to Parquet."""
    table_id = BQ_TABLE_TMPL.format(chrom=chrom)
    LOGGER.info("Querying allele data for chromosome %s", chrom)

    query = f"""
    SELECT
      name AS rsid,
      reference_name AS chr,
      start_position + 1 AS chr_pos,
      reference_bases AS ref_allele,
      alternate_bases[OFFSET(0)].alt AS alt_allele
    FROM `{table_id}`
    CROSS JOIN UNNEST(names) AS name
    WHERE STARTS_WITH(name, 'rs')
    """

    job = client.query(query)
    df = job.result().to_dataframe(create_bqstorage_client=True, progress_bar_type="tqdm")

    dest = out_dir / f"alleles_chr{chrom}.parquet"
    LOGGER.info("Writing %d rows → %s", len(df), dest)
    df.to_parquet(dest, index=False)
    return dest

def ensure_parquet_files(parquet_dir: Path, gcp_project: Optional[str] = None) -> List[Path]:
    """Return list of allele Parquet files; download from BigQuery if none exist."""
    parquet_dir.mkdir(parents=True, exist_ok=True)
    existing = sorted(parquet_dir.glob("alleles_chr*.parquet"))

    if existing:
        LOGGER.info("Found %d existing Parquet files — using cached copies.", len(existing))
        return existing

    LOGGER.info("No cached allele Parquet files found; downloading from BigQuery …")

    if bigquery is None:
        raise ImportError("google-cloud-bigquery is not installed. Run 'pip install -r gwas/scripts/reference_allele_enrichment/requirements.txt' to enable automatic download.")

    client = bigquery.Client(project=gcp_project)
    generated: List[Path] = []
    for chrom in CHROMS:
        try:
            generated.append(run_query_to_parquet(chrom, client, parquet_dir))
        except Exception as exc:  # pragma: no cover
            LOGGER.error("Failed to fetch chromosome %s: %s", chrom, exc)
    return generated

def import_into_duckdb(db_path: Path, parquet_files: List[Path], chunk_size: int = 100000):
    """Import Parquet allele files into DuckDB using chunked processing.

    Chunked workflow to prevent pin block errors:
    - Process each parquet file in small chunks
    - Commit after each chunk to reset buffer pool
    - Higher memory limits but frequent commits
    - Individual file processing with immediate cleanup
    """
    start_ts = time.time()
    total_files = len(parquet_files)

    for idx, pf in enumerate(parquet_files, start=1):
        LOGGER.info("[%d/%d] Processing %s", idx, total_files, pf.name)

        # Derive chromosome token from filename
        chrom_token = pf.name.replace("alleles_chr", "").replace(".parquet", "")

        # Fresh connection per file with generous memory limits
        con = duckdb.connect(str(db_path))
        try:
            # Higher memory limits but frequent commits to reset buffer pool
            con.execute("PRAGMA memory_limit='8GB'")
            con.execute("PRAGMA max_memory='6GB'")
            con.execute("PRAGMA threads=1")  # Single thread to reduce contention
            con.execute("PRAGMA temp_directory='/tmp'")

            # Ensure target tables exist
            con.execute(CREATE_SNP_ALLELES_SQL)
            con.execute(CREATE_IMPORT_STATE_SQL)

            # Check if already imported (resume capability)
            already = con.execute("SELECT 1 FROM allele_import_state WHERE chrom = ?", [chrom_token]).fetchone()
            if already is not None:
                LOGGER.info("[%d/%d] Skipping %s (resume): chrom %s already imported", idx, total_files, pf.name, chrom_token)
                continue

            # Get total row count for progress tracking
            try:
                total_rows = con.execute("SELECT COUNT(*) FROM parquet_scan(?)", [str(pf)]).fetchone()[0]
                LOGGER.info("[%d/%d] %s contains %s rows, processing in chunks of %s",
                           idx, total_files, pf.name, f"{total_rows:,}", f"{chunk_size:,}")
            except Exception as e:
                LOGGER.warning("Could not count rows in %s: %s", pf.name, e)
                total_rows = None

            # Process file in chunks with frequent commits
            offset = 0
            chunk_count = 0
            rows_processed = 0

            while True:
                chunk_count += 1
                LOGGER.info("[%d/%d] Processing chunk %d (offset %s) for %s",
                           idx, total_files, chunk_count, f"{offset:,}", pf.name)

                con.execute("BEGIN")
                try:
                    # Process chunk with LIMIT/OFFSET
                    result = con.execute("""
                        INSERT OR REPLACE INTO snp_alleles
                        SELECT rsid, chr, chr_pos, ref_allele, alt_allele
                        FROM parquet_scan(?)
                        LIMIT ? OFFSET ?
                    """, [str(pf), chunk_size, offset])

                    # Get number of rows inserted in this chunk
                    chunk_rows = result.fetchone()[0] if result else 0
                    con.execute("COMMIT")

                    if chunk_rows == 0:
                        LOGGER.info("[%d/%d] No more rows, finished %s", idx, total_files, pf.name)
                        break

                    rows_processed += chunk_rows
                    offset += chunk_size

                    # Progress update
                    progress_msg = f"[{idx}/{total_files}] Chunk {chunk_count}: +{chunk_rows:,} rows"
                    if total_rows:
                        pct = (rows_processed / total_rows) * 100
                        progress_msg += f" ({rows_processed:,}/{total_rows:,} = {pct:.1f}%)"
                    LOGGER.info(progress_msg)

                    # Force buffer pool reset after each chunk
                    con.execute("PRAGMA enable_optimizer")

                    # Small delay to allow memory cleanup
                    time.sleep(0.1)

                except Exception as e:
                    LOGGER.error("Failed chunk %d for %s: %s", chunk_count, pf.name, e)
                    try:
                        con.execute("ROLLBACK")
                    except Exception:
                        pass
                    raise

            # Mark chromosome as successfully imported
            con.execute(
                "INSERT OR REPLACE INTO allele_import_state (chrom, parquet_file) VALUES (?, ?)",
                [chrom_token, str(pf)]
            )
            con.commit()  # Ensure completion state is persisted

            LOGGER.info("[%d/%d] ✅ Completed %s: %s total rows processed",
                       idx, total_files, pf.name, f"{rows_processed:,}")

        finally:
            con.close()

        # Force garbage collection after each file
        gc.collect()

    # Final optimization: create index and view
    LOGGER.info("Finalizing: creating index and view …")
    con = duckdb.connect(str(db_path))
    try:
        con.execute("PRAGMA memory_limit='12GB'")
        con.execute("CREATE INDEX IF NOT EXISTS snp_alleles_rsid ON snp_alleles(rsid);")
        con.execute(CREATE_VIEW_SQL)
        con.execute("CHECKPOINT")
    finally:
        con.close()

    elapsed = time.time() - start_ts
    LOGGER.info("DuckDB enrichment complete ✅  Processed %d files in %.1fs", total_files, elapsed)

def main():
    parser = argparse.ArgumentParser(description="Enrich GWAS DuckDB with reference/alternate allele information from gnomAD v3")
    parser.add_argument("--duckdb", type=Path, default=Path("gwas/gwas.duckdb"), help="Path to local DuckDB database file")
    parser.add_argument("--parquet_dir", type=Path, default=Path("gwas/parquet/alleles"), help="Directory to cache Parquet allele files")
    parser.add_argument("--gcp_project", type=str, default=os.getenv("GOOGLE_CLOUD_PROJECT", "jerkytreats"), help="GCP project ID for BigQuery billing")
    parser.add_argument("--chunk_size", type=int, default=100000, help="Number of rows per chunk (lower = less memory per commit)")
    args = parser.parse_args()

    parquet_files = ensure_parquet_files(args.parquet_dir, args.gcp_project)
    import_into_duckdb(args.duckdb, parquet_files, chunk_size=args.chunk_size)

if __name__ == "__main__":
    main()
