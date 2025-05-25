# PHITE Risk Scoring Pipeline

This directory contains the core modules for the PHITE polygenic risk scoring pipeline. It enables local, privacy-preserving computation of polygenic risk scores (PRS) and trait-based risk summaries from user genotype and GWAS data.

---

## Developer Setup

### 1. Clone the Repository
```bash
git clone <your-phite-repo-url>
cd PHITE/risk-scoring
```

### 2. Create and Activate a Virtual Environment
It is recommended to use [venv](https://docs.python.org/3/library/venv.html) for isolated Python environments.

```bash
python3 -m venv .venv
source .venv/bin/activate
```

### 3. Install Dependencies
Install all required Python packages:

```bash
pip install --upgrade pip
pip install -r requirements.txt
```

The main dependencies are listed in `requirements.txt`:

```
pandas>=2.0.0
jinja2>=3.0.0
duckdb>=0.8.0
markdown>=3.4.0
pytest>=7.0.0
```
Install them with:

```bash
pip install -r requirements.txt
```
### 4. Directory Structure
- `scripts/` – Core pipeline scripts (e.g., `load_user_genotype.py`, `query_gwas.py`, etc.)
- `reports/` – Output reports (Markdown or HTML)
- `.agent/` – Agent system briefs, specs, and audit outputs

### 5. Running Scripts
Example (replace with actual script names and arguments):

```bash
python scripts/load_user_genotype.py --input path/to/genotype.csv
python scripts/query_gwas.py --input path/to/genotype.parquet
python scripts/report_generator.py --prs path/to/prs.csv --grouping path/to/grouping.csv
```

### 6. Testing

Add your tests in a `tests/` directory. To run all tests using [pytest](https://docs.pytest.org/):

```bash
pytest tests/
```

```bash
pip install pytest
pytest tests/
```

---

## Additional Notes
- All processing remains local; no data is transmitted externally.
- See `.agent/data_model_spec.md` for canonical data structure specifications and required DataFrame schemas.
- For feature briefs and implementation details, see `.agent/feature-*.md` files.

---

## License
[Add license information here]
