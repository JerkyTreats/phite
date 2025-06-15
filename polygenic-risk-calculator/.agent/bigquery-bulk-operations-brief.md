# BigQuery Bulk Operations Optimization Brief

## Project: Polygenic Risk Calculator Cost Optimization
**Issue**: BigQuery pay-per-query costs optimization through bulk operations
**Estimated Cost Reduction**: 40-70% in BigQuery costs

---

## Problem Statement

The current polygenic-risk-calculator implementation performs individual BigQuery operations for:
- Allele frequency lookups per variant set
- Reference statistics retrieval per trait
- Cache operations per entry
- Model loading per request

This results in high BigQuery costs due to multiple small queries instead of batched operations.

---

## Current Architecture Analysis

### Cost Drivers Identified

1. **Reference Service Operations** (`internal/reference/service.go`):
   - `LoadModel()`: Individual model queries
   - `GetAlleleFrequencies()`: OR-based queries per variant batch
   - `GetReferenceStats()`: Individual cache lookups

2. **Pipeline Processing** (`internal/pipeline/pipeline.go`):
   - Per-trait sequential processing
   - Individual BigQuery calls per trait

3. **Cache Operations** (`internal/reference/cache/cache.go`):
   - Single-row INSERT operations
   - Individual cache lookups

---

## Implementation Phases

### Phase 1: Cache Bulk Operations

#### Changes Required:

1. **Modify Cache Interface** (`internal/reference/cache/cache.go`):
```go
type Cache interface {
    Get(ctx context.Context, req StatsRequest) (*reference_stats.ReferenceStats, error)
    GetBatch(ctx context.Context, reqs []StatsRequest) (map[string]*reference_stats.ReferenceStats, error)
    Store(ctx context.Context, req StatsRequest, stats *reference_stats.ReferenceStats) error
    StoreBatch(ctx context.Context, entries []CacheEntry) error
}
```

2. **Implement Batch Methods**:
   - `StoreBatch()`: Accumulate cache entries and insert in single BigQuery operation
   - `GetBatch()`: Single query with IN clause for multiple cache lookups

3. **Pipeline Integration**:
   - Collect all cache operations during trait processing
   - Execute bulk cache storage at end of pipeline

#### Files to Modify:
- `internal/reference/cache/cache.go`
- `internal/reference/cache/cache_test.go`
- `internal/pipeline/pipeline.go`

---

### Phase 2: Variant Query Optimization

#### Changes Required:

1. **Cross-Trait Variant Batching** (`internal/reference/service.go`):
```go
func (s *ReferenceService) GetAlleleFrequenciesForTraits(
    ctx context.Context,
    traitVariants map[string][]model.Variant,
    ancestry *ancestry.Ancestry,
) (map[string]map[string]float64, error)
```

2. **Consolidated Variant Query**:
   - Collect all variants across all traits
   - Single BigQuery query for all variants
   - Partition results by trait in memory

3. **Pipeline Restructuring**:
   - Pre-collect all variant requirements
   - Single bulk variant frequency query
   - Process results for each trait

#### Files to Modify:
- `internal/reference/service.go`
- `internal/reference/service_test.go`
- `internal/pipeline/pipeline.go`

---

### Phase 3: Reference Stats Batching

#### Changes Required:

1. **Batch Reference Stats Computation**:
```go
func (s *ReferenceService) GetReferenceStatsBatch(
    ctx context.Context,
    requests []ReferenceStatsRequest,
) (map[string]*reference_stats.ReferenceStats, error)
```

2. **Pre-Collection Strategy**:
   - Identify all required (ancestry, trait, model) combinations upfront
   - Batch cache lookups
   - Batch computation for cache misses
   - Bulk cache storage

#### Files to Modify:
- `internal/reference/service.go`
- `internal/pipeline/pipeline.go`

---

### Phase 4: Pipeline Architecture Optimization

#### Changes Required:

1. **Batch-Aware Pipeline**:
```go
func RunOptimized(input PipelineInput) (PipelineOutput, error) {
    // 1. Analysis phase: collect all requirements
    // 2. Bulk data retrieval phase
    // 3. In-memory processing phase
    // 4. Bulk storage phase
}
```

2. **Requirements Collection**:
   - Pre-analyze all traits and variants
   - Determine all BigQuery operations needed
   - Execute in minimal number of bulk operations

---

## Technical Specifications

### New Data Structures

```go
// Batch cache entry
type CacheEntry struct {
    Request StatsRequest
    Stats   *reference_stats.ReferenceStats
}

// Batch reference stats request
type ReferenceStatsRequest struct {
    Ancestry *ancestry.Ancestry
    Trait    string
    ModelID  string
}

// Bulk operation context
type BulkOperationContext struct {
    VariantRequests map[string][]model.Variant // trait -> variants
    StatsRequests   []ReferenceStatsRequest
    CacheEntries    []CacheEntry
}
```

### Configuration Changes

Add bulk operation settings to config:
```json
{
  "bigquery": {
    "bulk_operations": {
      "cache_batch_size": 100,
      "variant_batch_size": 1000,
      "enable_batching": true
    }
  }
}
```

---

## Testing Strategy

### Unit Tests
- Mock bulk repository operations
- Test batch size limits
- Test partial failure scenarios
- Test backward compatibility

---

## Implementation Plan

### Implementation Status Table

| Phase | Task | Description | Files to Modify | Status | Priority | Estimated Effort |
|-------|------|-------------|------------------|--------|----------|------------------|
| **Phase 1: Cache Bulk Operations** | | | | | | |
| 1.1 | Design Batch Cache Interface | Add `GetBatch()` and `StoreBatch()` methods to Cache interface | `internal/reference/cache/cache.go` | DONE | High | 2h |
| 1.2 | Implement StoreBatch Method | Bulk INSERT operations for cache entries | `internal/reference/cache/cache.go` | DONE | High | 3h |
| 1.3 | Implement GetBatch Method | Single query with IN clause for multiple cache lookups | `internal/reference/cache/cache.go` | DONE | High | 3h |
| 1.4 | Add Batch Configuration | Add bulk operation settings to config system | `internal/config/config.go` | DONE | Medium | 1h |
| 1.5 | Update Pipeline Cache Usage | Collect cache operations and execute in bulk | `internal/pipeline/pipeline.go` | DONE | High | 4h |
| 1.6 | Write Cache Batch Tests | Unit tests for batch cache operations | `internal/reference/cache/cache_test.go` | DONE | Medium | 3h |
| **Phase 2: Variant Query Optimization** | | | | | | |
| 2.1 | Design Cross-Trait Batching | Create `GetAlleleFrequenciesForTraits()` method | `internal/reference/service.go` | DONE | High | 2h |
| 2.2 | Implement Consolidated Queries | Single BigQuery query for all variants across traits | `internal/reference/service.go` | DONE | High | 4h |
| 2.3 | Add Result Partitioning | Partition frequency results by trait in memory | `internal/reference/service.go` | DONE | High | 2h |
| 2.4 | Update Pipeline Variant Processing | Use bulk variant queries in pipeline | `internal/pipeline/pipeline.go` | DONE | High | 3h |
| 2.5 | Write Variant Batch Tests | Unit tests for cross-trait variant batching | `internal/reference/service_test.go` | DONE | Medium | 3h |
| 2.6 | Ancestry Validation Testing | Test with multiple ancestry combinations | Test files | DONE | Medium | 2h |
| **Phase 3: Reference Stats Batching** | | | | | | |
| 3.1 | Design Batch Stats Interface | Add `GetReferenceStatsBatch()` method | `internal/reference/service.go` | DONE | High | 2h |
| 3.2 | Implement Requirements Collection | Pre-collect all (ancestry, trait, model) combinations | `internal/reference/service.go` | DONE | High | 3h |
| 3.3 | Add Batch Cache Integration | Integrate with batch cache operations | `internal/reference/service.go` | DONE | High | 3h |
| 3.4 | Implement Bulk Computation | Batch computation for cache misses | `internal/reference/service.go` | DONE | High | 4h |
| 3.5 | Update Pipeline Stats Usage | Use batch reference stats in pipeline | `internal/pipeline/pipeline.go` | DONE | High | 2h |
| 3.6 | Write Stats Batch Tests | Unit tests for batch reference stats | `internal/reference/service_test.go` | DONE | Medium | 3h |
| 3.7 | Validate Computation Accuracy | Ensure batch results match individual results | Test files | DONE | High | 2h |
| **Phase 4: Pipeline Architecture Optimization** | | | | | | |
| 4.1 | Design Optimized Pipeline Flow | Create batch-aware pipeline architecture | `internal/pipeline/pipeline.go` | TODO | High | 3h |
| 4.2 | Implement Requirements Analysis | Pre-analyze all traits and data requirements | `internal/pipeline/pipeline.go` | TODO | High | 4h |
| 4.3 | Add Bulk Data Retrieval Phase | Execute minimal BigQuery operations upfront | `internal/pipeline/pipeline.go` | TODO | High | 4h |
| 4.4 | Implement In-Memory Processing | Process all traits using cached data | `internal/pipeline/pipeline.go` | TODO | High | 3h |
| 4.5 | Add Bulk Storage Phase | Store all results in single bulk operation | `internal/pipeline/pipeline.go` | TODO | High | 2h |
| 4.6 | Create Bulk Operation Context | Data structures for bulk operation management | `internal/pipeline/pipeline.go` | TODO | Medium | 2h |
| **Configuration & Infrastructure** | | | | | | |
| C.1 | Add Bulk Config Schema | Define bulk operation configuration structure | Config files | TODO | Medium | 1h |
| C.2 | Add Batch Size Settings | Configurable batch sizes for different operations | Config files | TODO | Medium | 1h |
| C.3 | Add Feature Flags | Enable/disable bulk operations for rollback | Config files | TODO | Medium | 1h |
| **Testing & Validation** | | | | | | |
| T.1 | Mock Bulk Operations | Create mocks for bulk BigQuery operations | Test files | TODO | Medium | 2h |
| T.2 | Partial Failure Testing | Test handling of partial batch failures | Test files | TODO | High | 3h |
| T.3 | Backward Compatibility Testing | Ensure non-bulk operations still work | Test files | TODO | High | 2h |
| T.4 | Load Testing | Test with large datasets and many traits | Test files | TODO | Medium | 3h |
