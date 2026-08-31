# AGENTS.md

## 1. Project Overview & Architecture

### Purpose

Command-line compiler/transpiler written in Go that parses PlantUML Class Diagrams and emits formatted, idiomatic source code (starting with Go, structured for multi-target language extensibility).

### Core Pipeline Flow

  1. **Lexical Analysis:** Hand-written lexer in [`pkg/tokenizer`](file:///home/yur4uwe/Projects/puml-as-code/pkg/tokenizer) converts raw PUML into token streams.
  2. **Parsing & AST Construction:** Hand-written recursive descent parser in [`pkg/parser`](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser) builds language-agnostic AST nodes defined in [`pkg/parser/ast`](file:///home/yur4uwe/Projects/puml-as-code/pkg/parser/ast).
  3. **Include Resolution:** Post-parse preprocessing in [`pkg/resolver`](file:///home/yur4uwe/Projects/puml-as-code/pkg/resolver) evaluates `!include` directives and resolves file dependencies/diagram tags.
  4. **Code Generation:** Template-based generator in [`pkg/generator`](file:///home/yur4uwe/Projects/puml-as-code/pkg/generator) (e.g. Go backend in [`pkg/generator/go`](file:///home/yur4uwe/Projects/puml-as-code/pkg/generator/go)) performs semantic validation, populates view models, renders templates, and runs target language formatters (`go/format`).
* **Key Invariants & Guardrails:**
  * **Parser I/O Isolation:** The parser has no direct filesystem access; all external file inclusion is deferred to the [`resolver`](file:///home/yur4uwe/Projects/puml-as-code/pkg/resolver) pass.
  * **AST Node Semantics:** Preprocessor directives are classified as `*Directive` (e.g. `IncludeDirective`), whereas diagram render/visibility instructions are classified as `*Command` (e.g. `VisibilityCommand`, `DirectionCommand`).

---

## 2. Codebase & Package Structure

```
puml-as-code/
+-- cmd/
|   +-- main.go                       # Root CLI entry point
|   +-- docs-examples-gen/            # Utility tool to generate documentation examples
+-- internal/
|   +-- helpers/                      # Shared internal helper functions
+-- pkg/
|   +-- tokenizer/                    # Lexical scanner & token definitions
|   +-- parser/                       # Recursive-descent parser
|   |   +-- ast/                      # AST node definitions & enums
|   |   +-- dialect/                  # Target-language syntax parsing for members
|   |   +-- keyword/                  # Keyword tables & classification
|   +-- resolver/                     # Include directives & symbol resolver pass
|   +-- generator/                    # Generator factory & interface
|       +-- go/                       # Go code generator & templates
+-- input/
|   +-- integration_testdata/         # Test diagrams & golden fixtures
+-- output/                           # Target directory for generated output
+-- docs/                             # Language specs & dialect documentation
```

### Inspecting Package APIs
When detailed information is needed about package responsibilities, types, or exported functions, inspect the package doc comments and API dynamically using `go doc`:
```bash
go doc ./pkg/<package-path>
# Examples:
go doc ./pkg/parser
go doc ./pkg/parser/ast
go doc ./pkg/generator/go
```
