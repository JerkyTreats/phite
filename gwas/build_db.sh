#!/bin/bash

set -e  # Exit on any error

DB="${1:-gwas}"

echo "📦 Rebuilding ${DB}.duckdb..."

# Remove old DB if it exists
if [ -f "${DB}.duckdb" ]; then
  echo "🗑️ Removing existing DuckDB..."
  rm "${DB}.duckdb"
fi


# Create new DuckDB and execute SQL
echo "🧱 Creating new DuckDB and executing schema scripts..."
duckdb "${DB}.duckdb" <<'EOF'
-- Load associations.tsv into associations_raw (required by associations_clean)
CREATE OR REPLACE TABLE associations_raw AS
SELECT * FROM read_csv_auto('associations.tsv', delim='\t', header=true);

-- Create main schema tables
.read sql/create_table_associations_clean.sql
.read sql/create_table_studies_clean.sql
.read sql/create_table_traits.sql
EOF


echo "⚙️  Enriching DuckDB with reference/alternate alleles…"
bash scripts/reference_allele_enrichment/enrich_with_alleles.sh --duckdb "${DB}.duckdb" --gcp_project jerkytreats

echo "✅ Rebuild complete: gwas/${DB}.duckdb"
