# 🧬 Data Architecture Agent Brief

**Project**: SNP-to-Trait Mapping with Human-Friendly Groupings and Ontological Rigor  
**Objective**: Enable scalable, ontology-aware mapping of SNPs to user-facing trait categories using EFO and NLP-enriched annotations.

---

## 🧱 Core Tables

### `associations_clean`
- **Source**: Parsed from GWAS Catalog TSV
- **Normalization**: Comma-separated `MAPPED_TRAIT_URI` exploded to one trait per row
- **Fields**: `rsid`, `trait_uri`, `pvalue`, `beta`, gene context, study ID, etc.

### `entailed_edge`
- **Purpose**: Encodes `rdfs:subClassOf` trait hierarchy from EFO/HPO
- **Usage**: Supports recursive descent to find all trait descendants

### `trait_descendant_map`
- **Materialized View**: Recursive expansion of `entailed_edge`
- **Fields**: `descendant`, `ancestor`
- **Purpose**: Enables fast trait group lookups (e.g. SNPs under "Metabolic Health")

### `trait_to_group_map`
- **Derived**: From EFO trait labels/synonyms + human-defined groupings
- **Methods**:
  - Keyword match overlay
  - Embedding-based semantic search via `sentence-transformers`
- **Fields**: `trait_uri`, `group_name`, `match_score`

---

## 🔄 Workflow Summary

1. **Ingest GWAS TSV → `associations_clean`**
2. **Extract Trait Hierarchies → `trait_descendant_map`**
3. **Apply NLP/semantic enrichment → `trait_to_group_map`**
4. **Join for downstream outputs**:
   - `rsid → trait_uri → group_name`
   - Trait-level aggregation with hierarchy awareness

---

## ⚙️ Design Considerations

- Recursive trait hierarchy traversal cached for performance
- Embedding-based mapping enables dynamic grouping logic
- Modular layers (trait normalization, grouping enrichment) decoupled for extensibility
- Human-friendly groupings do not replace EFO—they augment it