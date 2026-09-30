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

---

## 🎯 Active Milestone: Token Source Fidelity & Unhandled Statements Golden Suite

Addressing the token source width discrepancy identified during [`ast.UnhandledStatement`](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/ast/globals.go#L38-L42) testing (documented in [`TOKEN_SOURCE_FIDELITY.md`](file:///home/yur4uwe/Projects/puml-as-code/TOKEN_SOURCE_FIDELITY.md)), and establishing full golden test coverage for unhandled statements.

### Phase 1: Tokenizer Source Fidelity (`pkg/tokenizer`)
* [ ] **1.1 Add SourceSpan to Token**
  * Update [`Token`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/tokens.go#L71-L75) to embed `Span SourceSpan`, while retaining `Pos Pos` as a backward-compatible alias for `Span.Start`.
  * Update [`EndOffset()`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/tokens.go#L77-L79) to return `t.Span.End.Offset`.
  * Update [`EndPos()`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/tokens.go#L81-L88) to return `t.Span.End`.
* [ ] **1.2 Update Token Emission in Lexer**
  * Ensure every token scanning path in [`pkg/tokenizer/tokens.go`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/tokens.go) and [`pkg/tokenizer/lexer.go`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/lexer.go) captures `start := l.getPos()` before reading and `end := l.getPos()` after reading.
  * Correctly record spans for strings (with delimiters and escape sequences), line comments, and multiline block comments.
* [ ] **1.3 Tokenizer Invariant & Tiling Tests**
  * Add unit tests in [`pkg/tokenizer/lexer_test.go`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/lexer_test.go) asserting:
    1. `source[tok.Span.Start.Offset : tok.Span.End.Offset] == expectedRaw`
    2. `tok.Literal == expectedSemanticValue`
  * Test matrix: `"abc"`, `"abc\"def"`, non-ASCII `"Привіт"`, `'  comment  '`, `/' multiline\nblock comment '/`.

### Phase 2: Parser Verification (`pkg/parser`)
* [ ] **2.1 Unhandled Statement Slicing Verification**
  * Verify that [`consumeUnhandledLine`](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/unhandled.go#L35-L59) slices string-ending directives (e.g. `!define VERSION "v1.2.3"`) with full fidelity without truncation.
* [ ] **2.2 Layout Statement Slicing Verification**
  * Verify that [`parseLayoutStatement`](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/branches.go#L1702) and [`SliceInputBetweenTokens`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/stream.go#L491-L497) accurately slice single-line titles/headers ending in quoted strings.
* [ ] **2.3 Parser Unit Tests Regression**
  * Run `go test ./pkg/parser/...` and verify all tests pass.

### Phase 3: Integration & Golden Test Suite
* [ ] **3.1 Finalize Integration Test PUML Diagram**
  * Complete [`input/integration_testdata/unhandled_statements.puml`](file:///home/yur4uwe/Projects/puml-as-code/input/integration_testdata/unhandled_statements.puml) covering all 5 core verifications:
    1. Root single-line statements (`caption`, `sprite`).
    2. Root single-line directives (`!define`, `!global`, `!theme`) with quoted and unquoted values.
    3. Root block directives (`!function`, `!procedure`, `!definelong`, nested `!if`, `!while`, `!foreach`).
    4. Container-scoped unhandled statements (inside `package CoreService`).
    5. Trivia preservation (leading/trailing comments) and coexistence with real entities (`AuthService`, `UserStore`) and relationships.
* [ ] **3.2 Generate & Validate Golden AST JSON**
  * Run `go test ./pkg/parser -update-golden` to generate [`input/integration_testdata/unhandled_statements.golden.json`](file:///home/yur4uwe/Projects/puml-as-code/input/integration_testdata/unhandled_statements.golden.json).
  * Verify AST node JSON contains exact `Text`, `Span`, and `Trivia` fields.
* [ ] **3.3 Verify Generator Tolerance**
  * Confirm that diagrams containing `ast.UnhandledStatement` nodes resolve cleanly and do not break code generation.
* [ ] **3.4 Full Test Suite Run**
  * Run `go test ./...` across the entire workspace to ensure 100% green status.
