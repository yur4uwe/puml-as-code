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

* [x] **E1: Struct Generation** — Map `class` and `struct` entities to Go `type Foo struct { ... }`. Field names and receiver scoping are properly mapped in `struct.go.tmpl`.
* [x] **E2: Interface Generation** — Map `interface` entities to Go `type Foo interface { ... }`. Methods and embeddings are emitted.
* [x] **E3: Abstract Class → Interface** — Map `abstract class` to Go `interface`. Hard error if the abstract class declares any fields (Go has no abstract structs with state). Implemented in `SemanticPass`.
* [x] **E4: Enum Generation** — Map `enum` entities to an `int` type + `iota` const block: `type Color int` + `const ( ColorRed Color = iota ... )`. Enum member names are prefixed with the enum type name.
* [ ] **E5: Visibility Mapping** — `+` (public) → exported PascalCase name. `-` (private), `#` (protected), `~` (package) → unexported camelCase name. `#` and `~` emit `// protected` / `// package-private` comments.
* [ ] **E6: Member Modifiers** — `{static}` members are emitted as package-level functions or package-level variables/constants (not receiver methods/struct fields). `{abstract}` methods are emitted only into companion interfaces.
* [x] **E7: Inheritance (--|>) → Embedding** — `Bar --|> Foo` where `Foo` is a struct → anonymous embed: `type Bar struct { Foo; ... }`. Where `Foo` is an interface → interface embedding: `type Bar interface { Foo; ... }`. Multiple struct embedding is fully supported (Go naturally supports embedding multiple structs and resolves field selectors through embedding hierarchy).
* [x] **E8: Realization (..|>) → Compile-time Interface Check** — `Bar ..|> IFoo` → emit `var _ IFoo = (*Bar)(nil)` as a compile-time satisfaction assertion. Fixed template reference scoping (`(*{{$.Name}})(nil)`).
* [x] **E9: Composition & Aggregation → Struct Fields** — Correct relationship ownership: add fields to the owning (source) struct pointing to the target. Relationship cardinality drives the field type: `1` or unset → value type (`Engine Engine`), `0..1` → pointer (`Engine *Engine`), `0..*` or `*` → slice (`Engines []Engine`), fixed `N` → array (`[N]Engine`).
* [x] **E10: Association (-->) → Pointer Field** — `Car --> Engine` → non-owning reference field (`Engine *Engine` or `[]*Engine`). Cardinality rules from E9 apply.
* [ ] **E11: Dependency (..>) → Comment** — `Service ..> Repository` → emit `// Service depends on Repository` as a type-level comment.
* [ ] **E12: Generic Types & Type Parameters** — Parse and render Go type parameters (`[T any]`, `[K comparable, V any]`) on structs and interfaces from `ast.Entity.Generic`.
* [ ] **E13: Doc Comments, Trivia & Notes** — Map leading/trailing `ast.Trivia` and attached UML notes (`ast.Note`) to idiomatic Go doc comments preceding types, fields, and methods.
* [x] **E14: Cross-Package Qualified References** — Resolve cross-package entity references and prefix types with their package names (e.g. `auth.User`) when referencing entities in other packages.
* [x] **E15: Imports Resolution** — Generate an `import (...)` block in `file.go.tmpl` containing both standard library imports (e.g., `time`, `context`) and internal cross-package imports. Normalize short package names to full import paths. (Deferred: package alias disambiguation when multiple packages share the same short name, e.g. `auth.v1` vs `api.v1`).
* [ ] **E16: Extended Entity Kinds** — Map `record` and `dataclass` to structs, `protocol` to interface, and `exception` to a struct implementing Go's `error` interface (`Error() string`).
* [x] **E17: Relationship Labels & Field Naming** — Use relationship labels or role names (e.g. `Order *-- "*" Item : items` or `: -items`) to derive field names instead of always defaulting to the target type name.
* [x] **E18: Sketch-Grade / Untyped Fields & Parameters** — Safely handle Level 1 sketch types where field or parameter types are omitted (`Type == nil`), defaulting to `any` instead of panicking.
* [ ] **E19: Class Separators / Section Comments** — Render `ast.ClassSeparator` dividers (`-- Section --`, `.. Private ..`) as formatted section comments within struct and interface declarations.
* [ ] **E20: Struct Tags (Deferred)** — Defer struct tags syntax design (` `...` ` or modifiers) until core generation features are completed.

### Category F: Generator Infrastructure
> Cross-cutting concerns that E-series items are built on top of.

* [x] **F1: Symbol Resolution Pass** — Before generation, walk all `Diagram.Statements` and build symbol tables linking entities and relationships.
* [x] **F2: Template Engine Setup** — Go templates parsed via `embed.FS` (`file.go.tmpl`, `struct.go.tmpl`, `interface.go.tmpl`, `enum.go.tmpl`).
* [x] **F3: Multi-File Output Strategy** — Group entities by package path into subdirectories (`<pkg>/types.go`), with root entities written to `types.go`. Ensure package names align with diagram or folder names.
* [x] **F4: go/format Output Pass** — After template rendering, run `go/format.Source` on each generated file. Surface formatting errors with raw source context.
* [x] **F5: Generator Test Harness** — Golden-file integration tests for the generator, mirroring the parser's `integration_test.go` pattern. Each fixture is a `.puml` input paired with expected `.go` files.
* [ ] **F6: Generator CLI Integration** — Connect generator to CLI pipeline in `cmd/` for automated code generation from input `.puml` files to target output directories.

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
* **Goal:** Produce valid, `go/format`-clean Go source files from a parsed class diagram handling full data scope.
* **Tasks:**
  1. [x] Fix template scoping bugs in `struct.go.tmpl` and ensure baseline tests pass.
  2. [x] Implement **F5** (generator golden test harness) with initial fixtures.
  3. [x] Fix relationship ownership in **E9** (composition/aggregation) and implement **E10** (association) & **E17** (relation labels/roles).
  4. [x] Implement **E15** (imports block generation) and **E14** (cross-package qualification).
  5. [ ] Implement **E12** (generics) and **E18** (Level 1 untyped fallback to `any`).
  6. [ ] Implement **E13** (doc comments from notes & trivia) and **E19** (class separators).
  7. [ ] Implement **E6** (static modifiers) and **E16** (extended entity kinds: exceptions, records, protocols).
  8. [ ] Connect generator into CLI in **F6**.
