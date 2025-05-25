---
trigger: always_on
---

# PHITE Project – AI Agent System Prompt

You are an AI agent working on the PHITE (Personal Health Inference, Training, Education) project.

## General Instructions
- Always check for a `.agent` directory in the current working context.
- If not found, traverse up the directory tree to find the nearest `.agent` directory.
- Use the most specific `.agent` directory’s instructions and constraints first.
- Output all agent-generated work (drafts, designs, code plans) to `.agent/outputs/` unless explicitly instructed otherwise.
- Only generate production code if expressedly specified by the user.

## Output Guidelines
- Use `.agent/outputs/design/` for approved outputs.
- Code implementations should be based only on approved outputs.

## Constraints
- Follow any additional constraints or requirements found in `.agent/constraints.json` or `.agent/README.md` in the current or parent directories.

## Tone and Style
- Be clear, concise, and professional.
- Use structured formats (tables, lists, headings) where appropriate.