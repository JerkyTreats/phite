# Change Proposal: Flexible Grouping by Topic or Group in TSV Parser

## Background
Currently, the TSV parser in `converter/internal/converter/tsvparser.go` groups records by the "Group" column and outputs a JSON file per group. There is a need to optionally group by the "Topic" column instead, while retaining the ability to group by Group.

## Proposal
- **Change the output JSON structure for parsed TSV files**:
  - In **topic mode**, output a top-level `Topic` and a `Groupings` object (not array), where each property is a group name mapping to an array of SNPs for that group.
  - In **group mode**, output a single `Grouping` object as before.
  - This ensures that SNPs are correctly associated with their group under each topic.
- **Add a `GroupingMode` option to the TSV parsing logic**, allowing users to specify whether to group by "Topic" or "Group".
- The grouping mode can be set via a new field in the `TSVParser` struct, a constructor argument, or a configuration parameter.
- Output file naming and structure will reflect the selected grouping mode.
- Default behavior will remain grouping by "Group" to preserve backward compatibility.

### Example Output Structure

**Topic mode:**
```json
{
  "Topic": "Nutrients – Vitamins and Minerals",
  "Groupings": {
    "Acupuncture": [
      { /* SNP 1 */ },
      { /* SNP 2 */ }
    ],
    "MTHFR": [
      { /* SNP 3 */ }
    ]
  }
}
```
**Group mode:**
```json
{
  "Grouping": {
    "Topic": "Immune Response",
    "Name": "Acupuncture",
    "SNP": [
      { /* ... */ }
    ]
  }
}
```

## Implementation Steps
1. **Update `TSVParser` struct** to include a `GroupingMode` field (string: "topic" or "group").
2. **Update constructor** (`NewTSVParser`) to accept a grouping mode argument.
3. **Update the `Parse` method** to use the selected grouping mode for grouping records and naming output files.
4. **Update documentation** and usage examples to describe the new option.

## Benefits
- Adds flexibility for downstream consumers and workflows.
- Maintains backward compatibility.
- Easy to extend for future grouping modes if needed.

## Example Usage
```go
parser := converter.NewTSVParser(inputFile, outputDir, "topic") // or "group"
files, errors, err := parser.Parse()
```

## Constraints
- If grouping by "topic", all output and metadata should consistently use the topic value as the key.
- If grouping by "group", current behavior is preserved.

---

*Proposed: 2025-05-26*
*Author: Cascade AI Agent*

## Testing Plan

The following tests will be implemented or updated to ensure the new JSON structure and flexible grouping feature are robust and backward compatible:

1. **TestParseGroupByTopic**
   - Verify that when grouping by Topic, the parser outputs one file per unique Topic and the JSON structure matches the new spec: top-level `Topic` and a `Groupings` object mapping group names to arrays of SNPs.
   - Confirm that each group under a topic has the correct set of SNPs.
2. **TestParseGroupByGroup**
   - Ensure that grouping by Group (the current default) outputs a file with a single `Grouping` object containing the correct topic, group name, and SNPs.
3. **TestParseWithMixedTopicsAndGroups**
   - Use a TSV with multiple Topics and Groups to confirm correct output for both grouping modes, and that all groupings for a topic appear together in topic-mode output, with SNPs correctly associated by group.
4. **TestInvalidGroupingMode**
   - Ensure the parser handles invalid grouping modes gracefully (error or fallback behavior).
5. **TestOutputIsValidJSON**
   - Validate that all output files are valid JSON and the required fields are present (`Topic` and `Groupings` for topic mode; `Grouping` for group mode).
6. **TestDocumentation**
   - Confirm that documentation and usage examples reflect the new grouping mode option and output structure.

These tests will be added to or extend `tsvparser_test.go`.

