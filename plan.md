# Project Implementation Plan & Backlog 📋

This document outlines the roadmap, feature backlog, and implementation rules for `puml-as-code`. It is structured to scale incrementally as new features are requested, without requiring document restructures.

---

## 🚦 Architectural Guardrails
* **Parser:** Hand-written recursive descent.
* **Symbol Resolution:** Single-pass using pointers to shared [ast.Entity](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/ast/structs.go#L20-L25) nodes.
* **Generators:** Language-agnostic AST mapped via `text/template` and formatted via target language formatting packages (e.g. Go's standard `go/format`).
* **AST Node Naming Convention:** `*Directive` — preprocessor-style directives that modify parsing context or file resolution (e.g. `IncludeDirective`). `*Command` — runtime diagram instructions that alter rendering or visibility (e.g. `VisibilityCommand`, `SetCommand`, `DirectionCommand`).
* **Include Resolution:** `!include` is parsed as `ast.IncludeDirective` and left in the AST. A separate post-parse `Resolver` pass splices included files' statements in place. Path resolution follows PlantUML's own policy: paths are relative to the file currently being processed. The parser has no file system access.

---

## 🛠️ Feature Implementation Protocol
For any feature added to the backlog, follow this 5-step implementation checklist:
1. **Define Test Input:** Add a test case `.puml` in `input/` demonstrating the target syntax.
2. **Lexer Check:** Ensure the tokenizer emits proper tokens (add definitions in [tokens.go](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/tokens.go) and tests in [lexer_test.go](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer/lexer_test.go) if new syntax tokens are needed).
3. **Parser Implementation:** Implement statement parsing in [parser.go](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/parser.go) (or related files) and verify it registers correctly inside the symbol table.
4. **Code Generation:** Update the template and rendering methods in [generator.go](file:///home/yur4uwe/Projects/puml-as-code/pkg/generator/go/generator.go) to output correct source code structures.
5. **Regression Verification:** Run tests and ensure the output matches expected target code (e.g. in `output/`).

---

## 🗂️ Feature Backlog

### Category A: Entity Declarations
* [x] **A1: Class Headers** — Parse basic `class Name` and aliases (`class Name as "Alias"`).
* [x] **A2: Interfaces & Structs** — Support `interface` and `struct` keywords.
* [x] **A3: Enums** — Support `enum` type parsing.
* [x] **A4: Generic Types** — Parse class template/generic types (e.g., `class List<T>`).
* [x] **A5: Stereotypes** — Parse double-bracket stereotypes (e.g., `class Service <<API>>`).

### Category B: Entity Members (Body Parsing)
* [x] **B1: Body Enclosures** — Parse `{ ... }` curly braces for classes/interfaces.
* [x] **B2: Fields & Types** — Parse variables with their visibility modifiers (`+`, `-`, `#`, `~`) and basic/custom types (e.g. `+size: int`).
* [x] **B3: Methods & Parameters** — Parse functions with parameters and return types (e.g. `+Drive(dest: string): error`).
* [x] **B4: Member Modifiers** — Parse `{static}` and `{abstract}` keyword prefixes.
* [x] **B5: Visual Dividers** — Handle visual groupings inside entities (e.g. `..`, `--`, `==`).

### Category C: Relationships & Associations
* [x] **C1: Basic Arrows** — Parse inheritance (`<|--`), composition (`*--`), aggregation (`o--`), dependency (`..>`), and association (`-->`).
* [x] **C2: Directional Modifiers** — Parse directional tokens (e.g., `-up->`, `-left-`).
* [x] **C3: Multiplicity** — Extract multiplicities (e.g., `"1" *-- "0..*"`).
* [x] **C4: Relation Labels** — Parse relationship titles and arrows (e.g. `: contains >`).

### Category D: Global Directives & Scoping
* [x] **D1: Scoping Blocks (Packages)** — Parse `package Name { ... }` boundaries and group entities cleanly.
* [x] **D2a: Parse Include Directives** — Parse `!include <path>` and `!include <path>!<tag>` into `ast.IncludeDirective{Path, Tag}`. `!include_many` and `!include_once` are accepted with a parse-time warning and produce the same node; no `Kind` field is stored in the AST since the resolver treats all three identically. `Tag` holds either a numeric index (`"0"`, `"1"`) or a named ID (`"MY_ID"`); the resolver distinguishes them via `strconv.Atoi`.
* [x] **D2b: Include Resolver Pass** — Post-parse `Resolver` that walks `Diagram.Statements`, finds `IncludeDirective` nodes, and splices the included file's statements in place. Path resolution is relative to the including file (PlantUML's own policy). If `Tag` is set, only the matching `@startuml` block (by index or `id=` attribute) is extracted. Detects circular includes via a `visited` path set and returns a hard error. When a file contains multiple blocks and no `Tag` is specified, emits a warning and uses block 0.
* [x] **D3: Skinparam & Styles** — Parse global design parameters and variables.

### Category E: Go Code Generation
> These items convert parsed AST nodes into valid, idiomatic Go source code.
> They depend on Category F infrastructure being in place first.

* [ ] **E1: Struct Generation** — Map `class` and `struct` entities to Go `type Foo struct { ... }`. Field names are exported/unexported based on visibility (see E5).
* [ ] **E2: Interface Generation** — Map `interface` entities to Go `type Foo interface { ... }`. Only methods are emitted; fields in an interface body are a hard error.
* [ ] **E3: Abstract Class → Interface** — Map `abstract class` to Go `interface`. Hard error if the abstract class declares any fields (Go has no abstract structs with state).
* [ ] **E4: Enum Generation** — Map `enum` entities to an `int` type + `iota` const block: `type Color int` + `const ( ColorRed Color = iota ... )`. Enum member names are prefixed with the enum type name.
* [ ] **E5: Visibility Mapping** — `+` (public) → exported PascalCase name. `-` (private), `#` (protected), `~` (package) → unexported camelCase name. `#` and `~` also emit a `// protected` / `// package-private` comment since Go has no equivalent modifiers.
* [ ] **E6: Member Modifiers** — `{static}` members are emitted as package-level functions (not methods). `{abstract}` methods are emitted only into a companion interface, not the struct.
* [ ] **E7: Inheritance (--|>) → Embedding** — `Bar --|> Foo` where `Foo` is a struct → anonymous embed: `type Bar struct { Foo; ... }`. Where `Foo` is an interface → interface embedding: `type Bar interface { Foo; ... }`. Hard error if more than one struct/concrete parent is listed (Go forbids multiple struct inheritance).
* [ ] **E8: Realization (..|>) → Compile-time Interface Check** — `Bar ..|> IFoo` → emit `var _ IFoo = (*Bar)(nil)` as a compile-time satisfaction assertion. No method stubs are generated (the diagram defines structure, not logic).
* [ ] **E9: Composition & Aggregation → Struct Fields** — Relationship cardinality drives the field type: `1` or unset → value type (`Engine Engine`), `0..1` → pointer (`Engine *Engine`), `0..*` or `*` → slice (`Engines []Engine`). Field name is derived from the RHS entity name. Composition and aggregation produce identical output (ownership semantics are not expressible in Go).
* [ ] **E10: Association (-->) → Pointer Field** — `Car --> Engine` → `Engine *Engine` field on `Car`. Treated as a non-owning reference. Cardinality rules from E9 apply.
* [ ] **E11: Dependency (..>) → Comment** — `Service ..> Repository` → emit `// Service depends on Repository` as a top-of-file or type-level comment. No structural code is generated; this is a logical relationship with no direct Go equivalent.

### Category F: Generator Infrastructure
> Cross-cutting concerns that E-series items are built on top of.

* [x] **F1: Symbol Resolution Pass** — Before generation, walk all `Diagram.Statements` and build a `map[string]*ast.Entity` keyed by both `Identifier` and `Alias`. Required by E7–E11 to resolve both sides of a relationship to their AST nodes and determine their kind (struct vs. interface).
* [ ] **F2: Template Engine Setup** — Load Go templates from a `templates/go/` directory at runtime using `os.DirFS` (not `embed.FS`, to preserve symlink intent). Directory structure mirrors entity kinds: `struct.go.tmpl`, `interface.go.tmpl`, `enum.go.tmpl`, with `class.go.tmpl` as a symlink to `struct.go.tmpl`. Templates are composed into per-file output.
* [ ] **F3: Multi-File Output Strategy** — One `.go` file is emitted per `package` block, placed in a subdirectory matching the package name (lowercase). Root-level entities (outside any package block) are written to `<out dir>/types.go` with `package <diagram name>`. The full directory tree mirrors the package hierarchy in the diagram.
* [ ] **F4: go/format Output Pass** — After template rendering, run `go/format` on each generated file. Surface formatting errors as generator errors (a formatting failure indicates a template or logic bug).
* [ ] **F5: Generator Test Harness** — Golden-file integration tests for the generator, mirroring the parser's `integration_test.go` pattern. Each fixture is a `.puml` input paired with an expected `.go` output. Tests fail if generated output diverges from the golden file.

---

## ✅ Completed Milestone: Phase 1 (Parser Foundation)
* Parsed all entity kinds (class, interface, struct, enum, abstract class) with headers, members, modifiers, and stereotypes.
* Parsed all relationship types with directionality, multiplicity, and labels.
* Parsed package/namespace scoping blocks (nested).
* Parsed skinparam and `<style>` blocks.
* Implemented Go dialect for field and method member syntax.
* Full trivia (comment) attachment to statements and members.

---

## 🎯 Active Milestone: Phase 2 (Go Code Generation)
* **Goal:** Produce valid, `go/format`-clean Go source files from a parsed class diagram.
* **Tasks:**
  1. Implement **F1** (symbol resolution pass) as a pre-generation visitor.
  2. Implement **F2** (template engine setup) with the `templates/go/` directory and `os.DirFS` loader.
  3. Implement **F3** (multi-file output) — update `GenerateFromClassDiagram` to return `[]GeneratedFile` instead of `string`.
  4. Implement **E1** (struct) and **E2** (interface) generation end-to-end, including **E5** (visibility) and **F4** (`go/format` pass).
  5. Add **F5** (golden-file test harness) and a first fixture covering structs and interfaces.
  6. Implement **E3** (abstract class), **E4** (enum), **E6** (modifiers).
  7. Implement **E7** (inheritance), **E8** (realization), **E9** (composition/aggregation), **E10** (association), **E11** (dependency).
