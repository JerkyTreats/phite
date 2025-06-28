# Statistical Correctness & Invariance Validation Implementation

## 🎯 Objective
Ensure mathematical correctness of PRS calculations and implement comprehensive invariance validation to guarantee statistical accuracy. Address critical mathematical flaws in reference statistics computation while introducing runtime validation to prevent calculation errors.

## 📐 2025 Industry-Standard PRS Methodology

### Core Mathematical Framework
| Component | Formula | Description |
|-----------|---------|-------------|
| **Additive Model** | `PRS_i = Σ_j β_j · G_{ij}` | G_{ij}: genotype dosage (0/1/2), β_j: posterior-mean shrinkage estimates |
| **Population Mean** | `μ_pop = Σ_j 2p_j β_j` | p_j: effect-allele frequency in reference population |
| **Population Variance** | `Var(PRS) = Σ_j β_j² · 2p_j(1-p_j)` | Hardy-Weinberg equilibrium assumption |
| **Normalization** | `Z_i = (PRS_i – μ_pop) / √Var(PRS)` | Continuous-ancestry PCA or most-similar-population centering |

### Effect-Size Processing
1. **Discovery GWAS**: Fit per-variant marginal statistics
2. **Shrinkage Methods**: LDpred2-auto, SBayesRC, PRS-CSx, RICE-CV ensemble
3. **Posterior Estimation**: Approximate E[β_j \| z, LD] under point-normal/continuous-shrinkage prior
4. **Scale Preservation**: Posterior β's kept on liability (binary) or natural (quantitative) scale

### Performance Benchmarks (2025)
| Trait | Ensemble Method | OR/SD | AUC (EUR) |
|-------|-----------------|-------|-----------|
| Lipids | RICE (CV + RV) | 1.60–1.85 | 0.77–0.83 |
| CAD | SBayesRC + PRS-CSx | 1.65 | 0.78 |
| T2D | LDpred2-auto (meta-GWAS) | 1.50 | 0.73 |
| Breast Cancer | PRS-CSx + functional annot. | 1.70 | 0.71 |

## 🚨 Critical Issues Identified

### Mathematical Flaws in Current Implementation
| Component | Current (WRONG) | Correct Formula |
|-----------|-----------------|-----------------|
| **Population Mean** | `Σ(2×freq×effect) / n` ✗ | `Σ_j(2×p_j×β_j)` ✓ |
| **Population Variance** | `E[expected²] - E[expected]²` ✗ | `Σ(2×freq×(1-freq)×effect²)` ✓ |
| **Reference Stats** | Sample statistics ✗ | Population parameters ✓ |
| **β Coefficients** | Raw GWAS β ✗ | Posterior-mean shrinkage estimates ✓ |

### Absolute Truth Assertions for PHITE
1. Variance equation must be used **exactly** for reference-stat calculations
2. β_j must be posterior-mean shrinkage estimates; raw GWAS betas are **disallowed**
3. PRS normalization must incorporate ancestry-matched allele frequencies
4. End-to-end validation: OR_per_SD, AUC and DOR must fall within ±3% of benchmark table

## 🏗️ Architecture Overview

### Core Components
| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `internal/invariance/` | Statistical assertions | `AssertValidProbability()`, `AssertValidVariance()`, `AssertValidDosage()`, `AssertValidBetaCoefficient()` |
| `internal/reference/stats/` | Corrected calculations | Population variance, Hardy-Weinberg validation |
| `internal/prs/` | Enhanced calculator | Pre/post-condition validation, mathematical consistency |

### Corrected Statistical Implementation
```go
// CORRECT: Population parameters
func Compute(alleleFreqs, effectSizes map[string]float64) (*ReferenceStats, error) {
    var populationMean, populationVariance float64

    for variant, freq := range alleleFreqs {
        effect := effectSizes[variant]
        populationMean += 2 * freq * effect
        populationVariance += 2 * freq * (1 - freq) * effect * effect
    }

    return &ReferenceStats{
        Mean: populationMean,
        Std:  math.Sqrt(populationVariance),
    }, nil
}
```

## 📋 Implementation Plan

### Phase 1: Mathematical Test Suite
| Step | Task | Files | Success Criteria |
|------|------|-------|------------------|
| 1.1 | Mathematical Correctness Tests | `prs_statistical_correctness_test.go` | Tests expose all mathematical flaws |
| 1.2 | Hardy-Weinberg Population Tests | `stats_mathematical_correctness_test.go` | Known theoretical values validated |
| 1.3 | PRS Calculation Validation | `prs_calculation_validation_test.go` | Individual PRS accuracy verified |
| 1.4 | Normalization Tests | `normalization_mathematical_test.go` | Z-score/percentile accuracy confirmed |

**Test Cases**:
- **Known HWE Population**: 3 SNPs, p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
  - Expected Mean: `2×(0.2×0.1 + 0.5×(-0.3) + 0.8×0.2) = 0.06`
  - Expected Variance: `2×(0.2×0.8×0.01 + 0.5×0.5×0.09 + 0.8×0.2×0.04) = 0.0610`

### Phase 2: Statistical Corrections
| Step | Task | Files | Success Criteria |
|------|------|-------|------------------|
| 2.1 | Fix Reference Statistics | `internal/reference/stats/stats.go` | Population variance formula corrected |
| 2.2 | Statistical Utilities | `statistical_utils.go` | HWE validation, numerical stability |
| 2.3 | Population Parameter Logic | Various calculation files | Sample vs population distinction clear |
| 2.4 | Validate Corrections | All test files | 100% test pass rate achieved |

### Phase 3: Invariance Package
| Step | Task | Files | Success Criteria |
|------|------|-------|------------------|
| 3.1 | Core Invariance Package | `internal/invariance/invariance.go` | Statistical assertions implemented |
| 3.2 | PRS-Specific Assertions | `internal/invariance/prs_invariance.go` | Domain-specific validations |
| 3.3 | Test Coverage | `invariance_test.go` | All assertions thoroughly tested |
| 3.4 | Performance Benchmarking | Benchmark tests | <10% overhead for 1e6 SNPs |

### Phase 4: Integration & Runtime Validation
| Step | Task | Files | Success Criteria |
|------|------|-------|------------------|
| 4.1 | PRS Calculator Integration | `internal/prs/prs_calculator.go` | Pre/post-condition assertions |
| 4.2 | Reference Stats Integration | `internal/reference/stats/stats.go` | Input/output validation |
| 4.3 | Score Normalizer Integration | `internal/prs/score_normalizer.go` | Normalization invariants |
| 4.4 | Pipeline Integration | `internal/pipeline/pipeline.go` | End-to-end validation |

### Phase 5: End-to-End Validation
| Step | Task | Files | Success Criteria |
|------|------|-------|------------------|
| 5.1 | Integration Testing | `integration_test.go` | Full pipeline statistical validation |
| 5.2 | Synthetic Data Validation | Test utilities | Known theoretical properties verified |
| 5.3 | Performance Optimization | Various files | Assertion overhead minimized |
| 5.4 | Production Configuration | Config files | Toggleable assertion levels |

## 🧪 Test Strategy

### Mathematical Correctness Tests
- **Theoretical Validation**: Known mathematical results with hand-calculated expectations
- **Synthetic Populations**: Generate HWE-compliant data, verify statistical properties
- **Edge Case Coverage**: Extreme frequencies, zero effects, single SNPs
- **Numerical Stability**: Large effect sizes, many SNPs, precision boundaries

### Invariance Validation Tests
- **Boundary Conditions**: Test all assertion boundaries (0, 1, 2 for dosage, etc.)
- **Error Path Coverage**: Verify all invariance violations properly detected
- **Performance Impact**: Benchmark assertion overhead across different input sizes
- **Integration Coverage**: End-to-end pipeline with comprehensive invariance checking

## 🎯 Success Criteria

### Mathematical Correctness
- ✅ **All Tests Pass**: 100% mathematical correctness validated
- ✅ **Theoretical Alignment**: Population statistics match Hardy-Weinberg expectations
- ✅ **Numerical Stability**: Robust handling of edge cases and precision limits
- ✅ **Reproducible Results**: Consistent calculations across different environments

### Invariance Validation
- ✅ **Comprehensive Coverage**: All critical mathematical properties validated
- ✅ **Runtime Safety**: Early detection of invalid inputs/computations
- ✅ **Clear Error Messages**: Actionable feedback for invariance violations
- ✅ **Performance Acceptable**: <10% overhead for genome-wide scores (1e6 SNPs)

### Production Readiness
- ✅ **Configurable Validation**: Toggleable assertion levels for production vs development
- ✅ **Backward Compatibility**: Existing API preserved with enhanced error handling
- ✅ **Documentation**: Clear mathematical foundations and usage examples
- ✅ **Maintainability**: Well-structured invariance checking for future extensions

## 🔬 100% Correct Tests (Math Validation Suite)

| # | Test File | Purpose | Input Fixture | Expected Output |
|---|-----------|---------|---------------|-----------------|
| 1 | `mean_variance_theory_test.go` | Closed-form population mean & variance | 3-SNP: p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2) | Mean=0.06±1e-12, Var=0.0610±1e-12 |
| 2 | `edge_case_bounds_test.go` | Invariants under extreme inputs | Single SNP, p=1, β=0.5; β=0, p=0.3; p→0 | Mean/Var analytically zero/finite, Var≥0 |
| 3 | `continuous_dosage_validation_test.go` | `AssertValidDosage` accepts 0≤d≤2 | Random dosage draws ∈[0,2]; d=-0.1,2.1,NaN | Valid pass; invalid raise invariance error |
| 4 | `posterior_shrinkage_enforcement_test.go` | Rejects raw GWAS β, accepts posterior-mean β | SNP with `BetaIsPosterior=false/true` | Expect violation/pass respectively |
| 5 | `synthetic_population_hwe_test.go` | 10,000-ind HWE population simulation | Same 3-SNP vector as Test 1 | \|empirical-theoretical\| ≤1e-3 |
| 6 | `reference_impl_crosscheck_test.go` | PHITE vs LDpred2-auto, PRS-CSx, pgsc_calc | 100×100 fixture in `testdata/` | RMSD ≤1e-6 across individuals |
| 7 | `auc_or_per_sd_benchmark_test.go` | OR_per_SD & AUC within ±3% of benchmarks | Synthetic phenotype labels + PRS scores | \|metric_PHITE-metric_ref\|/metric_ref ≤0.03 |

### Acceptance Criteria
1. `go test -race ./internal/prs/...` passes with **all tests green**
2. Line coverage on `internal/prs`, `internal/reference`, and `internal/invariance` ≥ 95%
3. Total test runtime ≤ 120s on 16-thread CI runner
4. Fixtures and golden values under `internal/prs/testdata/` stay below 2MB
