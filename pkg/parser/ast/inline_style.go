package ast

import "yur4uwe/pac/pkg/tokenizer"

//go:generate enumer -type=GradientKind -transform=lower -trimprefix=Gradient -json
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

//go:generate enumer -type=LineStyle -transform=lower -trimprefix=Line -json
type LineStyle byte

const (
	LinePlain LineStyle = iota
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

func (s *ColorRef) Span() tokenizer.SourceSpan {
	if s == nil {
		return tokenizer.SourceSpan{}
	}
	return s.ColorSpan
}
func (s *ColorRef) colorExpr() {}

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
	// There should be a non-nil guarantee
	startPos = g.Start.Span().Start
	endPos = g.End.Span().End
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

func (e *EntityStyle) IsEmpty() bool {
	if e == nil {
		return true
	}
	return e.Background == nil && e.Header == nil && e.LineColor == nil && e.LineStyle == LinePlain && e.TextColor == nil
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

func (c *ContainerStyle) IsEmpty() bool {
	if c == nil {
		return true
	}
	return c.Background == nil && c.LineColor == nil && c.LineStyle == LinePlain && c.TextColor == nil
}

// ArrowStyle models bracketed attribute annotations directly on arrows:
//
//	-[dashed,#red,thickness=2]->
type ArrowStyle struct {
	Span      tokenizer.SourceSpan
	LineColor *ColorRef `json:",omitempty"`
	LineStyle LineStyle `json:",omitempty"`
	Thickness int       `json:",omitempty"`
}

func (a *ArrowStyle) IsEmpty() bool {
	if a == nil {
		return true
	}
	return a.LineColor == nil && a.LineStyle == LinePlain && a.Thickness == 0
}

// RelationshipStyle models post-target inline style annotations on relationships:
//
//	#line:color;line.[bold|dashed|dotted];text:color
type RelationshipStyle struct {
	Span      tokenizer.SourceSpan
	LineColor ColorExpr `json:",omitempty"`
	LineStyle LineStyle `json:",omitempty"`
	TextColor ColorExpr `json:",omitempty"`
}

func (r *RelationshipStyle) IsEmpty() bool {
	if r == nil {
		return true
	}
	return r.LineColor == nil && r.LineStyle == LinePlain && r.TextColor == nil
}

// NoteStyle models inline style annotations on notes:
//
//	#color
type NoteStyle struct {
	Span       tokenizer.SourceSpan
	Background ColorExpr `json:",omitempty"`
}
