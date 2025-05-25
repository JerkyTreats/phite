"""
Unit tests for risk-scoring/scripts/query_gwas.py
"""
import os
import tempfile
import pandas as pd
import pytest
from scripts.query_gwas import query_gwas, GWASQueryError, REQUIRED_COLUMNS

def make_gwas_df(rsids):
    data = {col: [f"val_{col}_{i}" for i in range(len(rsids))] for col in REQUIRED_COLUMNS}
    data['rsid'] = rsids
    data['pvalue'] = [0.05] * len(rsids)
    data['beta'] = [0.1] * len(rsids)
    return pd.DataFrame(data)

def test_filtering_returns_only_matching_rsids(tmp_path):
    df = make_gwas_df(['rs1', 'rs2', 'rs3'])
    parquet_path = tmp_path / "gwas.parquet"
    df.to_parquet(parquet_path)
    result = query_gwas(str(parquet_path), ['rs2', 'rs3'])
    assert set(result['rsid']) == {'rs2', 'rs3'}

def test_output_matches_schema(tmp_path):
    df = make_gwas_df(['rs1', 'rs2'])
    parquet_path = tmp_path / "gwas.parquet"
    df.to_parquet(parquet_path)
    result = query_gwas(str(parquet_path), ['rs1'])
    assert list(result.columns) == REQUIRED_COLUMNS

def test_empty_input_returns_empty_df(tmp_path):
    df = make_gwas_df(['rs1'])
    parquet_path = tmp_path / "gwas.parquet"
    df.to_parquet(parquet_path)
    result = query_gwas(str(parquet_path), [])
    assert result.empty
    assert list(result.columns) == REQUIRED_COLUMNS

def test_missing_rsid_column_raises(tmp_path):
    df = make_gwas_df(['rs1'])
    df = df.drop(columns=['rsid'])
    parquet_path = tmp_path / "gwas.parquet"
    df.to_parquet(parquet_path)
    with pytest.raises(GWASQueryError, match="Missing required columns: rsid"):
        query_gwas(str(parquet_path), ['rs1'])

def test_missing_file_raises():
    with pytest.raises(GWASQueryError, match="File does not exist"):
        query_gwas("/tmp/nonexistent.parquet", ['rs1'])

def test_invalid_file_type_raises(tmp_path):
    txt_path = tmp_path / "gwas.txt"
    with open(txt_path, "w") as f:
        f.write("not a gwas file")
    with pytest.raises(GWASQueryError, match="Unsupported file type"):
        query_gwas(str(txt_path), ['rs1'])

def test_performance_large_input(tmp_path):
    # Use a small DataFrame but a large rsid list to test scalability
    df = make_gwas_df([f'rs{i}' for i in range(100)])
    parquet_path = tmp_path / "gwas.parquet"
    df.to_parquet(parquet_path)
    large_rsid_list = [f'rs{i}' for i in range(100)] * 6000  # 600k entries
    result = query_gwas(str(parquet_path), large_rsid_list)
    assert set(result['rsid']) == set(df['rsid'])
