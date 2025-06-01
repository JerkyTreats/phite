#!/bin/bash

set -e  # Exit on any error

echo "📦 Rebuilding gwas.duckdb..."

# Remove old DB if it exists
if [ -f gwas.duckdb ]; then
  echo "🗑️ Removing existing DuckDB..."
  rm gwas.duckdb
fi


# Create new DuckDB and execute SQL
echo "🧱 Creating new DuckDB and executing schema scripts..."
duckdb gwas.duckdb <<'EOF'
-- Create main schema tables
.read sql/create_table_associations_clean.sql
.read sql/create_table_studies_clean.sql
.read sql/create_table_traits.sql
.read sql/create_table_reference_panel.sql
.read sql/create_table_reference_stats.sql
EOF

echo "🚀 Running reference panel setup (Python + venv) ..."
bash scripts/setup_reference_panel.sh
echo "✅ Reference panel setup complete."

echo "🚀 Running VCF download setup (Python + venv) ..."
bash scripts/setup_vcf_download.sh
echo "✅ VCF download complete."
echo "✅ Rebuild complete: gwas/gwas.duckdb"