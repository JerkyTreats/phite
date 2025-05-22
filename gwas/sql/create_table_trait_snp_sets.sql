CREATE OR REPLACE TABLE trait_ontology_map AS
SELECT
  "Disease trait" AS trait,
  "EFO URI" AS trait_uri,
  "Parent term" AS topic,
  "Parent URI" AS topic_uri
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true);

CREATE OR REPLACE TABLE trait_snp_sets_with_topics AS
SELECT
  m.topic,
  m.topic_uri,
  s.*
FROM trait_snp_sets s
JOIN trait_ontology_map m
  ON s.trait_uri = m.trait_uri;
