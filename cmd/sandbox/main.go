package main

import (
	"fmt"

	"yur4uwe/pac/pkg/tokenizer"
)

func main() {
	l := tokenizer.NewLexer(`@startuml
class A
@enduml`)

	for tok := l.Emit(); ; tok = l.Emit() {
		fmt.Printf("Token: %s(%s) at %v\n", tok.Type, tok.Literal, tok.Span)
		if tok.Type == tokenizer.EOF {
			break
		}
	}
}
