# Polygenic Risk Calculator Pipeline - Technical Report

## Overview

The polygenic risk calculator pipeline is designed to compute polygenic risk scores (PRS) from user genotype data and GWAS summary statistics using a 4-phase bulk optimization approach. The pipeline prioritizes cost efficiency and performance by minimizing external database operations through intelligent batching and caching strategies.

## Intended Workflow for Bulk Pipeline Operations

The pipeline follows a 4-phase approach optimized for bulk operations:

1. **Phase 1: Requirements Analysis** - Pre-analyze all data needs across all traits
2. **Phase 2: Bulk Data Retrieval** - Execute minimal BigQuery operations to fetch all required data
3. **Phase 3: In-Memory Processing** - Process all traits using pre-loaded bulk data
4. **Phase 4: Bulk Storage** - Store all computed results in single bulk operations

This design minimizes external API calls and database operations by batching all requirements upfront, then processing everything in memory before performing bulk storage operations.

### Workflow Diagram

```mermaid
graph TD
    A[Input: Genotype File, SNPs, GWAS DB] --> B[Phase 1: Requirements Analysis]

    B --> B1[Fetch GWAS Records]
    B --> B2[Parse Genotype Data]
    B --> B3[Annotate SNPs with Traits]
    B --> B4[Identify All Traits]
    B --> B5[Build Cache Requests]

    B1 --> C[Phase 2: Bulk Data Retrieval]
    B2 --> C
    B3 --> C
    B4 --> C
    B5 --> C

    C --> C1[Bulk Cache Lookup]
    C --> C2[Identify Cache Misses]
    C --> C3[Load PRS Models]
    C --> C4[Bulk Allele Frequency Query]
    C --> C5[Bulk Reference Stats Computation]

    C1 --> D[Phase 3: In-Memory Processing]
    C2 --> D
    C3 --> D
    C4 --> D
    C5 --> D

    D --> D1[Calculate PRS for Each Trait]
    D --> D2[Prepare Cache Entries]
    D --> D3[Normalize PRS Scores]
    D --> D4[Generate Trait Summaries]

    D1 --> E[Phase 4: Bulk Storage]
    D2 --> E
    D3 --> E
    D4 --> E

    E --> E1[Bulk Cache Storage]
    E --> F[Output: Trait Summaries, Normalized PRS, Missing SNPs]

    %% Phase grouping
    subgraph "Phase 1: Requirements Analysis"
        B1
        B2
        B3
        B4
        B5
    end

    subgraph "Phase 2: Bulk Data Retrieval"
        C1
        C2
        C3
        C4
        C5
    end

    subgraph "Phase 3: In-Memory Processing"
        D1
        D2
        D3
        D4
    end

    subgraph "Phase 4: Bulk Storage"
        E1
    end

    %% Styling
    classDef phase1 fill:#e1f5fe
    classDef phase2 fill:#f3e5f5
    classDef phase3 fill:#e8f5e8
    classDef phase4 fill:#fff3e0
    classDef input fill:#ffebee
    classDef output fill:#e8f5e8

    class B1,B2,B3,B4,B5 phase1
    class C1,C2,C3,C4,C5 phase2
    class D1,D2,D3,D4 phase3
    class E1 phase4
    class A input
    class F output
```

## Phase 1: Requirements Analysis

Order of operations:
1. Fetch GWAS Records
2. Parse Genotype Data
3. Annotate SNPs with Traits
4. Identify All Traits
5. Build Cache Requests

### Phase 1 Call Graph

```mermaid
graph TD
    %% Root orchestrator
    A[analyzeAllRequirements]

    %% Sequence of calls
    A --> B1[ancestry.NewFromConfig]
    B1 --> B1a[ancestry.New]
    B1a --> B1b[buildInternalCode]

    A --> C1[gwas.NewGWASService]
    C1 --> C2[GWASService.FetchGWASRecords]
    C2 --> C2a[toString / toFloat64]

    A --> D1[genotype.ParseGenotypeData]
    D1 --> D1a[autoDetectFormat]
    D1 --> D1b[isValidGenotype]

    A --> E1[gwas.MapToGWASList]

    A --> F1[gwas.FetchAndAnnotateGWAS]
    F1 --> F1a[computeDosage]

    %% Data flow / dependencies
    C2 -- gwasMap --> E1
    D1 -- ValidatedSNPs --> F1
    E1 -- AssociationsClean --> F1
    F1 -- AnnotatedSNPs --> G1[buildTraitSet]
    G1 --> H1[buildCacheKeys]

    %% Styling
    classDef orchestrator fill:#ffe6e6,stroke:#333
    classDef service fill:#e6f7ff,stroke:#1f78b4
    classDef parser fill:#e6ffe6,stroke:#33a02c
    classDef helper fill:#f3f3f3,stroke:#888,stroke-dasharray: 5 5

    class A orchestrator
    class B1,B1a,B1b,C1,C2,C2a service
    class D1,D1a,D1b parser
    class E1,F1,F1a,G1,H1 service
    class B1b,C2a,D1a,D1b,F1a helper
```

### Fetch GWAS Records
- Purpose: Pre-load effect sizes, risk alleles, and trait tags for the requested rsids so later stages avoid per-SNP DB calls.
- Key function(s): `gwas.NewGWASService` → `GWASService.FetchGWASRecords` (DuckDB query).
- Inputs: slice of rsids, `gwas_db_path`, `gwas_table` from config.
- Outputs: `map[string]model.GWASSNPRecord` used downstream and converted to a slice via `MapToGWASList`.

### Parse Genotype Data
- Purpose: Read the user genotype file (AncestryDNA or 23andMe), keep rows matching the requested rsids, validate nucleotides, and flag missing ones.
- Key function(s): `genotype.ParseGenotypeData` (auto-detect format, validate, build output struct).
- Inputs: `GenotypeFilePath`, slice of rsids, GWAS record map (for cross-validation).
- Outputs: `ParseGenotypeDataOutput` containing `UserGenotypes`, `ValidatedSNPs`, and `SNPsMissing`.

### Annotate SNPs with Traits
- Purpose: Merge validated user SNPs with GWAS effect data, compute allele dosage, and attach trait identifiers for per-trait grouping.
- Key function(s): `gwas.FetchAndAnnotateGWAS` which calls `computeDosage`.
- Inputs: `ValidatedSNPs`, GWAS associations slice.
- Outputs: `AnnotatedSNPs` slice and the subset of `GWASRecords` actually used.

### Identify All Traits
- Purpose: Collect the unique set of traits present in `AnnotatedSNPs`; this drives all later per-trait loops.
- Implementation: inline set-building loop in `pipeline.analyzeAllRequirements`.
- Inputs: `AnnotatedSNPs`.
- Outputs: `map[string]struct{}` representing the trait universe for this run.

### Build Cache Requests
- Purpose: Convert the trait set into cache keys (`<ancestry>|<trait>|<model>`) so reference statistics can be fetched in one batch during Phase 2.
- Implementation: inline in `pipeline.analyzeAllRequirements` using ancestry code from config.
- Inputs: trait set, ancestry code.
- Outputs: `[]reference_cache.StatsRequest` later passed to the bulk cache lookup.

## Phase 2: Bulk Data Retrieval

### Phase 2 Call Graph

```mermaid
graph TD
    %% Root orchestrator
    A[retrieveAllDataBulk]

    %% Step 1 – Bulk cache lookup
    A --> B1[ReferenceCache.GetBatch]
    B1 --> B1a[RepositoryCache.GetBatch]

    %% Step 2 – Identify cache misses & build requests
    A --> C1[identifyCacheMisses]
    C1 --> C2[buildStatsRequests]

    %% Step 3 – Bulk reference-stats computation
    C2 --> D1[ReferenceService.GetReferenceStatsBatch]

    %% 3a – Model loading per trait
    D1 --> E1[LoadModel]
    E1 --> E1a[convertRowToVariant]
    E1a --> E1b[PRSModel.Validate]

    %% 3b – Consolidated allele-frequency query
    D1 --> F1[GetAlleleFrequenciesForTraits]
    F1 --> F1a[Ancestry.ColumnPrecedence]
    F1 --> F1b[gnomadDB.Query]
    F1b --> F1c[Ancestry.SelectFrequency]

    %% 3c – Stats computation
    D1 --> G1[reference_stats.Compute]

    %% Step 4 – Organise trait-level SNP slices
    A --> H1[partitionAnnotatedSNPs]
    H1 --> H1a[traitSNPs map]

    %% Styling (aligned with Phase 1)
    classDef orchestrator fill:#ffe6e6,stroke:#333
    classDef service fill:#e6f7ff,stroke:#1f78b4
    classDef helper fill:#f3f3f3,stroke:#888,stroke-dasharray: 5 5
    classDef db fill:#fff3e0,stroke:#ff8f00

    class A orchestrator
    class B1,B1a,D1,E1,F1 service
    class F1b db
    class C1,C2,E1a,E1b,F1a,F1c,G1,H1,H1a helper
```

Order of operations:
1. Bulk Cache Lookup
2. Identify Cache Misses
3. Load PRS Models
4. Bulk Allele Frequency Query
5. Bulk Reference Stats Computation

### Bulk Cache Lookup
- Purpose: Retrieve previously computed reference statistics for all traits in a single round-trip to the cache backend (BigQuery or DuckDB, depending on deployment).
- Key function(s): `ReferenceCache.GetBatch` invoked by `pipeline.retrieveAllDataBulk`.
- Inputs: `[]reference_cache.StatsRequest` – ancestry code, trait, model id.
- Outputs: `map[string]*reference_stats.ReferenceStats` stored in `bulkData.CachedStats`.

### Identify Cache Misses
- Purpose: Determine which traits still need reference stats after the cache lookup.
- Implementation: loop over `TraitSet`; if cache key absent, append to `cacheMisses` slice.
- Inputs: cache results from previous step, set of trait keys.
- Outputs: `[]string` of traits requiring fresh computation.

### Load PRS Models
- Purpose: For every cache-miss trait, load all SNP effect sizes in one DuckDB query per trait.
- Key function(s): `ReferenceService.LoadModel` (internally called inside `GetReferenceStatsBatch`).
- Inputs: trait name, model table (`reference.model_table` config).
- Outputs: `PRSModel` objects containing `[]model.Variant` used later for stats.

### Bulk Allele Frequency Query
- Purpose: Fetch allele frequencies for all unique variants across all cache-miss traits using a single BigQuery query, minimizing cost.
- Key function(s): `ReferenceService.GetAlleleFrequenciesForTraits`.
- Inputs: map `trait → []Variant`, ancestry object (for column precedence).
- Outputs: map `trait → variantID → frequency` consolidated in memory.

### Bulk Reference Stats Computation
- Purpose: Combine allele frequencies with model effect sizes to compute mean, std, min, and max PRS distribution parameters for each trait.
- Key function(s): `reference_stats.Compute` executed inside `ReferenceService.GetReferenceStatsBatch`.
- Inputs: allele frequency map, per-trait `PRSModel` effect sizes.
- Outputs: `map[string]*reference_stats.ReferenceStats` returned as `bulkData.ComputedStats` for downstream normalization.

## Phase 3: In-Memory Processing

### Phase 3 Call Graph

```mermaid
graph TD
    %% Root orchestrator
    A["processAllTraitsInMemory"]

    %% Per-trait loop
    A --> B["for trait in TraitSet"]

    %% PRS calculation
    B --> C1["prs.CalculatePRS"]
    C1 --> C1a["ValidateInputSNPs"]
    C1 --> C1b["loop SNPs"]
    C1b --> C1b1["ValidateVariantContribution"]
    C1 --> C1c["ValidatePRSResult"]

    %% Reference-stats resolution
    B --> D["getReferenceStats"]
    D --> D1["CachedStats hit"]
    D --> D2["ComputedStats miss"]
    D2 --> D3["appendCacheEntry"]

    %% Normalization
    B --> E["prs.NormalizePRS"]
    E --> E1["normCdf"]

    %% Trait summary
    B --> F["output.GenerateTraitSummaries"]

    %% Results aggregation
    F --> G["append TraitSummaries"]
    E --> H["store NormalizedPRS"]
    C1 --> I["store PRSResults"]

    %% Styling
    classDef orchestrator fill:#ffe6e6,stroke:#333
    classDef loop fill:#f3f3f3,stroke:#888,stroke-dasharray: 5 5
    classDef calc fill:#e6f7ff,stroke:#1f78b4
    classDef valid fill:#e6ffe6,stroke:#33a02c
    classDef retr fill:#e8f5e8,stroke:#2e7d32
    classDef db fill:#fff3e0,stroke:#ff8f00
    classDef helper fill:#ede7f6,stroke:#5e35b1
    classDef io fill:#fce4ec,stroke:#ad1457
    classDef out fill:#d1c4e9,stroke:#512da8

    class A orchestrator
    class B,C1b loop
    class C1,E calc
    class C1a,C1b1,C1c valid
    class D retr
    class D1,D2 db
    class E1 helper
    class D3,G,H,I io
    class F out
```

Order of operations:
1. Calculate PRS for Each Trait
2. Prepare Cache Entries
3. Normalize PRS Scores
4. Generate Trait Summaries

### Calculate PRS for Each Trait
- Purpose: Aggregate per-trait SNP contributions to produce an unnormalized polygenic risk score.
- Key function(s): `prs.CalculatePRS` (iterates over `[]model.AnnotatedSNP`, multiplies dosage by beta, returns `PRSResult`).
- Inputs: slice of annotated SNPs for a single trait.
- Outputs: `PRSResult` with total score and per-SNP contribution breakdown.

### Prepare Cache Entries
- Purpose: For traits whose reference stats were newly computed (cache miss), stage them for bulk cache write in Phase 4.
- Implementation: When stats come from `bulkData.ComputedStats`, build `reference_cache.CacheEntry` and append to `cacheEntries`.
- Inputs: `ComputedStats` map, ancestry code.
- Outputs: `[]reference_cache.CacheEntry` carried forward to bulk storage.

### Normalize PRS Scores
- Purpose: Convert raw PRS into an interpretable percentile using ancestry-/trait-specific reference parameters.
- Key function(s): `prs.NormalizePRS`, which wraps `reference_stats.NormalizePRS` logic.
- Inputs: `PRSResult` from previous step, `model.ReferenceStats` fetched/computed in Phase 2.
- Outputs: `prs.NormalizedPRS` (raw score, z-score, percentile) stored in `NormalizedPRS` map.

### Generate Trait Summaries
- Purpose: Produce a human-readable summary per trait (risk level, allele counts, effect-weighted contribution).
- Key function(s): `output.GenerateTraitSummaries`.
- Inputs: trait's annotated SNPs, its normalized PRS percentile.
- Outputs: `[]output.TraitSummary` appended to `TraitSummaries` slice.

## Phase 4: Bulk Storage

Order of operations:
1. Bulk Cache Storage

### Phase 4 Call Graph

```mermaid
graph TD
    %% Root orchestrator
    A[pipeline.storeBulkResults]:::orchestrator

    %% External cache abstraction
    A --> B[ReferenceCache.StoreBatch]:::cache

    %% Concrete cache layer
    B --> C[RepositoryCache.StoreBatch]:::cache

    %% Internal batching loop & validation
    C --> D[RepositoryCache.storeBatch]:::loop
    D --> E[ReferenceStats.Validate]:::validator

    %% Final DB write
    D --> F[db.Repository.Insert]:::db

    %% Styling
    classDef orchestrator fill:#ffe6e6,stroke:#333
    classDef cache fill:#e6f7ff,stroke:#1f78b4
    classDef loop fill:#f3f3f3,stroke:#888,stroke-dasharray: 5 5
    classDef validator fill:#ede7f6,stroke:#5e35b1
    classDef db fill:#fff3e0,stroke:#ff8f00

    class A orchestrator
    class B,C cache
    class D loop
    class E validator
    class F db
```

### Bulk Cache Storage
- Purpose: Persist any newly computed reference statistics to the cache in a single batched insert so future runs can skip Phase 2 computations.
- Key function(s): `pipeline.storeBulkResults` → `ReferenceCache.StoreBatch`.
- Inputs: `[]reference_cache.CacheEntry` (ancestry, trait, modelID, stats), gathered during Phase 3.
- Outputs: none (side-effect is rows written/updated in the cache backend); returns error if write fails.

## Output

The pipeline produces:
- Trait Summaries
- Normalized PRS
- Missing SNPs
