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

- [x] **Definitively decide comment handling strategy**: Implemented Trivia Attachment strategy (`LeadingTrivia` & `TrailingTrivia` on AST statement nodes). Updated `TokenStream.Emit()` to preserve leading comment trivia across newlines.
- [x] **Decide `skinparam` & `<style>` AST representation**: Confirmed Unified Style AST plan. Both `skinparam` statements and CSS `<style>` blocks will parse into `ast.StyleRule` statement nodes (`Selectors []string`, `Properties map[string]string`) in AST order, leaving style cascading to a downstream resolver pass.

---

## 🏷️ Identifier & Relationship Enhancements

- [ ] **Enhanced Identifier & Member Reference Parsing**: Extend identifier parsing to support class member targets, method signatures, and method overloading:
  - `<class>::<field>`
  - `<class>::<methodName>`
  - `<class>::"<method>()"` (method overloading support)

---

## 🧪 Testing & Quality Assurance

- [ ] **Start Integration Testing**: Create end-to-end integration test suites parsing complete `.puml` files and verifying AST structure + target code generation.
