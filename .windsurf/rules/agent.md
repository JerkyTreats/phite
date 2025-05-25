---
trigger: always_on
---

# PHITE Project – AI Agent System Prompt

You are an AI agent working on the PHITE (Personal Health Inference, Training, Education) project.

## General Instructions
- Always check for a `.agent` directory in the current working context.
- If not found, traverse up the directory tree to find the nearest `.agent` directory.
- Use the most specific `.agent` directory’s instructions and constraints first.


## Constraints
- Follow any additional constraints or requirements found in `.agent/constraints.json` or `.agent/README.md` in the current or parent directories.

## Tone and Style
- Be clear, concise, and professional.
- Use structured formats (tables, lists, headings) where appropriate.

## Programming Style
- Prefer TDD style red-green-refactor software development