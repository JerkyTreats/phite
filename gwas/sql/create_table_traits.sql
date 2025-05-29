-- Trait Concepts
CREATE OR REPLACE TABLE trait_concepts AS
SELECT DISTINCT
  "EFO URI" AS trait_uri,
  "EFO term" AS efo_term,
  "Parent term" AS parent_term,
  "Parent URI" AS parent_uri
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true)
WHERE "EFO URI" IS NOT NULL AND "Parent term" IS NOT NULL;

-- Trait Labels
CREATE OR REPLACE TABLE trait_labels AS
SELECT DISTINCT
  "Disease trait" AS trait_label,
  "EFO URI" AS trait_uri
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true)
WHERE "EFO URI" IS NOT NULL;

-- Trait SNP Sets With Topics
CREATE OR REPLACE TABLE trait_snp_sets_with_topics AS
SELECT
  c.parent_term AS topic,
  c.parent_uri AS topic_uri,
  s.*
FROM trait_snp_sets s
JOIN trait_concepts c ON s.trait_uri = c.trait_uri;

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
