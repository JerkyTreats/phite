import os
import pandas as pd
from typing import Union

class GenotypeFileError(Exception):
    """Custom exception for genotype file errors."""
    pass

def load_user_genotype(filepath: str) -> pd.DataFrame:
    """
    Loads a user genotype file (AncestryDNA or 23andMe format) and returns a DataFrame with columns: rsid, genotype.

    Parameters:
        filepath (str): Path to the genotype file (.txt or .csv)

    Returns:
        pd.DataFrame: DataFrame with columns 'rsid', 'genotype'

    Raises:
        GenotypeFileError: For invalid file types, missing columns, or malformed files.
    """
    # Validate file existence
    if not os.path.isfile(filepath):
        raise GenotypeFileError(f"File not found: {filepath}")

    # Validate file extension
    ext = os.path.splitext(filepath)[1].lower()
    if ext not in ['.txt', '.csv']:
        raise GenotypeFileError(f"Unsupported file type: {ext}. Only .txt and .csv are supported.")

    # Try reading file
    try:
        for sep in ['\t', ',', ' ']:
            try:
                temp_df = pd.read_csv(filepath, sep=sep, comment='#', dtype=str)
                columns = [c.lower() for c in temp_df.columns]
                temp_df.columns = columns
                # If allele1 and allele2 present, concatenate for genotype
                if 'rsid' in columns and 'allele1' in columns and 'allele2' in columns:
                    temp_df['genotype'] = temp_df['allele1'].astype(str) + temp_df['allele2'].astype(str)
                return temp_df
            except Exception:
                continue
        raise GenotypeFileError("Could not parse file with expected delimiters.")
    except Exception as e:
        raise GenotypeFileError(f"Error reading file: {e}")

    # Validate required columns
    missing = [col for col in ['rsid', 'genotype'] if col not in df.columns]
    if missing:
        raise GenotypeFileError(f"Input file is missing required column(s): {', '.join(missing)}.")

    # Output only required columns and drop rows with missing values
    df = df[['rsid', 'genotype']].dropna()
    if df.empty:
        raise GenotypeFileError("No valid genotype data found after filtering.")
    return df
