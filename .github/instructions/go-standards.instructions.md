---
applyTo: "**/*.go"
---
# PHITE Project - Go Coding Standards

## General Guidelines
- Follow standard Go style conventions (gofmt)
- Organize imports alphabetically
- Write descriptive function and variable names
- Include comments for exported functions and types
- Implement proper error handling patterns
- Use table-driven tests when appropriate

## Project-Specific Patterns
- Use TDD (Test-Driven Development) workflow
- Start with failing tests (red)
- Implement minimum code to pass tests (green)
- Refactor for clarity and performance while keeping tests passing
- Prefer dependency injection for testability
- Use the project's logger package for consistent logging

## SNP Data Handling
- Validate genetic data inputs thoroughly
- Document assumptions about data formats
- Include references to scientific literature where applicable
- Handle missing or incomplete genetic data gracefully
