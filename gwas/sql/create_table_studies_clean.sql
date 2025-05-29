-- Step 1: Create table with constraint
CREATE TABLE studies_clean (
  study_id TEXT PRIMARY KEY,
  pubmed_id TEXT,
  author TEXT,
  pub_date TEXT,
  journal TEXT,
  initial_sample_size TEXT,
  replication_sample_size TEXT,
  platform TEXT,
  genotyping_tech TEXT,
  cohort TEXT,
  has_summary_stats TEXT,
  summary_url TEXT
);

-- Step 2: Insert data
INSERT INTO studies_clean
SELECT DISTINCT
  "STUDY ACCESSION",
  "PUBMED ID",
  "FIRST AUTHOR",
  "DATE",
  "JOURNAL",
  "INITIAL SAMPLE SIZE",
  "REPLICATION SAMPLE SIZE",
  "PLATFORM [SNPS PASSING QC]",
  "GENOTYPING TECHNOLOGY",
  "COHORT",
  "FULL SUMMARY STATISTICS",
  "SUMMARY STATS LOCATION"
FROM read_csv_auto('studies.tsv', delim='\t', header=true)
WHERE "STUDY ACCESSION" IS NOT NULL;