-- Step 1: Clean, zip, and deduplicate by rsid + trait_uri in one go
CREATE OR REPLACE TABLE associations_clean AS
WITH exploded AS (
  SELECT
    "SNPS" AS rsid,
    TRIM(SPLIT_PART("STRONGEST SNP-RISK ALLELE", '-', 2)) AS risk_allele,
    TRY_CAST("P-VALUE" AS DOUBLE) AS pvalue,
    TRY_CAST("OR or BETA" AS DOUBLE) AS beta,
    "STUDY ACCESSION" AS study_id,

    -- Gene mapping
    "MAPPED_GENE" AS mapped_gene,
    "UPSTREAM_GENE_ID" AS upstream_gene_id,
    "DOWNSTREAM_GENE_ID" AS downstream_gene_id,
    "SNP_GENE_IDS" AS snp_gene_ids,

    -- Genomic info
    "CHR_ID" AS chr,
    "CHR_POS" AS chr_pos,
    "CONTEXT" AS context,
    "INTERGENIC" AS is_intergenic,
    "RISK ALLELE FREQUENCY" AS risk_allele_freq,
    "95% CI (TEXT)" AS ci_95_text,

    -- Arrays of traits and URIs
    STRING_SPLIT("MAPPED_TRAIT", ',') AS trait_array,
    STRING_SPLIT("MAPPED_TRAIT_URI", ',') AS uri_array

  FROM associations_raw
  WHERE
    "SNPS" IS NOT NULL AND
    "MAPPED_TRAIT" IS NOT NULL AND
    "MAPPED_TRAIT_URI" IS NOT NULL AND
    TRY_CAST("P-VALUE" AS DOUBLE) IS NOT NULL AND
    TRY_CAST("OR or BETA" AS DOUBLE) IS NOT NULL AND
    "STRONGEST SNP-RISK ALLELE" LIKE 'rs%-%' AND
    NOT (
      "MAPPED_GENE" IS NULL AND
      "UPSTREAM_GENE_ID" IS NULL AND
      "DOWNSTREAM_GENE_ID" IS NULL AND
      "SNP_GENE_IDS" IS NULL AND
      "CHR_ID" IS NULL AND
      "CHR_POS" IS NULL
    )
),

zipped AS (
  SELECT
    rsid,
    risk_allele,
    pvalue,
    beta,
    study_id,
    mapped_gene,
    upstream_gene_id,
    downstream_gene_id,
    snp_gene_ids,
    chr,
    chr_pos,
    context,
    is_intergenic,
    risk_allele_freq,
    ci_95_text,
    TRIM(trait_array[idx]) AS trait,
    TRIM(uri_array[idx]) AS trait_uri

  FROM exploded,
       UNNEST(GENERATE_SERIES(1, LEAST(ARRAY_LENGTH(trait_array), ARRAY_LENGTH(uri_array)))) AS t(idx)
),

ranked AS (
  SELECT *,
         ROW_NUMBER() OVER (
           PARTITION BY rsid, trait_uri
           ORDER BY pvalue ASC, ABS(beta) DESC
         ) AS rank
  FROM zipped
)

SELECT *
FROM ranked
WHERE rank = 1;
