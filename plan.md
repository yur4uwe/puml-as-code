# Project Implementation Plan & Backlog 📋

This document outlines the roadmap, feature backlog, and implementation rules for `puml-as-code`. It is structured to scale incrementally as new features are requested, without requiring document restructures.

---

## 🚦 Architectural Guardrails
* **Parser:** Hand-written recursive descent.
* **Symbol Resolution:** Single-pass using pointers to shared [ast.Entity](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/ast/structs.go#L20-L25) nodes.
* **Generators:** Language-agnostic AST mapped via `text/template` and formatted via target language formatting packages (e.g. Go's standard `go/format`).

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
* [ ] **D2: Import Directives** — Parse `!include` files.
* [x] **D3: Skinparam & Styles** — Parse global design parameters and variables.

---

## 🎯 Active Milestone: Phase 1 (Entity Headers)
* **Goal:** Parse headers and types for simple classes, structures, and interfaces.
* **Tasks:**
  1. Add a sample class-only diagram to `input/no-rel.puml`.
  2. Implement parser logic in [parser.go](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/parser.go) to match class/interface declarations.
  3. Verify entities are properly added to the `symbol_table`.
  4. Generate simple Go struct declarations.
