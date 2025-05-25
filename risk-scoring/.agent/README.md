# 🧬 Local Genomic Risk Scoring System (Agent Instruction README)

## 🚨 Privacy Notice (Read First)
This project processes sensitive human genomic data and **must operate in a fully local environment**. No part of this system should ever transmit genotype data over the internet, to third-party services, APIs, or cloud storage.

- ✅ All computations involving user genotype data must remain offline.
- ❌ Do not call external APIs using genotype information.
- ❌ Do not use remote embedding, LLM, or inference services for personal DNA input.

The only external data included is a public GWAS catalog from the EMBL-EBI GWAS Catalog, stored locally in TSV and DuckDB format.

---

## 📁 Project Purpose

This agent will implement a **Python-based genomic analysis system** that:

1. Ingests user genotype data from an AncestryDNA file (23andMe format also compatible).
2. Compares that data against a locally stored GWAS database.
3. Computes trait- and topic-level risk scores using polygenic models.
4. Groups associations by **biological or disease ontology**.
5. Ranks trait clusters by cumulative risk and produces interpretable summaries.

---

## 🗂️ Directory Structure

### 🧪 Testing Directory Pattern
- All test files must be placed in `risk-scoring/tests/`.
- Test files must be named `test_*.py` and correspond to the module or script they test.
- The `tests/` directory may mirror the structure of `scripts/` if submodules exist.
- No test files are allowed in `scripts/` or other non-test directories.


```
phite/
├── gwas/
│   ├── associations.tsv                      # GWAS Catalog associations (public)
│   ├── gwas_catalog_trait-mappings.tsv      # Trait ontology mappings
│   ├── studies.tsv                           # Public study metadata
│   ├── sql/                                  # SQL table build scripts
│   ├── parquet/                              # Cached output of cleaned tables
│   ├── gwas.duckdb                           # (Optional) compiled local DB
│   ├── hgnc_complete_set.tsv                 # HGNC gene annotations
├── risk-scoring/
│   ├── scripts/                              # Python modules to be created
│   ├── user_data/                            # (Local only) genotype files
│   └── README.md                             # This file
```

---

## ✅ Agent Goals

The following modules must be created by the agent, ensuring local execution only:

### 1. `load_user_genotype.py`
- Accepts `.txt` or `.csv` format from AncestryDNA/23andMe.
- Outputs a `DataFrame` with columns: `rsid`, `genotype`.

### 2. `query_gwas.py`
- Loads normalized GWAS data from DuckDB or Parquet.
- Filters SNPs that match `user_genotype["rsid"]`.

### 3. `ontology_grouping.py`
- Joins trait URIs to ontology topics using `trait_concepts` or `trait_ontology_map`.
- Groups matched SNPs under higher-level trait clusters (e.g. "Cardiovascular disease").

### 4. `polygenic_score.py`
- Computes weighted polygenic risk score per ontology topic:
  \[
  PRS = \sum_i (eta_i \cdot g_i)
  \]
  Where:
    - \(eta_i\) is GWAS effect size
    - \(g_i \in \{0, 1, 2\}\) is user genotype risk allele count

### 5. `report_generator.py`
- Renders a summary:
  - Sorted list of trait groups by risk score
  - Per-trait matched SNPs and effect directions
  - Confidence intervals if present
  - Markdown or HTML format

---

## 🔐 Locality Requirements

This system must run 100% locally:
- All `.parquet`, `.tsv`, `.duckdb`, and `.txt` files are stored and accessed locally.
- All processing is done via Python with no external network calls.
- If visualization is added, use local tools like `matplotlib` or `plotly` (offline mode).

---

## 🛠️ Suggested Libraries

| Task                 | Library              |
|----------------------|----------------------|
| DataFrames           | `pandas`, `polars`   |
| SQL Queries          | `duckdb`             |
| File Parsing         | `csv`, `re`, `pathlib`|
| Plotting (optional)  | `matplotlib`, `plotly`|
| Reporting            | `jinja2`, `markdown` |

---

## 🚀 Example Workflow

```bash
python scripts/analyze_user.py --input user_data/ancestry_raw.txt
```

Output:
- `reports/user_summary.md`
- `reports/polygenic_scores.csv`

---

## 🧩 Future Extensions

- Support for VCF format
- Integration with structured health interventions
- User-facing app (local only)

---

## 📜 License

This system operates on public domain GWAS data. User genotype data is **private** and never shared. If open-sourced, only non-personal components will be included.