# Configuration Architecture Analysis: Centralized vs Domain-Specific

## Current Problem
5 different keys for GCP projects across domains:
- `db.project_id`, `bq_project`, `bq_billing_project`, `user.gcp_project`, `cache.gcp_project`

## Approach 1: Centralized Configuration (What I just implemented)

### ✅ Pros
- **Single source of truth** - All config keys in one place
- **No duplication** - Prevents the 5-way GCP project split
- **Consistent naming** - Enforced conventions across all domains
- **Easy system overview** - Can see all dependencies at a glance
- **Better validation** - Centralized config validation logic
- **Documentation friendly** - Auto-generate complete config schema

### ❌ Cons
- **God object antipattern** - `config.go` knows about all domains
- **Tight coupling** - Config package now depends on every domain concept
- **Violates domain boundaries** - Config shouldn't know about ancestry, cache, etc.
- **Harder modularization** - Can't extract domains without central config
- **Single responsibility violation** - Config does too many things
- **Testing complexity** - Every domain change affects central config
- **Change amplification** - New domains require touching central config

## Approach 2: Domain-Specific Configuration (Current)

### ✅ Pros
- **Domain ownership** - Each module owns its configuration
- **Loose coupling** - Domains are self-contained and testable
- **Single responsibility** - Each domain defines only what it needs
- **Easy modularization** - Can extract domains independently
- **Domain expertise** - Domain experts define their own config needs
- **Change isolation** - New domains don't affect existing configuration

### ❌ Cons
- **Duplication** - The 5-way GCP project problem we have
- **Inconsistent naming** - `ancestry.population` vs `bq_project` style mixing
- **Hidden dependencies** - Hard to see full system config requirements
- **Potential conflicts** - Risk of key name collisions
- **Distributed validation** - No central place to validate config coherence

## Approach 3: Hybrid - Domain Configuration with Shared Infrastructure

### Architecture
```go
// Shared infrastructure concepts
package config

const (
    // Common infrastructure patterns - not domain specific
    GCPDataProjectKey    = "gcp.data_project"
    GCPBillingProjectKey = "gcp.billing_project"
    GCPCacheProjectKey   = "gcp.cache_project"
)

// Domain-specific configuration
package ancestry

const (
    PopulationKey = "ancestry.population"
    GenderKey     = "ancestry.gender"
)

// Each domain uses shared infrastructure
func init() {
    config.RegisterRequiredKey(config.GCPBillingProjectKey)
    config.RegisterRequiredKey(PopulationKey)
}
```

### ✅ Pros
- **Eliminates duplication** - Shared infrastructure concepts are centralized
- **Maintains domain boundaries** - Domains still own their specific config
- **Clear separation** - Infrastructure vs domain configuration
- **Prevents conflicts** - Shared keys are defined once
- **Easier testing** - Domains can still be tested independently
- **Better documentation** - Can generate both infrastructure and domain config docs

### ❌ Cons
- **Complexity** - Two-tier configuration system
- **Coordination needed** - Changes to shared config affect multiple domains
- **Still some coupling** - Domains depend on shared config constants

## Approach 4: Configuration Composition with Interfaces

### Architecture
```go
// Define configuration interfaces for common patterns
type GCPProjectProvider interface {
    GetDataProject() string
    GetBillingProject() string
    GetCacheProject() string
}

// Each domain implements what it needs
type AncestryConfig struct {
    Population string `json:"ancestry.population"`
    Gender     string `json:"ancestry.gender"`
}

type GCPConfig struct {
    DataProject    string `json:"gcp.data_project"`
    BillingProject string `json:"gcp.billing_project"`
    CacheProject   string `json:"gcp.cache_project"`
}

// Compose configurations at runtime
type RuntimeConfig struct {
    GCP      GCPConfig      `json:"gcp"`
    Ancestry AncestryConfig `json:"ancestry"`
}
```

## Recommendation: Hybrid Approach (Approach 3)

### Why This Works Best

1. **Solves the duplication problem** - GCP projects defined once
2. **Maintains domain ownership** - Each domain still controls its specific needs
3. **Clear boundaries** - Infrastructure vs domain separation
4. **Practical implementation** - Minimal changes to existing codebase

### Implementation Strategy

#### Phase 1: Extract Shared Infrastructure
```go
// config/infrastructure.go
const (
    // Infrastructure - used across multiple domains
    GCPDataProjectKey    = "gcp.data_project"
    GCPBillingProjectKey = "gcp.billing_project"
    GCPCacheProjectKey   = "gcp.cache_project"

    BigQueryGnomadDatasetKey = "bigquery.gnomad_dataset"
    BigQueryCacheDatasetKey  = "bigquery.cache_dataset"
)
```

#### Phase 2: Keep Domain-Specific Constants in Domains
```go
// ancestry/config.go
const (
    PopulationKey = "ancestry.population"
    GenderKey     = "ancestry.gender"
)

// invariance/config.go
const (
    EnableValidationKey = "invariance.enable_validation"
    StrictModeKey       = "invariance.strict_mode"
)
```

#### Phase 3: Domains Import Shared Infrastructure
```go
// ancestry/config.go
func init() {
    config.RegisterRequiredKey(PopulationKey)
    // Uses shared infrastructure for billing
    config.RegisterRequiredKey(config.GCPBillingProjectKey)
}
```

## Benefits of Hybrid Approach

1. **🎯 Addresses root cause** - Eliminates GCP project duplication
2. **🏗️ Maintains architecture** - Domains still own their configuration
3. **📚 Clear documentation** - Infrastructure vs domain config separation
4. **🔧 Easier maintenance** - Shared infrastructure changes propagate automatically
5. **🧪 Testable** - Domains can still be tested independently
6. **📈 Scalable** - New domains only need to import relevant shared infrastructure

## Final Configuration Schema

```json
{
  "gcp": {
    "data_project": "bigquery-public-data",
    "billing_project": "my-billing-project",
    "cache_project": "my-cache-project"
  },
  "bigquery": {
    "gnomad_dataset": "gnomad",
    "cache_dataset": "reference_cache"
  },
  "ancestry": {
    "population": "EUR",
    "gender": ""
  },
  "invariance": {
    "enable_validation": true,
    "strict_mode": false
  },
  "cache": {
    "batch_size": 100
  },
  "logging": {
    "level": "INFO"
  }
}
```

This approach gives us the best of both worlds: eliminates duplication while preserving domain boundaries.
