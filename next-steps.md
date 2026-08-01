# Next Steps & TODO Backlog 📌

This document lists the architectural refactoring tasks, API enhancements, and design decisions to tackle next in `puml-as-code`.

---

## 🛠️ Tokenizer & TokenStream Refactoring

- [ ] **Move separator change detection from lexer to token stream**: Shift package separator (`::`) customization/detection logic out of `Lexer` into `TokenStream`.
- [ ] **Move comment parsing to token stream**: Relocate line and block comment reading logic into `TokenStream` so that `Lexer` remains lightweight and focused strictly on raw token emission.
- [ ] **Re-evaluate `TokenCollector` necessity**: Audit the `TokenCollector` utility alongside the `ConsumeUntil` family of methods to determine if stream peeking/slicing can replace them cleanly.
- [ ] **Enhance & Audit Stream API**: Perform a holistic review of `TokenStream` methods to eliminate awkward patterns, redundant functions, or poor UX/API ergonomics.
- [ ] **Create `tokenization-quirks.md`**: Document all edge cases, lexer quirks, raw modes, and tokenization subtleties in a dedicated markdown document.

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
