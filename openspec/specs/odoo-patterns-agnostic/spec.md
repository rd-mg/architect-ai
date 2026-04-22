# Odoo Patterns Agnostic Specification

## Purpose
Define the structural and content requirements for version-agnostic Odoo development patterns within the Architect-AI skill overlays. This ensures high signal density, version accuracy, and consistent coding standards across all supported Odoo versions (14.0 - 19.0).

## Requirements

### Requirement: Distilled File Structure
Each pattern markdown file MUST follow the standardized "Distilled" hierarchy to ensure consistent navigation and automated parsing.

- **Header**: Standardized domain metadata block.
- **Consolidated from**: References to original component files.
- **Source**: Domain-specific pattern blocks.
- **Anti-Patterns**: Explicit "do-not-do" rules to prevent common Odoo pitfalls.
- **Version Notes**: A mandatory compatibility matrix (v14-v19).

### Requirement: Version Sensitivity Matrix
To prevent LLM hallucinations of deprecated or future syntax, every pattern file MUST include a version matrix covering v14-v19.

### Requirement: Clean Naming Standards
Code patterns MUST promote professional development practices by avoiding UI-generated naming conventions (e.g., NO `x_` prefixes).

### Requirement: Size and Signal Density
Pattern files SHOULD NOT exceed 600 lines to optimize for LLM context windows.
