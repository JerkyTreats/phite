-- Trait SNP Sets With Topics
CREATE OR REPLACE TABLE trait_snp_sets_with_topics AS
SELECT
  c.parent_term AS topic,
  c.parent_uri AS topic_uri,
  s.*
FROM trait_snp_sets s
JOIN trait_concepts c ON s.trait_uri = c.trait_uri;
