package parser

import (
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

func unamb(tok tokenizer.TokenType) tokenizer.Token {
	return tokenizer.Token{Type: tok}
}

func amb(tok tokenizer.TokenType, literal string) tokenizer.Token {
	return tokenizer.Token{Type: tok, Literal: literal}
}

func toInteger(f float64) (int, bool) {
	i := int(f)
	if float64(i) == f {
		return i, true
	}
	return 0, false
}

func (p *Parser) mapTokenToEntityKind(tok tokenizer.Token) ast.EntityKind {
	kw := keyword.Classify(tok.Literal)
	if _, ok := p.stream.TryConsumeKW(keyword.Class); kw == keyword.Abstract && ok {
		return ast.AbstractClassKind
	}

	switch kw {
	case keyword.Class:
		return ast.ClassKind
	case keyword.Abstract:
		return ast.AbstractClassKind
	case keyword.Interface:
		return ast.InterfaceKind
	case keyword.Enum:
		return ast.EnumKind
	case keyword.Entity:
		return ast.EntityClassKind
	case keyword.Struct:
		return ast.StructKind
	case keyword.Annotation:
		return ast.AnnotationKind
	case keyword.Protocol:
		return ast.ProtocolKind
	case keyword.Circle:
		return ast.CircleKind
	case keyword.Diamond:
		return ast.DiamondKind
	case keyword.Exception:
		return ast.ExceptionKind
	case keyword.Metaclass:
		return ast.MetaclassKind
	case keyword.Record:
		return ast.RecordKind
	case keyword.Dataclass:
		return ast.DataclassKind
	case keyword.Stereotype:
		return ast.StereotypeKind
	default:
		return ast.UnknownEntityKind
	}
}

func (p *Parser) mapTokenToVisibility(tok tokenizer.TokenType) ast.VisibilityKind {
	switch tok {
	case tokenizer.PLUS:
		return ast.Public
	case tokenizer.DASH:
		return ast.Private
	case tokenizer.HASH:
		return ast.Protected
	case tokenizer.TILDE:
		return ast.Package
	default:
		return ast.UnknownVisibility
	}
}

func (p *Parser) mapKeywordToContainerKind(kw keyword.KeywordKind) ast.ContainerKind {
	switch kw {
	case keyword.Package:
		return ast.PackageKind
	case keyword.Folder:
		return ast.FolderKind
	case keyword.Frame:
		return ast.FrameKind
	case keyword.Rectangle:
		return ast.RectangleKind
	case keyword.Cloud:
		return ast.CloudKind
	case keyword.Database:
		return ast.DatabaseKind
	case keyword.Node:
		return ast.NodeKind
	default:
		return ast.UnknownContainerKind
	}
}
