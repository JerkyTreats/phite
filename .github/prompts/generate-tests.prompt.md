---
mode: 'edit'
description: 'Generate tests for a module following TDD approach'
---
# Generate Tests for PHITE Module

Create comprehensive tests for the selected module following the PHITE project's TDD (Test-Driven Development) approach.

## Test Requirements

1. Start with basic functionality tests (happy path)
2. Add edge case tests (empty input, malformed data, etc.)
3. Include performance tests where relevant (especially for data processing)
4. Ensure test names clearly describe what is being tested
5. Follow the project's test naming conventions
6. Use table-driven tests for Go code where appropriate

## Project Context
- Apply PHITE's layered information architecture principles
- Tests should validate genetic data handling is accurate and safe
- Include tests for error conditions and recovery

Reference the project's existing test patterns in similar modules to maintain consistency.

## Language-Specific Guidelines
- For Go: Use the standard testing package and follow Go test conventions
- For Python: Use pytest fixtures and parameterized tests when appropriate

I'll review the existing code and generate appropriate tests that follow these guidelines.
