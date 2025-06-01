-- ========================================
-- Trait Concept Normalization Pipeline
-- ========================================

-- STEP 1: Normalize Trait Concepts from GWAS Catalog TSV
CREATE OR REPLACE TABLE trait_concepts (
    trait_uri TEXT PRIMARY KEY,       -- e.g. http://www.ebi.ac.uk/efo/EFO_0004468
    trait_label TEXT,                 -- e.g. glucose measurement
    parent_uri TEXT,                  -- e.g. http://www.ebi.ac.uk/efo/EFO_0000400
    parent_label TEXT                 -- e.g. metabolic trait
);

INSERT INTO trait_concepts
SELECT DISTINCT
    "EFO URI",
    "EFO term",
    "Parent URI",
    "Parent term"
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true)
WHERE "EFO URI" IS NOT NULL;

-- STEP 2: Build a Trait Topic Map (trait_uri → topic_uri)
CREATE TABLE IF NOT EXISTS trait_topic_map (
  trait_id TEXT,
  topic_id TEXT,
  topic_label TEXT,
  PRIMARY KEY (trait_id, topic_id)
);

INSERT INTO trait_topic_map
SELECT DISTINCT
  REPLACE(SPLIT_PART("EFO URI", '/', -1), '#', '') AS trait_id,
  REPLACE(SPLIT_PART("Parent URI", '/', -1), '#', '') AS topic_id,
  "Parent term" AS topic_label
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true)
WHERE "EFO URI" IS NOT NULL AND "Parent term" IS NOT NULL;


-- STEP 3: (Optional) Trait Labels (for display in UI or export)
CREATE OR REPLACE TABLE trait_labels AS
SELECT DISTINCT
  REPLACE(SPLIT_PART("EFO URI", '/', -1), '#', '') AS trait_id,
  TRIM("Disease trait") AS trait_label
FROM read_csv_auto('gwas_catalog_trait-mappings_r2025-05-13.tsv', delim='\t', header=true)
WHERE "EFO URI" IS NOT NULL;


-- -- STEP 4: SNP-to-Topic Join (if snp_trait_link exists)
-- -- Joins each SNP to its associated trait's high-level topic
-- CREATE OR REPLACE TABLE snp_trait_topics AS
-- SELECT
--     l.rsid,
--     m.topic_uri
-- FROM snp_trait_link l
-- JOIN trait_topic_map m
--   ON l.trait_id = m.trait_uri;

-- -- STEP 5: (Optional) Enrich snp_trait_topics with topic label
-- CREATE OR REPLACE TABLE snp_trait_topics_labeled AS
-- SELECT
--     s.rsid,
--     s.topic_uri,
--     c.parent_label AS topic_label
-- FROM snp_trait_topics s
-- JOIN trait_concepts c ON s.topic_uri = c.parent_uri;