# 🧠 Trait Category Mapping – Embedding Strategy Brief

**Objective**: Enable ontology-consistent, user-relevant SNP categorization by mapping EFO trait URIs to a rich set of actionable, human-friendly trait categories using semantic embeddings.

---

## 🎯 Problem Context

- EFO trait labels are technically precise but not user-oriented.
- Consumer-facing groupings (e.g. “Heart Health”) are too coarse for precise trait mapping.
- Manual keyword maps are brittle and non-scalable.

---

## ✅ Strategy

### 1. **Develop a Semantically Rich Trait Category List**
A set of ~100 trait categories was designed to:
- Capture mechanistic biology (e.g., "folate absorption")
- Support user-facing outputs (e.g., Nutrition, Sleep, Hormones)
- Bridge the gap between EFO traits and personalized recommendations

📄 [rich_trait_category_list.md](sandbox:/mnt/data/rich_trait_category_list.md)

---

### 2. **Match EFO Traits to Categories via Embeddings**
- Use `sentence-transformers` to embed:
  - EFO trait labels (and synonyms, descriptions, parents)
  - Trait categories from the above list
- Use cosine similarity to generate best-match mappings with confidence scores
- Store in `trait_to_group_map` table with match provenance

---

### 3. **Bridge to Higher-Level Consumer Groupings**
- Optional secondary mapping: `"glucose-insulin regulation"` → `"Metabolic Health"`
- Enables flexible UX filtering while preserving trait resolution

---

## 🧱 Architecture Layer Reference

Refer to the broader system architecture:  
📄 [data_architecture_brief.md](sandbox:/mnt/data/data_architecture_brief.md)

---

## 🔄 Next Steps

- Encode and cache EFO trait embeddings
- Run matching pipeline and log top-N scores
- Validate matches above threshold (e.g. >0.85)
- Build `trait_to_group_map` table as semantic join layer