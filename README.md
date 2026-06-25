# puml-as-code 🚀

`puml-as-code` is a lightweight, high-performance command-line utility written in Go that translates **PlantUML Class Diagrams** into clean, idiomatic source code (starting with Go, but built for multi-language extensibility).

The goal is to bridge the gap between design and implementation, allowing developers to maintain a single source of truth in design documents and generate boilerplate code instantly.

---

## 🏗️ Architectural Core

The project is built on three fundamental architectural principles:

```
┌──────────────┐      ┌───────────────┐      ┌─────────────────────────┐
│  PUML Source │ ───> │ Manual Parser │ ───> │ Language-Agnostic AST   │
└──────────────┘      └───────────────┘      └─────────────────────────┘
                                                          │
                                                          ▼
                                             ┌─────────────────────────┐
                                             │  Template Generators    │
                                             │  (Go, TS, etc.+ Formats)│
                                             └─────────────────────────┘
```

1. **Hand-Written Recursive-Descent Parser:** Written completely from scratch in Go to maintain deep control over token parsing, boundaries, and syntax validation.
2. **Single-Pass Pointer-Based Symbol Table:** To avoid complex AST traversals, the parser uses a unified `symbol_table map[string]*ast.Entity` during execution. Relationships (like Composition or Inheritance) link directly to these shared pointers, resolving references on the fly.
3. **Template-Based Multi-Language Generation:** Target code generation uses flexible text templates coupled with language-specific post-formatters (e.g., `go/format`), keeping the parser and intermediate AST completely decoupled from target-language quirks.

---

## 🛠️ Tech Stack
* **Language:** Go 1.23.4
* **Dependencies:** Standard library, `github.com/stretchr/testify` for testing, and `gopkg.in/yaml.v3` for token regression data.
