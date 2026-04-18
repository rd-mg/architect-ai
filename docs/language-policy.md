# Language Policy — English-Only Voice, Multilingual Triggers

**Status**: Adopted. Effective immediately.

## Rule

All **framework-emitted content** is English. This includes:

- Any `.md` file under `internal/assets/` that the agent reads as instruction (SKILL.md, sdd-orchestrator.md, sdd-phase-protocols/*.md, persona-*.md, output-style-*.md, engram-protocol.md, etc.).
- Any comment, error message, log statement, or docstring in Go source.
- All user-facing strings (TUI labels, help text, CLI messages).
- All documentation under `docs/` and the top-level README.

**Trigger phrases are exempt.** A trigger phrase is an exact string the agent's matcher listens for in user input. It is data, not voice.
Examples of allowed trigger phrases:

- `"iniciar sdd"` (Spanish equivalent of "init sdd")
- `"vamos con sdd"` (Spanish equivalent of "let's do sdd")
- `"recordar"` (Spanish memory-retrieval trigger)
- `"listo"` (Spanish session-close trigger)

These trigger phrases exist to improve UX for Spanish-speaking developers. Removing them is a UX regression with no measurable gain.

## Where Trigger Phrases Are Allowed

Trigger phrases may appear ONLY inside these blocks:

1. A markdown table whose leftmost column is labeled `Trigger` or `Pattern`.
2. A YAML frontmatter block whose key is `triggers:` or `description:`.
3. An explicit comment `<!-- trigger-phrase-allowlist -->` at end of line.

Outside these blocks, Spanish content is a violation.

## Adding a New Trigger Phrase

1. Place the phrase in one of the three allowed blocks above.
2. Adjacent to the Spanish phrase, include the English equivalent.
3. Update `scripts/lint-language.sh` only if you introduce a new Spanish word that isn't already in the lint's allowlist regex.

## Removing a Trigger Phrase

Open an issue first. Trigger-phrase removal is a UX change that affects users who rely on them.

## Lint Enforcement

`scripts/lint-language.sh` runs on every PR. It searches for a curated word list of common Spanish words in Cat-A file scopes and fails if hits are found outside allowed blocks.
