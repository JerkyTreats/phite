-- Step 1: Load raw data first
CREATE OR REPLACE TABLE associations_raw AS
SELECT * FROM read_csv_auto('associations.tsv', delim='\t', header=true);

-- Step 2: Normalize trait_uri by exploding comma-separated URIs into multiple rows
CREATE OR REPLACE TABLE associations_clean AS
SELECT DISTINCT
  "SNPS" AS rsid,
  SPLIT_PART("STRONGEST SNP-RISK ALLELE", '-', 2) AS risk_allele,
  TRY_CAST("P-VALUE" AS DOUBLE) AS pvalue,
  TRY_CAST("OR or BETA" AS DOUBLE) AS beta,
  "MAPPED_TRAIT" AS trait,
  TRIM(uri.unnest) AS trait_uri,
  "STUDY ACCESSION" AS study_id,

  -- Gene fields
  "MAPPED_GENE" AS mapped_gene,
  "UPSTREAM_GENE_ID" AS upstream_gene_id,
  "DOWNSTREAM_GENE_ID" AS downstream_gene_id,
  "SNP_GENE_IDS" AS snp_gene_ids,

  -- Additional metadata
  "CHR_ID" AS chr,
  "CHR_POS" AS chr_pos,
  "CONTEXT" AS context,
  "INTERGENIC" AS is_intergenic,
  "RISK ALLELE FREQUENCY" AS risk_allele_freq,
  "95% CI (TEXT)" AS ci_95_text

FROM associations_raw,
     UNNEST(STRING_SPLIT("MAPPED_TRAIT_URI", ',')) AS uri

WHERE
  "SNPS" IS NOT NULL AND
  TRY_CAST("P-VALUE" AS DOUBLE) IS NOT NULL AND
  TRY_CAST("OR or BETA" AS DOUBLE) IS NOT NULL AND
  "STRONGEST SNP-RISK ALLELE" LIKE 'rs%-%' AND
  "MAPPED_TRAIT" IS NOT NULL AND
  "MAPPED_TRAIT_URI" IS NOT NULL;
