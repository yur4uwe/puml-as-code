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
		return ast.EntityAbstractClass
	}

	switch kw {
	case keyword.Class:
		return ast.EntityClass
	case keyword.Abstract:
		return ast.EntityAbstractClass
	case keyword.Interface:
		return ast.EntityInterface
	case keyword.Enum:
		return ast.EntityEnum
	case keyword.Entity:
		return ast.EntityEntityClass
	case keyword.Struct:
		return ast.EntityStruct
	case keyword.Annotation:
		return ast.EntityAnnotation
	case keyword.Protocol:
		return ast.EntityProtocol
	case keyword.Circle:
		return ast.EntityCircle
	case keyword.Diamond:
		return ast.EntityDiamond
	case keyword.Exception:
		return ast.EntityException
	case keyword.Metaclass:
		return ast.EntityMetaclass
	case keyword.Record:
		return ast.EntityRecord
	case keyword.Dataclass:
		return ast.EntityDataclass
	case keyword.Stereotype:
		return ast.EntityStereotype
	default:
		return ast.EntityUnknown
	}
}

func (p *Parser) mapTokenToVisibility(tok tokenizer.TokenType) ast.VisibilityKind {
	switch tok {
	case tokenizer.PLUS:
		return ast.VisibilityPublic
	case tokenizer.DASH:
		return ast.VisibilityPrivate
	case tokenizer.HASH:
		return ast.VisibilityProtected
	case tokenizer.TILDE:
		return ast.VisibilityPackage
	default:
		return ast.VisibilityUnknown
	}
}

func (p *Parser) mapKeywordToContainerKind(kw keyword.KeywordKind) ast.ContainerKind {
	switch kw {
	case keyword.Package:
		return ast.ContainerPackage
	case keyword.Folder:
		return ast.ContainerFolder
	case keyword.Frame:
		return ast.ContainerFrame
	case keyword.Rectangle:
		return ast.ContainerRectangle
	case keyword.Cloud:
		return ast.ContainerCloud
	case keyword.Database:
		return ast.ContainerDatabase
	case keyword.Node:
		return ast.ContainerNode
	case keyword.Namespace:
		return ast.ContainerNamespace
	case keyword.Together:
		return ast.ContainerTogether
	default:
		return ast.ContainerUnknown
	}
}
