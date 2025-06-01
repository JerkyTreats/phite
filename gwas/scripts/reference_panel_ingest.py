import os
import requests
import duckdb
import pandas as pd
import shutil
import logging

# Config
METADATA_URL = "https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/vcf/genomes/gnomad.genomes.v3.1.2.hgdp_1kg_subset_sample_meta.tsv.bgz"
DATA_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '../data'))
TSV_BGZ = os.path.join(DATA_DIR, 'gnomad.genomes.v3.1.2.hgdp_1kg_subset_sample_meta.tsv.bgz')
TSV = os.path.join(DATA_DIR, 'gnomad.genomes.v3.1.2.hgdp_1kg_subset_sample_meta.tsv')
DB_PATH = os.path.abspath(os.path.join(os.path.dirname(__file__), '../gwas.duckdb'))
TABLE_NAME = 'reference_panel'

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s')

def download_metadata():
    os.makedirs(DATA_DIR, exist_ok=True)
    if not os.path.exists(TSV_BGZ):
        logging.info(f"Downloading metadata from {METADATA_URL}")
        with requests.get(METADATA_URL, stream=True) as r:
            r.raise_for_status()
            with open(TSV_BGZ, 'wb') as f:
                shutil.copyfileobj(r.raw, f)
    else:
        logging.info("Metadata .bgz already exists; skipping download.")

def decompress_metadata():
    if not os.path.exists(TSV):
        import subprocess
        logging.info("Decompressing .bgz file...")
        subprocess.run(["bgzip", "-d", "-c", TSV_BGZ], stdout=open(TSV, 'wb'), check=True)
    else:
        logging.info("Metadata TSV already decompressed; skipping.")

def ingest_metadata():
    logging.info("Loading metadata TSV into DataFrame...")
    df = pd.read_csv(TSV, sep='\t')
    logging.info(f"Loaded {len(df)} rows. Ingesting into DuckDB...")
    con = duckdb.connect(DB_PATH)
    # Idempotency: drop and recreate table for clean ingest
    con.execute(f"DELETE FROM {TABLE_NAME}")
    con.register('df_view', df)
    con.execute(f"INSERT INTO {TABLE_NAME} SELECT * FROM df_view")
    con.close()
    logging.info("Ingestion complete.")

def main():
    download_metadata()
    decompress_metadata()
    ingest_metadata()

if __name__ == "__main__":
    main()
