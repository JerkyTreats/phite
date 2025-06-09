---
applyTo: "**/*.py"
---
# PHITE Project - Python Coding Standards

## General Guidelines
- Follow PEP 8 style guide
- Use type hints for function parameters and return values
- Document functions, classes, and modules with docstrings
- Organize imports: standard library, third-party, local
- Use virtual environments to manage dependencies
- Write unit tests for new code

## Project-Specific Patterns
- Follow TDD (Test-Driven Development) workflow
  - Write failing tests first (red)
  - Implement minimum code to pass tests (green)
  - Refactor while keeping tests passing
- Use pandas for data manipulation tasks
- Handle exceptions properly with context information
- Use logging rather than print statements

## Genetic Data Processing
- Document any assumptions about input data formats
- Validate genetic data inputs thoroughly
- Handle missing or incomplete data appropriately
- Include citations to relevant research papers
- Consider performance for large datasets
