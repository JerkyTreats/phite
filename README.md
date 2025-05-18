# 🧬 Genetic Insights: Vision & Design Guide

## 🎯 Project Vision

We are building a **personal health app** with a feature called **Genetic Insights**, which transforms a user’s raw genetic SNP data into structured, personalized, and actionable health information.

The goal is to guide users through layers of **information abstraction**, starting from individual SNPs and building up to a **Metagenetic Profile**—a high-level snapshot of health predispositions and opportunities for improvement.

---

## 🧱 Information Pyramid

### 1. 🏔️ Metagenetic Profile (Top Level)
**Purpose:** Summarize the user's overall genetic health landscape  
**Format:**
- One-paragraph summaries per major domain (e.g., Metabolic, Cognitive)
- Visual indicators (e.g., gauges, green/yellow/red risk bands)
- Links to category modules

---

### 2. 🧩 Category Modules (Middle Level)
**Example:** `Metabolic Health`  
**Purpose:** Group related conditions together (e.g., diabetes, lipid metabolism, thyroid)  
**Features:**
- Brief overview of the category
- Subsection summaries (with links to focus pages)
- Graphs of risk distribution and SNP density

---

### 3. 🔍 Focused Group Pages (Detailed Level)
**Example:** `Diabetes Risk`  
**Purpose:** Dive into a specific condition or mechanism  
**Content Structure:**
- Overview paragraph
- Key Takeaways
- Mitigation Strategies
  - Exercise
  - Nutrition
  - Supplementation
- Gene-specific table (SNP / Genotype / Health Implication)

---

### 4. 📊 Base Layer: SNP Dataset
**Format:**
- SNP ID
- Gene
- Genotype
- Reference genotype
- Risk direction (↑↓)
- Interpretation (risk/protection/neutral)

**Audience:** Data-savvy users, backend processing

---

## 🧠 Design Principles

### ✅ Progressive Disclosure
Guide users from macro insights down to micro data. Prevent overwhelm by hiding complexity until requested.

### ✅ Narrative Layering
Each level answers:
- What’s important?
- Why does it matter?
- What can I do about it?
- Where’s the proof?

### ✅ Consistent Structure
Use templated content blocks:
- `.section`, `.strategy-block`, `<table>`
- Headings + concise paragraphs
- Tables for gene summaries
- Lists for recommendations

---

## 🌟 Future-Oriented Features

### 📈 Modular Visualization
- Risk pie charts by domain
- SNP burden bar graphs
- Icons for each health category

### 🧩 Detail Toggles
- Expandable tables
- Glossary hover definitions
- Confidence indicators (low/moderate/high evidence)

### 🔗 Cross-Linking
- Category to group
- Gene → SNP explorer
- Mitigation → resources

### ⚖️ Actionability Score
- Show not just risk, but how modifiable it is
- "You carry X risk, but lifestyle changes may reduce Y%"

---

## 📁 Recommended File Structure

```
/genetic-insights/
├── index.html                 # Top-level dashboard
├── metabolic-health/
│   ├── index.html             # Metabolic Health category overview
│   ├── diabetes.html          # Focus page for diabetes risk
│   ├── lipid-profile.html     # Other focus pages
├── neurocognitive/
│   ├── memory.html
│   ├── mood.html
├── assets/
│   ├── styles.css
│   └── icons/
│       ├── brain.svg
│       ├── metabolism.svg
```

---

## 📚 Next Steps

- Prototype `Metagenetic Profile` HTML layout
- Design `Metabolic Health` category overview page
- Define core component templates (card, chart, tooltip, table)
