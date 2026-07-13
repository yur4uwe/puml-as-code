# Design Decision: Contextual Keyword Resolution

## Problem

PlantUML places no restriction on keyword usage — any keyword (`folder`, `namespace`, `class`, `note`, etc.) can also appear as an identifier. A tiered design that hard-tokenizes some keywords and soft-resolves others is therefore not viable; the ambiguity is universal, not limited to containers.

## Decision

Adopt uniform contextual (soft) keyword resolution across the entire grammar. No keyword receives a dedicated token type.

## Logic

1. **Tokenization stays keyword-blind.** Every keyword-shaped literal is emitted as a generic identifier token. The lexer carries no knowledge of the keyword set and requires no changes as new keywords are added.

2. **Keyword identity is resolved on demand, not cached.** A single lookup table maps literal text to a keyword kind, queried only at the specific points where a statement's kind must be decided. Most tokens (punctuation, operators, braces) never touch this table. This keeps classification cheap without needing to store keyword metadata on every token.

3. **One canonical keyword registry.** All keywords in the grammar are defined in a single table. There is no split between "always safe as identifier" and "sometimes a keyword" — that distinction doesn't exist in PlantUML, so the registry doesn't encode it either.

4. **Declaration vs. reference is resolved by one shared rule, reused everywhere.** Whenever an identifier's literal matches a known keyword, a single shared lookahead routine determines whether it's being *declared* (followed by things like a brace, alias marker, stereotype, or color spec) or *referenced* (followed by a relationship operator). Every keyword-dispatch site in the parser calls this same routine — there is exactly one definition of what "declaration context" means, not one per keyword family.

5. **Name, alias, and member parsing require no changes.** These paths already accept generic identifier tokens. Since keywords are never distinguished at the token level, they pass through unchanged and require no keyword-awareness.

## Rationale

- Matches PlantUML's actual grammar (no keyword is ever off-limits as an identifier), so no future keyword addition can violate the design's assumptions.
- Keeps the lexer fully decoupled from keyword semantics; all keyword knowledge lives in one place in the parser.
- Avoids per-family duplication of declaration/reference logic, which was the main consistency risk in the original two-option framing.
- Negligible runtime cost at typical file sizes; not a driving factor in the decision either way.

## Open item before implementation

The shared declaration/reference lookahead depends on an exhaustive terminator set — every relationship operator form (`-->`, `..>`, `o--`, `*--`, `<|--`, etc.) must be correctly recognized, whether the lexer emits these as single compound tokens or decomposed sequences. This set needs to be enumerated against the actual token definitions before the rule can be trusted.
