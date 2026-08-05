# Next Steps & TODO Backlog 📌

This document lists the architectural refactoring tasks, API enhancements, and design decisions to tackle next in `puml-as-code`.

---

## 🛠️ Tokenizer & TokenStream Refactoring

- [x] (with alterations to the plan) **Move separator change detection from lexer to token stream**: Shift package separator (`::`) customization/detection logic out of `Lexer` into `TokenStream`.
- [x] (Dropped) **Move comment parsing to token stream**: Retained in lexer (`ResolveUnambiguousToken`) because comment scanning is stateless, unambiguous (0-1 lookahead), and requires no stream state.
- [x] **Re-evaluate `TokenCollector` necessity**: Refactored `ConsumeUntil`/`ConsumeUntilType` to return consumed `[]Token` slice directly without consuming terminator. Retained `TokenCollector` and `TokenSink` as unused declarations for future extension.
- [x] **Enhance & Audit Stream API**: Completed 4-step audit (purged 4 dead methods, standardized parameter names, exported `ReadBetween`, moved PUML domain logic to `pkg/parser/domain_helpers.go`, detaching `ast` package imports from `pkg/tokenizer`).
- [x] (Dropped) **Create `tokenization-quirks.md`**: Dropped as tokenizer behavior has no distinct parser-divergent quirks requiring standalone documentation.

---

## 💬 Comment & AST Handling

- [ ] **Definitively decide comment handling strategy**: Design and implement a consistent policy for comment preservation, trivia attachment, or comment statement nodes across the parser.
- [ ] **Decide `skinparam` AST representation**:
  - Determine whether `skinparam` rules should be explicitly preserved in AST order:
    ```go
    ast.SkinparamRule{
        Ident: "backgroundcolor",
        Val:   "#00ff00",
    }
    ```
  - Or whether they should only be aggregated into an unordered lookup registry/map.

---

## 🏷️ Identifier & Relationship Enhancements

- [ ] **Enhanced Identifier & Member Reference Parsing**: Extend identifier parsing to support class member targets, method signatures, and method overloading:
  - `<class>::<field>`
  - `<class>::<methodName>`
  - `<class>::"<method>()"` (method overloading support)

---

## 🧪 Testing & Quality Assurance

- [ ] **Start Integration Testing**: Create end-to-end integration test suites parsing complete `.puml` files and verifying AST structure + target code generation.
