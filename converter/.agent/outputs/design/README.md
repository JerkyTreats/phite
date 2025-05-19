# Converter

The converter tool is a prototype tool for the PHITE Genetic Insights feature.

The input will be a .tsv file with the following format:

```tsv
Topic	Group	Gene	RS ID	Allele 	Subject Genotype	Notes
```

The output will be a simple html output files, reference `genetic-reports/diabetes`. In this example, `diabetes` is a `Group` in the .tsv file.

## TSV File Relation

* `Gene`, `RS ID`, `Allele`, `Notes` belong to an `SNP` type.
* Groupings have a set of `SNP` types. Groupings have a `Topic`, and a `Name`.

## Conversion 1: SNP sets as JSON

This converter takes the TSV file and converts it into a JSON file. The JSON files will have a set of Groupings:

```json
{
    "Groupings": [
        {
            "Topic" : "Nutrients – Vitamins and Minerals",
            "Name" : "MTHFR",
            "SNP" : [
                {
                    "Gene" : "MTHFR C677T",
                    "RSID" : "rs1801133",
                    "Allele" : "A",
                    "Notes" : "40-70% decrease in MTHFR enzyme function (folate metabolism)",
                    "Subject" : {
                        "Genotype" : "AG",
                        "Match" : "Partial"
                    }
                },
                {
                    "Gene" : "MTHFR A1298C",
                    "RSID" : "rs1801131",
                    "Allele" : "G",
                    "Notes" : "10-20% decrease in MTHFR enzyme function (folate metabolism)",
                    "Subject" : {
                        "Genotype" : "TT",
                        "Match" : "None"
                    }
                },
            ]
        }
    ]
}
```

* The script converts the input TSV and forms this JSON file.
* Topic and Name are fairly simple to map, as should be creating the SNP objects.
* The `Subject` type maps `Subject Genotype` to `Allele` to generate the `Match`
  * Match is one of [`None`, `Partial`, `Full`]
  * Subject Genotype can be blank or `--`. In this case the SNP is removed from the set.

---

## 🔁 Conversion 2: LLM Inference of SNPs

This conversion step takes the structured JSON input and runs **LLM inference for each Grouping**.

The output is saved using a **hybrid format** combining Markdown (for readable content) and JSON (for structured metadata).

### 📦 Output Format (Hybrid: JSON Metadata + Markdown Sections)

Each Grouping will be stored in a directory with the following structure:

```
/groupings/{grouping_id}/
├── meta.json
├── overview.md
├── key-takeaways.md
├── mitigation.json
├── gene-insights.json
```

---

### 📄 `meta.json`

Contains identifiers and metadata for the grouping.

```json
{
  "id": "diabetes",
  "title": "Genetic Insights: Diabetes Risk",
  "category": "Metabolic Health",
  "tags": ["insulin", "glucose", "resistance"],
  "last_updated": "2025-05-17"
}
```

---

### 📄 `overview.md`

Markdown-formatted summary of the genetic risk and condition context.

---

### 📄 `key-takeaways.md`

Markdown bullet-point list of user-relevant highlights.

---

### 📄 `mitigation.json` (Flexible, Evidence-Aware Structure)

Mitigation strategies are now structured by `type` with support for optionality and rationale:

```json
[
  {
    "type": "exercise",
    "applies": true,
    "recommendations": [
      "Engage in moderate aerobic activity at least 150 minutes per week to enhance insulin sensitivity."
    ],
    "rationale": "Shown to benefit glucose metabolism in individuals with insulin resistance SNPs like IRS1/rs2943641.",
    "confidence": "high"
  },
  {
    "type": "nutrition",
    "applies": false,
    "recommendations": [],
    "rationale": "No significant nutritional interventions are known for this SNP.",
    "confidence": "low"
  },
  {
    "type": "pharmacological",
    "applies": true,
    "recommendations": [
      "Metformin may be considered for early intervention under clinical supervision."
    ],
    "rationale": "Supported by clinical evidence for certain genotypes with impaired glucose uptake.",
    "confidence": "moderate"
  }
]
```

This approach allows the LLM to omit strategies that don’t apply and explain why—improving transparency and tailoring.

---

### 📄 `gene-insights.json`

Structured SNP interpretations for each gene:

```json
[
  {
    "Gene": "SLC30A8",
    "SNPs": [
      {
        "RSID": "rs13266634",
        "Allele": "TC",
        "Match": "Partial match",
        "Implication": "One protective allele for T2D; supports zinc-based insulin storage"
      }
    ]
  }
]
```

