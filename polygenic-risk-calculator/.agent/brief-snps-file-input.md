# Brief: Support SNP List Input via File (JSON and CSV)

## Objective
Enable the polygenic-risk-calculator CLI to accept SNP rsids from a file, supporting both CSV and JSON formats, in addition to the existing comma-separated `--snps` flag. This improves usability for large SNP lists and integration with external tools.

## Requirements
- Add a new CLI flag: `--snps-file` (string). Mutually exclusive with `--snps`.
- Supported formats:
  - **CSV**: File contains either:
    - One rsid per line (no header), **or**
    - A header row with a column labeled `rsid` (case-insensitive), in any position; rsids are extracted from that column in all subsequent rows. Other columns are ignored.
      - Example:

        ```csv
        effect,rsid,pval
        0.12,rs123,0.01
        0.08,rs456,0.02
        ```

      - The parser must extract `rs123` and `rs456` from the `rsid` column regardless of its position.
  - **JSON**: File is a JSON array of strings, e.g. `["rs123", "rs456"]`.
- If both `--snps` and `--snps-file` are provided, return an error and usage message.
- If neither is provided, return an error and usage message.
- Parse the provided file according to its extension (`.csv` or `.json`).
- Validate all parsed rsids (non-empty, unique, trimmed).
- Use the resulting rsid list for downstream processing (GWAS, genotype, etc).
- Update CLI help and documentation to describe the new flag and input formats.

## Acceptance Criteria
- CLI accepts SNPs from either `--snps` or `--snps-file` (not both).
- Both JSON and CSV file formats are supported and tested.
- Clear error messages for invalid input or usage.
- All downstream logic works identically regardless of input method.
- Tests cover file parsing and error cases.

## Test Requirements
- **Unit tests** for SNP file parsing logic:
  - Parse valid JSON array of strings → correct rsid slice
  - Parse valid CSV (one rsid per line) → correct rsid slice
  - Parse valid CSV with header (column labeled `rsid`, any position) → correct rsid slice
  - Parse valid multi-column CSV, extract rsids from the correct column
  - Ignore blank lines and whitespace
  - Reject malformed JSON, malformed CSV, or files with missing/empty rsids
  - Reject unsupported file extensions (e.g., `.xls`)
  - Reject duplicate rsids (or deduplicate, according to design)
- **Integration/CLI tests:**
  - CLI succeeds with `--snps-file` for valid JSON and CSV
  - CLI errors if both `--snps` and `--snps-file` are provided
  - CLI errors if neither is provided
  - CLI errors if file is missing, unreadable, or is a directory
  - CLI errors if file is empty or contains only invalid rsids
  - CLI output and downstream logic is identical regardless of input method
- **Documentation:**
  - CLI help output includes new flag and usage
  - Example files and usage in README

## Out of Scope
- No support for Excel (`.xls`, `.xlsx`) or other formats at this time.
- No automatic detection of format beyond extension.

---

**Rationale**: This change enables users to work with large or externally generated SNP lists in a convenient, robust, and script-friendly manner.
