-- Step 1: Load raw data
CREATE OR REPLACE TABLE associations_raw AS
SELECT *
FROM read_csv_auto('associations.tsv', delim='\t', header=true);

-- Step 2: Normalize, explode trait_uri, and exclude unmapped intergenic entries
CREATE OR REPLACE TABLE associations_clean AS
SELECT DISTINCT
  "SNPS" AS rsid,
  TRIM(SPLIT_PART("STRONGEST SNP-RISK ALLELE", '-', 2)) AS risk_allele,
  TRY_CAST("P-VALUE" AS DOUBLE) AS pvalue,
  TRY_CAST("OR or BETA" AS DOUBLE) AS beta,
  "MAPPED_TRAIT" AS trait,
  TRIM(uri.unnest) AS trait_uri,
  "STUDY ACCESSION" AS study_id,

  -- Gene mapping
  "MAPPED_GENE" AS mapped_gene,
  "UPSTREAM_GENE_ID" AS upstream_gene_id,
  "DOWNSTREAM_GENE_ID" AS downstream_gene_id,
  "SNP_GENE_IDS" AS snp_gene_ids,

  -- Location & annotations
  "CHR_ID" AS chr,
  "CHR_POS" AS chr_pos,
  "CONTEXT" AS context,
  "INTERGENIC" AS is_intergenic,
  "RISK ALLELE FREQUENCY" AS risk_allele_freq,
  "95% CI (TEXT)" AS ci_95_text

FROM associations_raw,
     UNNEST(STRING_SPLIT("MAPPED_TRAIT_URI", ',')) AS uri

WHERE
  -- Valid core fields
  "SNPS" IS NOT NULL AND
  TRY_CAST("P-VALUE" AS DOUBLE) IS NOT NULL AND
  TRY_CAST("OR or BETA" AS DOUBLE) IS NOT NULL AND
  "STRONGEST SNP-RISK ALLELE" LIKE 'rs%-%' AND
  "MAPPED_TRAIT" IS NOT NULL AND
  "MAPPED_TRAIT_URI" IS NOT NULL

  -- Exclude fully unmapped intergenic variants
  AND NOT (
    "MAPPED_GENE" IS NULL AND
    "UPSTREAM_GENE_ID" IS NULL AND
    "DOWNSTREAM_GENE_ID" IS NULL AND
    "SNP_GENE_IDS" IS NULL AND
    "CHR_ID" IS NULL AND
    "CHR_POS" IS NULL
  );

-- Step 3: Deduplicate by rsid + trait_uri using best pvalue, then strongest effect
CREATE OR REPLACE TABLE associations_prs_ready AS
SELECT *
FROM (
  SELECT *,
         ROW_NUMBER() OVER (
           PARTITION BY rsid, trait_uri
           ORDER BY pvalue ASC, ABS(beta) DESC
         ) AS rank
  FROM associations_clean
)
WHERE rank = 1;
