# PHITE Risk Scoring Data Model Specification

## Overview
This document defines the canonical data structures, column names, and relationships required for all modules in the PHITE risk scoring pipeline. All feature implementations should conform to these specifications for interoperability and data integrity.

---

## 1. User Genotype DataFrame
- **Columns:**
  - `rsid` (str): SNP identifier (e.g., 'rs123456')
  - `genotype` (str or int): User genotype (e.g., 'AA', 'AG', or dosage 0/1/2)
- **Notes:**
  - Must be parsed from `.txt` or `.csv` user upload (AncestryDNA/23andMe format)
  - Used to join with GWAS associations on `rsid`

---

## 2. GWAS Association DataFrame (`associations_clean`)
- **Columns:**
  - `rsid`: SNP identifier
  - `risk_allele`: Effect/risk allele (e.g., 'A')
  - `pvalue`: Association p-value (float)
  - `beta`: Effect size (float, OR or BETA)
  - `trait`: Trait label (e.g., 'Type 2 diabetes')
  - `trait_uri`: Ontology URI for trait (e.g., 'http://www.ebi.ac.uk/efo/EFO_0001360')
  - `study_id`: Study accession
  - `mapped_gene`, `upstream_gene_id`, `downstream_gene_id`, `snp_gene_ids`: Gene annotations
  - `chr`, `chr_pos`: Chromosome and position
  - `context`, `is_intergenic`, `risk_allele_freq`, `ci_95_text`: Additional metadata
- **Notes:**
  - Each row links a SNP to a trait and study, possibly multiple traits per SNP

---

## 3. Trait Ontology Mapping DataFrames
- **`trait_concepts`**
  - `trait_uri`, `efo_term`, `parent_term`, `parent_uri`
- **`trait_labels`**
  - `trait_label`, `trait_uri`
- **`trait_ontology_map`**
  - `trait`, `trait_uri`, `topic`, `topic_uri`
- **Notes:**
  - Used to map traits/SNPs to higher-level ontology topics (disease clusters)

---

## 4. Polygenic Score Summary DataFrame
- **Columns:**
  - `topic`: Ontology topic/cluster name
  - `topic_uri`: Ontology URI for topic
  - `PRS`: Polygenic risk score (float)
  - (Optional) `ci_95_text`: Confidence interval
  - (Optional) `contributing_snps`: List of contributing SNPs

---

## 5. Study Metadata DataFrame (`studies_clean`)
- **Columns:**
  - `study_id`, `pubmed_id`, `author`, `pub_date`, `journal`, `initial_sample_size`,
    `replication_sample_size`, `platform`, `genotyping_tech`, `cohort`,
    `has_summary_stats`, `summary_url`
- **Notes:**
  - Used for provenance and reporting

---

## General Guidance
- All modules must validate input DataFrames for presence and correct typing of required columns.
- If additional columns are needed, document and justify them in the implementation.
- For further details or example schemas, review the SQL scripts in `gwas/sql/`.
