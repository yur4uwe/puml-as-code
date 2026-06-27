package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func unamb(tok tokenizer.TokenType) tokenizer.Token {
	return tokenizer.Token{Type: tok}
}

func amb(tok tokenizer.TokenType, literal string) tokenizer.Token {
	return tokenizer.Token{Type: tok, Literal: literal}
}

func (p *Parser) parseCommand(tok tokenizer.Token) (any, error) {
	switch tok.Type {
	case tokenizer.SET_CMD:
	case tokenizer.SCALE:
	case tokenizer.EXCLAMATION:
	}
	return nil, nil
}

func toInteger(f float64) (int, bool) {
	i := int(f)
	if float64(i) == f {
		return i, true
	}
	return 0, false
}

func (p *Parser) mapTokenToEntityKind(tok tokenizer.TokenType) ast.EntityKind {
	if _, ok := p.stream.TryConsumeType(tokenizer.CLASS); tok == tokenizer.ABSTRACT && ok {
		return ast.AbstractClassKind
	}

	switch tok {
	case tokenizer.CLASS:
		return ast.ClassKind
	case tokenizer.ABSTRACT:
		return ast.AbstractClassKind
	case tokenizer.INTERFACE:
		return ast.InterfaceKind
	case tokenizer.ENUM:
		return ast.EnumKind
	case tokenizer.ENTITY:
		return ast.EntityClassKind
	case tokenizer.STRUCT:
		return ast.StructKind
	case tokenizer.ANNOTATION:
		return ast.AnnotationKind
	case tokenizer.PROTOCOL:
		return ast.ProtocolKind
	case tokenizer.CIRCLE:
		return ast.CircleKind
	case tokenizer.DIAMOND:
		return ast.DiamondKind
	case tokenizer.EXCEPTION:
		return ast.ExceptionKind
	case tokenizer.METACLASS:
		return ast.MetaclassKind
	case tokenizer.RECORD:
		return ast.RecordKind
	case tokenizer.DATACLASS:
		return ast.DataclassKind
	case tokenizer.STEREOTYPE:
		return ast.StereotypeKind
	default:
		return ast.UnknownEntityKind
	}
}

func (p *Parser) mapTokenToVisibility(tok tokenizer.TokenType) ast.VisibilityKind {
	switch tok {
	case tokenizer.PLUS:
		return ast.Public
	case tokenizer.HYPHEN:
		return ast.Private
	case tokenizer.HASH:
		return ast.Protected
	case tokenizer.TILDE:
		return ast.Package
	default:
		return ast.UnknownVisibility
	}
}
