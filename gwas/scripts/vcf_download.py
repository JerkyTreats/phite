import os
import requests
import logging

BASE_URL = "https://storage.googleapis.com/gcp-public-data--gnomad/release/3.1.2/vcf/genomes"
DEST_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '../gnomad_grch38_vcf'))
CHROMOSOMES = list(range(1, 23))
EXTENSIONS = ["vcf.bgz", "vcf.bgz.tbi"]

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s')

def download_file(url, dest):
    if os.path.exists(dest):
        logging.info(f"File already exists, skipping: {dest}")
        return
    logging.info(f"Downloading {url} -> {dest}")
    with requests.get(url, stream=True) as r:
        r.raise_for_status()
        with open(dest, 'wb') as f:
            for chunk in r.iter_content(chunk_size=8192):
                f.write(chunk)
    logging.info(f"Download complete: {dest}")

def main():
    os.makedirs(DEST_DIR, exist_ok=True)
    for chrom in CHROMOSOMES:
        for ext in EXTENSIONS:
            fname = f"gnomad.genomes.v3.1.2.sites.chr{chrom}.{ext}"
            url = f"{BASE_URL}/{fname}"
            dest = os.path.join(DEST_DIR, fname)
            try:
                download_file(url, dest)
            except Exception as e:
                logging.error(f"Failed to download {url}: {e}")

if __name__ == "__main__":
    main()
