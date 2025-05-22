-- Trait Concepts
CREATE OR REPLACE TABLE trait_concepts AS
SELECT DISTINCT
  "EFO URI" AS trait_uri,
  "EFO term" AS efo_term,
  "Parent term" AS parent_term,
  "Parent URI" AS parent_uri
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true)
WHERE "EFO URI" IS NOT NULL AND "Parent term" IS NOT NULL;
