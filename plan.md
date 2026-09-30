# Project Execution Plan & Backlog 📋

This document outlines the architecture, completed foundation, and active execution plan for `puml-as-code`.

---

## 🚦 Architectural Guardrails

* **Parser:** Hand-written recursive descent. The parser has no direct filesystem access; file inclusion is deferred to the [`resolver`](file:///home/yur4uwe/Projects/puml-as-code/pkg/resolver) pass.
* **Lexer / Tokenizer:** Hand-written scanner in [`pkg/tokenizer`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer).
* **Token Source Fidelity Invariant:**
  ```go
  source[token.Span.Start.Offset : token.Span.End.Offset] == rawSourceLexeme
  ```
  `Token.Literal` holds the semantic/normalized value (e.g. unquoted string, trimmed comment), while `Token.Span` captures the exact source coordinates. Lexical/source coordinates must never be derived from semantic values.
* **Symbol Resolution:** Single-pass symbol table construction in [`pkg/resolver`](file:///home/yur4uwe/Projects/puml-as-code/pkg/resolver) with pointers to shared [`ast.Entity`](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/ast/structs.go#L70-L80) nodes. Statements that do not produce Go types (such as [`ast.UnhandledStatement`](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/ast/globals.go#L38-L42)) are safely ignored during symbol resolution.
* **Generators:** Language-agnostic AST mapped via `text/template` in [`pkg/generator`](file:///home/yur4uwe/Projects/puml-as-code/pkg/generator) and formatted via target language formatting packages (e.g., standard `go/format`).
* **AST Node Semantics:**
  * `*Directive` — preprocessor directives modifying parsing context or file resolution (e.g. `IncludeDirective`).
  * `*Command` — diagram instructions altering rendering or visibility (e.g. `VisibilityCommand`, `SetCommand`, `DirectionCommand`).
  * `UnhandledStatement` — lossless fallback statement capturing unhandled directives or standalone diagram instructions.

---

## ✅ Completed Foundation

### Phase 1: Parser & Lexer Foundation
* [x] Lexical analysis with token streaming, trivia collection, and source position tracking.
* [x] All entity headers: `class`, `interface`, `struct`, `enum`, `abstract class`, `record`, `dataclass`, `protocol`, `exception`.
* [x] Entity bodies: fields, methods, parameters, types, visibility modifiers (`+`, `-`, `#`, `~`), and static/abstract member modifiers.
* [x] Relationships: inheritance (`<|--`), composition (`*--`), aggregation (`o--`), association (`-->`), dependency (`..>`), with multiplicities and labels.
* [x] Containers & scoping: `package`, `namespace`, `together`, `folder`, `frame`, `node`, `database`, `cloud`, `rectangle`.
* [x] Layout commands: `title`, `header`, `footer`, `legend`, `caption`, `sprite`.
* [x] Preprocessor directives: `!include`, `!include_once`, `!include_many`.
* [x] Skinparam and `<style>` blocks.
* [x] Comments & trivia preservation: leading/trailing comments attached to statements, entities, and members.

### Phase 2: Go Code Generation & Symbol Resolution
* [x] Symbol table resolution pass linking cross-package entities and relationships.
* [x] Template-based code generator producing formatted Go code via `go/format`.
* [x] Multi-file output strategy grouping entities into package subdirectories.
* [x] Struct generation with composition/aggregation mapped to struct fields, and inheritance mapped to struct embedding.
* [x] Interface generation with embedding and compile-time satisfaction checks (`var _ Interface = (*Struct)(nil)`).
* [x] Enum generation with typed constants and `iota`.
* [x] Cross-package imports and standard library imports resolution.
* [x] Generic types and type parameters (`[T any]`).
* [x] Comprehensive generator golden test harness ([`pkg/generator/go/integration_test.go`](file:///home/yur4uwe/Projects/puml-as-code/pkg/generator/go/integration_test.go)).

