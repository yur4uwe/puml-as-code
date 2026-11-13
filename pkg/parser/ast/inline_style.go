package ast

import "yur4uwe/pac/pkg/tokenizer"

type GradientKind byte

const (
	GradientNone       GradientKind = iota
	GradientHorizontal              // -
	GradientVertical                // |
	GradientSlash                   // /
	GradientBackslash               // \
)

func FromTokType(typ tokenizer.TokenType) GradientKind {
	switch typ {
	case tokenizer.DASH:
		return GradientHorizontal
	case tokenizer.PIPE:
		return GradientVertical
	case tokenizer.SLASH:
		return GradientSlash
	case tokenizer.BACKSLASH:
		return GradientBackslash
	default:
		return GradientNone
	}
}

type LineStyle byte

const (
	LineNormal LineStyle = iota
	LineBold
	LineDashed
	LineDotted
)

// ColorExpr represents any valid color expression in PlantUML.
// Sealed interface implemented by *ColorRef and *GradientColor.
type ColorExpr interface {
	Span() tokenizer.SourceSpan
	colorExpr()
}

type ColorRef struct {
	Literal   string
	ColorSpan tokenizer.SourceSpan
}

func (s *ColorRef) Span() tokenizer.SourceSpan { return s.ColorSpan }
func (s *ColorRef) colorExpr()                 {}

func FromToken(tok tokenizer.Token) *ColorRef {
	return &ColorRef{
		Literal:   tok.Literal,
		ColorSpan: tok.Span,
	}
}

// GradientColor represents a two-stop color gradient: Start <dir> End.
type GradientColor struct {
	Start     *ColorRef
	End       *ColorRef
	Direction GradientKind
}

func (g *GradientColor) Span() tokenizer.SourceSpan {
	if g == nil {
		return tokenizer.SourceSpan{}
	}
	var startPos, endPos tokenizer.Pos
	if g.Start != nil {
		startPos = g.Start.Span().Start
	}
	if g.End != nil {
		endPos = g.End.Span().End
	}
	return tokenizer.SourceSpan{Start: startPos, End: endPos}
}

func (g *GradientColor) colorExpr() {}

// EntityStyle models inline style annotations on entities:
//
//	#[color|back:color];header:color;line:color;line.[bold|dashed|dotted];text:color ##[lineStyle]lineColor
type EntityStyle struct {
	Span       tokenizer.SourceSpan
	Background ColorExpr `json:",omitempty"`
	Header     ColorExpr `json:",omitempty"`
	LineColor  ColorExpr `json:",omitempty"`
	LineStyle  LineStyle `json:",omitempty"`
	TextColor  ColorExpr `json:",omitempty"`
}

// ContainerStyle models inline style annotations on packages/containers:
//
//	#[color|back:color];line:color;line.[bold|dashed|dotted];text:color
type ContainerStyle struct {
	Span       tokenizer.SourceSpan
	Background ColorExpr `json:",omitempty"`
	LineColor  ColorExpr `json:",omitempty"`
	LineStyle  LineStyle `json:",omitempty"`
	TextColor  ColorExpr `json:",omitempty"`
}

// RelationshipStyle models inline style annotations on relationships/arrows:
//
//	#line:color;line.[bold|dashed|dotted];text:color
type RelationshipStyle struct {
	Span      tokenizer.SourceSpan
	LineColor ColorExpr `json:",omitempty"`
	LineStyle LineStyle `json:",omitempty"`
	TextColor ColorExpr `json:",omitempty"`
	Thickness int       `json:",omitempty"`
}

// NoteStyle models inline style annotations on notes:
//
//	#color
type NoteStyle struct {
	Span       tokenizer.SourceSpan
	Background ColorExpr `json:",omitempty"`
}
