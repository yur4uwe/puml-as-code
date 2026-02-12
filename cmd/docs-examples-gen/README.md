# Docs Examples Generator

Generates tokenizer example test cases from the `<plantuml>` blocks in the class-diagram docs.

## Usage

```zsh
go run ./cmd/docs-examples-gen -docs-dir docs/class-diagram -out pkg/tokenizer/examples_generated_testdata.go
```

After generation, run tokenizer tests:

```zsh
go test ./pkg/tokenizer -run Examples
```
