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
