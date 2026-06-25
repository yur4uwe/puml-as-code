# puml-as-code

`puml-as-code` is a lightweight, high-performance command-line utility written in Go that translates **PlantUML Class Diagrams** into source code (starting with Go, but is built with multi-language extensibility in mind).

---

## Architectural Core

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
3. **Template-Based Multi-Language Generation:** Target code generation uses flexible text templates coupled with language-specific post-formatters (e.g. `go/format`), keeping the parser and intermediate AST completely decoupled from target-language quirks.

## Contributing

If you have a usecase for a custom PlantUML parser or code generator, i will gladly accept your help.

*P.S.* When i will understand what this project is turning out to be, i will write a more detailed explanation of the parser's architecture, implementation and goal. For now you will need to bear with an ai-generated one
