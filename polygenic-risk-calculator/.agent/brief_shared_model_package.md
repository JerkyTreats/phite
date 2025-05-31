# Brief: Shared Model Package Refactor (`internal/model/`)

## Purpose

Unify all canonical data models used throughout the polygenic-risk-calculator pipeline into a single shared package (`internal/model`). This eliminates redundant struct definitions and conversion helpers, ensuring consistency, maintainability, and clarity across all agents and pipeline stages.

## Motivation

Currently, nearly identical structs (e.g., `GWASSNPRecord`, `ValidatedSNP`, `AnnotatedSNP`, `ReferenceStats`) are defined in multiple internal packages, requiring error-prone conversion functions and increasing maintenance burden. Centralizing these types in a shared `model` package will:

- Remove the need for conversion helpers in the entrypoint and other packages.
- Ensure all agents operate on a single, canonical representation of each domain model.
- Simplify future enhancements, documentation, and onboarding.

---

## Scope of Change

### 1. **Create `internal/model/model.go`**
Define the following canonical structs (fields shown for clarity):

- **GWASSNPRecord**
  ```go
  type GWASSNPRecord struct {
      RSID       string
      RiskAllele string
      Beta       float64
      Trait      string // optional
  }
  ```
- **ValidatedSNP**
  ```go
  type ValidatedSNP struct {
      RSID        string
      Genotype    string
      FoundInGWAS bool
  }
  ```
- **AnnotatedSNP**
  ```go
  type AnnotatedSNP struct {
      RSID       string
      Genotype   string
      RiskAllele string
      Beta       float64
      Dosage     int
      Trait      string // optional
  }
  ```
- **ReferenceStats**
  ```go
  type ReferenceStats struct {
      Mean     float64
      Std      float64
      Min      float64
      Max      float64
      Ancestry string
      Trait    string
      Model    string
  }
  ```
- **UserGenotype**
  ```go
  type UserGenotype struct {
      RSID     string
      Genotype string
  }
  ```
- **ParseGenotypeDataInput / Output**  
  (if used across packages, otherwise keep local)

- **PRSResult, SNPContribution, NormalizedPRS**  
  (if used across packages, otherwise keep local)

---

### 2. **Update All Affected Files**
Replace package-local struct definitions and type references with imports from `internal/model`. Remove any now-unnecessary conversion helpers.

#### **Files to Update:**
- `internal/gwas/gwas_data_fetcher.go`
- `internal/gwas/gwas_duckdb_loader.go`
- `internal/genotype/genotype_parser.go`
- `internal/prs/prs_calculator.go`
- `internal/prs/score_normalizer.go`
- `internal/reference/reference_stats_loader.go`
- `internal/output/output_formatter.go`
- `internal/output/trait_summary_generator.go`
- `cmd/risk-calculator/main.go`

#### **Test Files to Update:**
- `internal/gwas/gwas_data_fetcher_test.go`
- `internal/gwas/gwas_duckdb_loader_test.go`
- `internal/genotype/genotype_parser_test.go`
- `internal/prs/prs_calculator_test.go`
- `internal/prs/score_normalizer_test.go`
- `internal/reference/reference_stats_loader_test.go`
- `internal/output/output_formatter_test.go`
- `internal/output/trait_summary_generator_test.go`
- Any other test files referencing affected structs

---

### 3. **Remove Conversion Helpers**
- Delete all conversion functions between similar structs, especially from `cmd/risk-calculator/main.go`.

---

### 4. **Update Documentation**
- Update GoDoc comments to reference the shared models.
- Update `.agent/README.md` and any pipeline documentation to clarify that all canonical types are defined in `internal/model`.

---

## Acceptance Criteria

- All canonical data models are defined only in `internal/model/model.go`.
- All agents and pipeline stages use the shared types directly.
- No conversion helpers remain for these types.
- All tests and builds pass.
- Documentation reflects the new structure.

---

## Constraints

- Maintain backwards compatibility with current CLI and agent APIs.
- Follow idiomatic Go practices for package structure and visibility.
- Ensure all exported types and fields are documented.

---

**This refactor will significantly improve code health and reduce future maintenance effort.**
