-- Create the studies_clean table from the studies.tsv file
CREATE OR REPLACE TABLE studies_clean AS
SELECT DISTINCT
  "STUDY ACCESSION" AS study_id,
  "PUBMED ID" AS pubmed_id,
  "FIRST AUTHOR" AS author,
  "DATE" AS pub_date,
  "JOURNAL" AS journal,
  "INITIAL SAMPLE SIZE" AS initial_sample_size,
  "REPLICATION SAMPLE SIZE" AS replication_sample_size,
  "PLATFORM [SNPS PASSING QC]" AS platform,
  "GENOTYPING TECHNOLOGY" AS genotyping_tech,
  "COHORT" AS cohort,
  "FULL SUMMARY STATISTICS" AS has_summary_stats,
  "SUMMARY STATS LOCATION" AS summary_url
FROM read_csv_auto('gwas/studies.tsv', delim='\t', header=true)
WHERE "STUDY ACCESSION" IS NOT NULL;
