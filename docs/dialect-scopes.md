# Dialect Type Resolution Scopes

This document describes the design and scope of **Dialect Type Resolution** in `puml-as-code`.

To keep the dialect parser simple while avoiding structural information loss, the project uses **Lazy Token-Based / Recursive Type Resolution** for language dialects. Dialects split member syntax and represent type references using language-specific recursive models that are later inspected by target code generators.

---

## Supported Type Scopes (Levels 1–3)

We support type syntax structures up to Level 3. Anything more complex (Level 4+) is out of scope to avoid compiler-grade parsing complexity in dialects.

### Level 1: Sketch-Grade (Implicit Types)
These represent diagrams where types are omitted, or method parameters/returns are unnamed:
* **Implicit Fields:** `name` (no type specified).
* **Unnamed Parameters:** `Method(int)` (unnamed parameter of type `int`).
* **Unnamed Returns:** `Method() error` (unnamed return of type `error`).

**Implementation:**
* Fields can hold a `nil` type reference (`*GoTypeRef(nil)`), which generators render as `any` or `interface{}`.
* Method parameters and returns are mapped to `GoParameter` with an empty `Name` field (e.g. `GoParameter{Name: "", Type: ...}`).

---

### Level 2: Simple Types
Basic primitive types and custom object/class identifiers:
* **Syntax:** `name string` or `user User`.
* **Implementation:** Modeled as a terminal `Named` ref type:
  ```go
  GoTypeRef{Typ: Named, Name: "string"}
  ```

---

### Level 3: Language-Flavored / Structured Types
Complex type layouts containing modifiers, multiple values, or namespace paths:
* **Pointers:** `*User`
* **Slices & Arrays:** `[]int`, `[4]int`
* **Qualified Names:** `context.Context`
* **Multiple Return Values:** `Method() (int, error)`
* **Named Return Values:** `Method() (res *Result, err error)`

**Implementation:**
* **Recursive Modifiers:** Type modifiers are chained as a singly-linked list via the `Base` pointer:
  ```
  *[]User  =>  Pointer -> Slice -> Named("User")
  ```
* **Consecutive Identifiers Heuristic:** In parameter and return lists, if two consecutive `IDENTIFIER` tokens are encountered in a comma-separated chunk (e.g., `res *Result`), the dialect parser interprets the first as the name and the second as the type. Because Level 1-3 types (including dot-separated qualified names) never produce consecutive identifiers, this heuristic is watertight for this scope.
* **Return Parameter Lists:** Go's return list uses `[]GoParameter` (instead of raw type refs) to natively support both named and unnamed returns.

---

## Excluded (Level 4+ / Compiler-Grade Types)

The following language constructs are intentionally excluded from the dialect parser:
* Maps (`map[string]int`)
* Channels (`chan bool`)
* Function types (`func(int) bool`)
* Variadic parameters (`...string`)
* Generic/Type parameters (`List[T]`)

**Workaround:**
If any of these types are required in a diagram, represent them as custom Level 2 types (e.g. `StringToIntMap` or `context.Context` class identifiers) rather than using inline language syntax.
