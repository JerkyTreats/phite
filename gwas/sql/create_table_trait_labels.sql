-- Trait Labels
CREATE OR REPLACE TABLE trait_labels AS
SELECT DISTINCT
  "Disease trait" AS trait_label,
  "EFO URI" AS trait_uri
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true)
WHERE "EFO URI" IS NOT NULL;
