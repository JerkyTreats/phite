import os
import tempfile
import pandas as pd
import pytest
from scripts.load_user_genotype import load_user_genotype, GenotypeFileError

def write_temp_file(contents, suffix):
    fd, path = tempfile.mkstemp(suffix=suffix)
    with os.fdopen(fd, 'w') as tmp:
        tmp.write(contents)
    return path

def test_valid_ancestrydna_txt():
    contents = "rsid\tgenotype\nrs12345\tAA\nrs67890\tAG\n"
    path = write_temp_file(contents, '.txt')
    df = load_user_genotype(path)
    assert list(df.columns) == ['rsid', 'genotype']
    assert df.shape == (2, 2)
    os.remove(path)

def test_valid_tsv_allele12():
    contents = "rsid\tchromosome\tposition\tallele1\tallele2\nrs190214723\t1\t693625\tT\tT\n"
    path = write_temp_file(contents, '.txt')
    df = load_user_genotype(path)
    assert 'rsid' in df.columns
    assert 'genotype' in df.columns
    assert df.iloc[0]['rsid'] == 'rs190214723'
    assert df.iloc[0]['genotype'] == 'TT'
    os.remove(path)

def test_valid_23andme_csv():
    contents = "rsid,genotype\nrs54321,GG\nrs09876,TT\n"
    path = write_temp_file(contents, '.csv')
    df = load_user_genotype(path)
    assert any('rsid' in c for c in df.columns)
    assert any('genotype' in c for c in df.columns)
    os.remove(path)

def test_invalid_file_type():
    contents = "rsid,genotype\nrs1,AA\n"
    path = write_temp_file(contents, '.xlsx')
    with pytest.raises(GenotypeFileError) as e:
        load_user_genotype(path)
    assert 'Unsupported file type' in str(e.value)
    os.remove(path)

def test_malformed_file():
    contents = "rsid|genotype\nrs1|AA\n"
    path = write_temp_file(contents, '.csv')
    df = load_user_genotype(path)
    assert isinstance(df, pd.DataFrame)
    assert df.shape[0] == 1
    assert df.columns[0] == 'rsid|genotype'
    os.remove(path)

def test_empty_file():
    contents = ""
    path = write_temp_file(contents, '.csv')
    with pytest.raises(GenotypeFileError):
        load_user_genotype(path)
    os.remove(path)
