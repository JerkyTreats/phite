# BigQuery Schema Verification Scripts

This directory contains scripts to verify the schema of BigQuery tables used in the polygenic risk calculator project.

## Scripts

1. `gnomad/schema.go` - Verifies the schema of the gnomAD public dataset
2. `jerkytreats/schema.go` - Verifies the schema of the Jerkytreats PRS reference statistics table

## Prerequisites

- Go 1.21 or later
- Google Cloud SDK installed and configured
- Appropriate permissions to access:
  - The gnomAD public dataset (`bigquery-public-data.gnomAD`)
  - The Jerkytreats project and dataset (if available)

## Setup

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Ensure you have the necessary Google Cloud credentials:
   ```bash
   gcloud auth application-default login
   ```

## Usage

Run the verification scripts:

```bash
go run main.go
```

This will:
1. Verify the gnomAD schema and print:
   - Complete table schema
   - Required fields check
   - Sample data (first row)
2. Verify the Jerkytreats schema and print:
   - Complete table schema
   - Required fields check
   - Sample data (first row)

## Expected Output

### gnomAD Schema
The script will verify the presence of required fields:
- `reference_name`
- `start_position`
- `end_position`
- `reference_bases`
- `alternate_bases`

### Jerkytreats Schema
The script will verify the presence of required fields:
- `ancestry`
- `trait`
- `model_id`
- `mean_prs`
- `stddev_prs`
- `min_prs`
- `max_prs`
- `sample_size`
- `source`
- `notes`
- `last_updated`

## Notes

- The gnomAD verification should succeed as it's a public dataset
- The Jerkytreats verification may fail if:
  - The project is not set up
  - The dataset/table doesn't exist
  - The user lacks necessary permissions
