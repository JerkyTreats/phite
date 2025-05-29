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


EOF

echo "✅ Rebuild complete: gwas/gwas.duckdb"
