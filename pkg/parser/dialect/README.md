# Dialect

## What is a Dialect?

Dialect is target-language-specific syntax for field and method declarations in PlantUML diagrams that allows converting ambiguous and unstructured PlantUML member declarations into language-specific structures for downstream code generation. Dialects are strongly tied to the programming language they inherit syntax from*.

\* This can be false, but i struggle to imagine an ability to reuse dialects across languages.

Because each programming language has distinct type syntax and idioms (e.g. pointers, slices, nullability, return signatures), each target language requires its own dedicated Dialect implementation.

## Example

### Field declaration

```@startuml
class Foo {
  +id string
}
```

Field declaration in the diagram above follow go's field syntax - <name> <type>. Encapsulation symbol ('+') is consumed by parser and not dialect, and so are modifiers (`{static}`, `{abstract}`). So for a dialect parser an 'id' field looks like: "id string"

### Method declaration

```@startuml
class Foo {
  +GetId() string
}
```

Method declaration in above diagram follow go's method syntax - <name> (<parameters>) <return types>. Everything about field parsing is also true about method parsing

## Contract

Dialects must implement the following contract:

1. Dialect parser must implement `Dialect` interface.
2. Parser must understand and parse at least 70% of target language's field and method declaration syntax to cover basic user needs for representing intention via diagrams

## Adding a target language

Start with one small diagram that exercises a field, a method, and a relationship.
Record the expected generated files before extending the grammar. Go is the
reference backend; `lax` is for formatting and does not expose enough type
information for generation.

1. **Parse members.** Implement [`Dialect`](iface.go) under this package and add
   a case to [`Factory`](factory.go). Implement `Name`, `ParseField` and
   `ParseMethod`. Return target-specific values that satisfy
   [`ast.Field` and `ast.Method`](../ast/class_member.go), including the member
   name and a source-like `String` representation. The host parser consumes
   visibility, modifiers, and statement boundaries. Keep the dialect focused on
   the member tokens it receives. Wrap malformed input with `ErrParsingDialect`;
   leave package lookup and I/O to later stages.
2. **Generate files.** Add a target package alongside the
   [Go generator](../../generator/go). Implement `SemanticPass` against the
   resolved symbol table and `GenerateFromClassDiagram` to return relative paths
   and bytes. Keep file writes in `pac gen`. Register the backend with the same
   name in [`CodeGeneratorByLang`](../../generator/factory.go). The current
   generator interface uses `go.GeneratedFile` as a shared file envelope.
3. **Test both boundaries.** Add member tests for ordinary types, methods with
   arguments or return values, and malformed input. Include a `.puml` and a
   golden output fixture in an integration test for the new backend, using the
   [Go golden tests](../../generator/go/integration_test.go) as a model. Check
   that `pac gen -lang <name> -in <diagram> -outd <directory>` produces the
   expected files and run `go test ./...`.

When asking a coding agent to add a backend, give it the exact member syntax,
relationship mapping, a small expected output, and the commands above. Start
with target-specific code. Change the shared AST only when the new information
belongs to the diagram itself and is useful across backends.
