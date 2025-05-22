#!/bin/bash

set -e  # Exit on any error

echo "📦 Rebuilding gwas.duckdb..."

# Remove old DB if it exists
if [ -f gwas/gwas.duckdb ]; then
  echo "🗑️ Removing existing DuckDB..."
  rm gwas/gwas.duckdb
fi

# Create new DuckDB and execute SQL
echo "🧱 Creating new DuckDB and executing schema scripts..."
duckdb gwas/gwas.duckdb <<EOF
.read gwas/sql/create_table_associations_clean.sql
.read gwas/sql/create_table_trait_labels.sql
.read gwas/sql/create_table_trait_concepts.sql
.read gwas/sql/create_table_trait_snp_sets.sql
.read gwas/sql/create_table_trait_snp_sets_with_topics.sql
EOF

echo "✅ Rebuild complete: gwas/gwas.duckdb"
