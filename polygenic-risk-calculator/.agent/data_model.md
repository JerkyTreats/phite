# Polygenic Risk Calculator: Data Model Specification

This document defines the canonical data model for all components of the polygenic risk calculator pipeline. All agent briefs and implementations must reference this as the source of truth for input/output data structures.

---

## 1. User Genotype Data
- **UserGenotype**
  - `rsid`: string  // SNP identifier
  - `genotype`: string  // e.g., "AA", "CT"

## 2. SNP Request
- **SNPRequest**
  - `rsid`: string

## 3. Validated SNP
- **ValidatedSNP**
  - `rsid`: string
  - `genotype`: string
  - `found_in_gwas`: bool

## 4. GWAS SNP Record
- **GWASSNPRecord**
  - `rsid`: string
  - `risk_allele`: string
  - `beta`: float  // effect size (OR or Beta)
  - // ...other GWAS fields as needed

## 5. Annotated SNP
- **AnnotatedSNP**
  - `rsid`: string
  - `genotype`: string
  - `risk_allele`: string
  - `beta`: float
  - `dosage`: int  // 0, 1, 2
  - `trait`: string (optional)

## 6. SNP Contribution
- **SNPContribution**
  - `rsid`: string
  - `dosage`: int
  - `beta`: float
  - `contribution`: float  // dosage × beta

## 7. PRS Result
- **PRSResult**
  - `prs_score`: float
  - `details`: list of SNPContribution

## 8. Normalized PRS
- **NormalizedPRS**
  - `raw_score`: float
  - `z_score`: float
  - `percentile`: float

## 9. Trait Summary
- **TraitSummary**
  - `trait`: string
  - `num_risk_alleles`: int
  - `effect_weighted_contribution`: float
  - `risk_level`: string  // e.g., low, moderate, high

## 10. Missing SNPs
- **snps_missing**
  - list of string (rsids)

---

## External Input Data Schemas

### GWAS Data (Tabular)
Expected columns (example):

| rsid         | chromosome | position | risk_allele | beta  | trait      | ... |
|--------------|------------|----------|-------------|-------|------------|-----|
| rs190214723  | 1          | 693625   | T           | 0.01  | height     | ... |
| rs3131972    | 1          | 752721   | G           | -0.02 | diabetes   | ... |
| rs12562034   | 1          | 768448   | G           | 0.03  | BMI        | ... |
| rs908743     | 1          | 2038824  | G           | 0.01  | height     | ... |

- Required columns: `rsid`, `chromosome`, `position`, `risk_allele`, `beta`, `trait`
- Additional columns may be present (e.g., p-value, allele frequency)

Example GWAS data can be found in `PHITE/gwas/`. Specifically, `PHITE/gwas/build_db.sh` will link to schema creation scripts for the expected input database.

### DNA/Genotype Input File (User Data)
Supported formats: AncestryDNA, 23andMe

Example (tab-delimited):

```
rsid	chromosome	position	allele1	allele2
rs190214723	1	693625	T	T
rs3131972	1	752721	G	G
rs12562034	1	768448	G	G
rs908743	1	2038824	A	G
```

- Required columns: `rsid`, `chromosome`, `position`, `allele1`, `allele2`
- Genotype is typically inferred as the concatenation of `allele1` and `allele2` (e.g., "TT", "AG")
- File must include a header row
- Malformed or missing lines should be handled gracefully

---

All components must use these definitions for their input and output objects. Any changes to the data model or input schemas must be reflected here and referenced in all agent briefs and code.
