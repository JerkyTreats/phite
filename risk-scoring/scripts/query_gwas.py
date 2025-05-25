"""
PHITE Risk Scoring: GWAS Query Module

Efficiently filter GWAS association files (.parquet or .duckdb) for a list of rsid values.
Validates schema and file type, supports large-scale filtering, and returns a pandas DataFrame.
"""
import os
from typing import List, Union
import pandas as pd

REQUIRED_COLUMNS = [
    'rsid', 'risk_allele', 'pvalue', 'beta', 'trait', 'trait_uri', 'study_id',
    'mapped_gene', 'upstream_gene_id', 'downstream_gene_id', 'snp_gene_ids',
    'chr', 'chr_pos', 'context', 'is_intergenic', 'risk_allele_freq', 'ci_95_text'
]

class GWASQueryError(Exception):
    """Custom exception for GWAS query errors."""
    pass

def query_gwas(
    gwas_path: str,
    rsid_list: List[str]
) -> pd.DataFrame:
    """
    Filter a GWAS association file for a list of rsid values.
    Args:
        gwas_path (str): Path to .parquet or .duckdb GWAS file.
        rsid_list (List[str]): List of rsid values to filter (can be large).
    Returns:
        pd.DataFrame: Filtered DataFrame with associations for input rsids.
    Raises:
        GWASQueryError: For validation or processing errors.
    """
    import pandas as pd
    import duckdb
    import os

    # Validate file existence
    if not os.path.exists(gwas_path):
        raise GWASQueryError("File does not exist: {}".format(gwas_path))

    # Validate file type
    ext = os.path.splitext(gwas_path)[1].lower()
    if ext not in {'.parquet', '.duckdb'}:
        raise GWASQueryError("Unsupported file type: {}. Only .parquet and .duckdb are supported.".format(ext))

    # If empty rsid list, return empty DataFrame with correct columns
    if not rsid_list:
        return pd.DataFrame(columns=REQUIRED_COLUMNS)

    try:
        if ext == '.parquet':
            # Use pyarrow backend for speed
            df = pd.read_parquet(gwas_path, engine='pyarrow')
        elif ext == '.duckdb':
            # For DuckDB, assume a table named 'associations_clean' or auto-detect
            con = duckdb.connect(gwas_path, read_only=True)
            # Try to auto-detect the associations table
            tables = [row[0] for row in con.execute("SHOW TABLES").fetchall()]
            if 'associations_clean' in tables:
                table = 'associations_clean'
            elif tables:
                table = tables[0]  # fallback to first table
            else:
                raise GWASQueryError("No tables found in DuckDB file.")
            # Use DuckDB's efficient filtering for large lists
            # Convert rsid_list to a DuckDB temp table for join
            con.execute("CREATE TEMP TABLE filter_rsid (rsid VARCHAR)")
            # Chunk insert to avoid parameter limits
            chunk_size = 50000
            for i in range(0, len(rsid_list), chunk_size):
                chunk = rsid_list[i:i+chunk_size]
                con.execute("INSERT INTO filter_rsid VALUES {}".format(
                    ','.join([f"('{r}')" for r in chunk])
                ))
            query = f"""
                SELECT a.* FROM {table} a
                INNER JOIN filter_rsid f ON a.rsid = f.rsid
            """
            df = con.execute(query).df()
            con.close()
        else:
            raise GWASQueryError("Unsupported file type: {}".format(ext))
    except Exception as e:
        raise GWASQueryError(f"Failed to load GWAS file: {e}")

    # Validate required columns
    missing = [col for col in REQUIRED_COLUMNS if col not in df.columns]
    if missing:
        raise GWASQueryError(f"Missing required columns: {', '.join(missing)}")

    # Filter for rsid (for Parquet, do this in pandas for now)
    if ext == '.parquet':
        # Use set for fast lookup, but preserve DataFrame order
        rsid_set = set(rsid_list)
        df = df[df['rsid'].isin(rsid_set)]

    # Ensure output columns are in correct order
    df = df[[col for col in REQUIRED_COLUMNS]]
    return df.reset_index(drop=True)
