---
mode: 'agent'
description: 'Analyze SNP data and provide health insights based on PHITE methodology'
---
# PHITE SNP Analysis

Your task is to analyze SNP (Single Nucleotide Polymorphism) data and provide health insights according to the PHITE methodology.

## Instructions

1. First, identify which SNPs are present in the provided dataset
2. Cross-reference these SNPs with known associations in the GWAS catalog
3. Organize findings by:
   - Health category (e.g., cardiovascular, metabolism, neurological)
   - Level of evidence (strong, moderate, preliminary)
   - Actionability (lifestyle modifications, further testing, medical consultation)

## Required Structure

For each identified health insight, provide:
- **What:** Clear statement of the genetic finding
- **Why:** Explanation of relevance to health
- **Action:** Specific, actionable recommendations
- **Evidence:** Citations to scientific literature

## Technical References
- Reference the [PHITE risk-scoring documentation](../../risk-scoring/README.md) for methodology
- Use the GWAS database for evidence weighting
- Follow the progressive disclosure principle: most important insights first

Remember to maintain a professional, evidence-based tone and highlight when findings are preliminary vs. well-established.
