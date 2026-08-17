package dialect

import (
	"fmt"
	"strconv"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

func (g GoDialect) parseField(toks []tokenizer.Token, visibility ast.VisibilityKind, modifiers []string) (*GoField, error) {
	// expects this structure:
	// <name> <type>
	if len(toks) < 2 {
		return nil, fmt.Errorf("%w: expected at least two tokens", ErrParsingDialect)
	}

	if toks[0].Type != tokenizer.IDENTIFIER {
		return nil, fmt.Errorf("%w: expected identifier for a field name, got %s", ErrParsingDialect, toks[0].Type.String())
	}
	field := &GoField{
		Name:       toks[0].Literal,
		Visibility: visibility,
		Modifiers:  modifiers,
	}

	var err error
	field.Type, err = g.parseType(toks[1:])
	if err != nil {
		return nil, err
	}

	return field, nil
}

func (g GoDialect) parseMethod(toks []tokenizer.Token, visibility ast.VisibilityKind, modifiers []string) (*GoMethod, error) {
	// expects this structure (no 'func' keyword, no receiver):
	// <name> '(' <params>? ')' <returns>?
	if len(toks) < 3 {
		return nil, fmt.Errorf("%w: expected at least 3 tokens for a method", ErrParsingDialect)
	}

	if toks[0].Type != tokenizer.IDENTIFIER {
		return nil, fmt.Errorf("%w: expected identifier for a method name, got %s", ErrParsingDialect, toks[0].Type.String())
	}
	methodName := toks[0].Literal
	pos := 1

	if toks[pos].Type != tokenizer.LPAREN {
		return nil, fmt.Errorf("%w: expected '(' after method name, got %s", ErrParsingDialect, toks[pos].Type.String())
	}
	pos++

	paramToks, pos, err := readUntilMatching(toks, pos, tokenizer.LPAREN, tokenizer.RPAREN)
	if err != nil {
		return nil, fmt.Errorf("%w: parsing parameter list: %w", ErrParsingDialect, err)
	}

	params, err := g.parseParamList(paramToks)
	if err != nil {
		return nil, err
	}

	var returns []GoParameter
	if pos < len(toks) {
		returns, err = g.parseReturnList(toks[pos:])
		if err != nil {
			return nil, err
		}
	}

	return &GoMethod{
		Name:       methodName,
		ReturnType: returns,
		Parameters: params,
		Modifiers:  modifiers,
		Visibility: visibility,
	}, nil
}

func (g GoDialect) parseType(toks []tokenizer.Token) (*GoTypeRef, error) {
	ref, consumed, err := g.parseTypeFrom(toks, 0)
	if err != nil {
		return nil, err
	}
	if consumed != len(toks) {
		return nil, fmt.Errorf("%w: unexpected trailing tokens in type",
			ErrParsingDialect)
	}
	return &ref, nil
}

func (g GoDialect) parseTypeFrom(toks []tokenizer.Token, pos int) (GoTypeRef, int,
	error,
) {
	if pos >= len(toks) {
		return GoTypeRef{}, pos, fmt.Errorf("%w: expected type, got end of tokens",
			ErrParsingDialect)
	}

	switch toks[pos].Type {
	case tokenizer.ASTERISK: // *T
		base, newPos, err := g.parseTypeFrom(toks, pos+1)
		if err != nil {
			return GoTypeRef{}, 0, err
		}
		return GoTypeRef{Typ: KindPointer, Base: &base}, newPos, nil

	case tokenizer.LBRACKET: // []T or [N]T
		pos++
		if pos >= len(toks) {
			return GoTypeRef{}, 0, fmt.Errorf("%w: unexpected end after '['",
				ErrParsingDialect)
		}
		if toks[pos].Type == tokenizer.RBRACKET {
			// []T — slice
			base, newPos, err := g.parseTypeFrom(toks, pos+1)
			if err != nil {
				return GoTypeRef{}, 0, err
			}
			return GoTypeRef{Typ: KindSlice, Base: &base}, newPos, nil
		}
		if toks[pos].Type == tokenizer.NUMBER {
			size, _ := strconv.Atoi(toks[pos].Literal)
			pos++
			if pos >= len(toks) || toks[pos].Type != tokenizer.RBRACKET {
				return GoTypeRef{}, 0, fmt.Errorf("%w: expected ']' after array size",
					ErrParsingDialect)
			}
			base, newPos, err := g.parseTypeFrom(toks, pos+1)
			if err != nil {
				return GoTypeRef{}, 0, err
			}
			return GoTypeRef{Typ: KindArray, ArraySize: size, Base: &base}, newPos, nil
		}
		return GoTypeRef{}, 0, fmt.Errorf("%w: expected ']' or number after '['",
			ErrParsingDialect)

	case tokenizer.IDENTIFIER: // named type, possibly qualified (pkg.Type)
		ref := GoTypeRef{Typ: KindNamed, Name: toks[pos].Literal}
		pos++
		if pos < len(toks) && toks[pos].Type == tokenizer.DOT {
			pos++ // skip dot
			if pos >= len(toks) || toks[pos].Type != tokenizer.IDENTIFIER {
				return GoTypeRef{}, 0, fmt.Errorf("%w: expected identifier after '.'",
					ErrParsingDialect)
			}
			qualified := GoTypeRef{Typ: KindNamed, Name: toks[pos].Literal}
			ref.Base = &qualified
			pos++
		}
		return ref, pos, nil

	default:
		return GoTypeRef{}, 0, fmt.Errorf("%w: unexpected token '%s' in type",
			ErrParsingDialect, toks[pos].Literal)
	}
}

// readUntilMatching returns the tokens strictly between the opener already
// consumed at pos-1 and its matching closer, plus the position right after the closer.
func readUntilMatching(toks []tokenizer.Token, pos int, open, close tokenizer.TokenType) ([]tokenizer.Token, int, error) {
	depth := 1
	start := pos
	for pos < len(toks) {
		switch toks[pos].Type {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return toks[start:pos], pos + 1, nil
			}
		}
		pos++
	}
	return nil, 0, fmt.Errorf("%w: unmatched '%s'", ErrParsingDialect, open.String())
}

func (g GoDialect) parseParamList(toks []tokenizer.Token) ([]GoParameter, error) {
	if len(toks) == 0 {
		return nil, nil
	}

	var params []GoParameter
	var sameTypeAmount int
	for _, chunk := range splitTopLevelCommas(toks) {
		if len(chunk) == 1 {
			// single token, presumably a name for a Parameter
			// with same type as the next one
			params = append(params, GoParameter{Name: chunk[0].Literal, Type: nil})
			sameTypeAmount++
			continue
		}
		field, err := g.parseField(chunk, ast.VisibilityUnknown, nil)
		if err != nil {
			return nil, fmt.Errorf("parsing parameter: %w", err)
		}
		for i := 0; i < sameTypeAmount; i++ {
			// update the type of the previous parameters
			params[len(params)-1-i].Type = field.Type
		}
		sameTypeAmount = 0
		params = append(params, GoParameter{Name: field.Name, Type: field.Type})
	}
	return params, nil
}

func splitTopLevelCommas(toks []tokenizer.Token) [][]tokenizer.Token {
	var chunks [][]tokenizer.Token
	depth := 0
	start := 0
	for i, t := range toks {
		switch t.Type {
		case tokenizer.LPAREN, tokenizer.LBRACKET:
			depth++
		case tokenizer.RPAREN, tokenizer.RBRACKET:
			depth--
		case tokenizer.COMMA:
			if depth == 0 {
				chunks = append(chunks, toks[start:i])
				start = i + 1
			}
		}
	}
	chunks = append(chunks, toks[start:])
	return chunks
}

func (g GoDialect) parseReturnList(toks []tokenizer.Token) ([]GoParameter, error) {
	if len(toks) == 0 {
		return nil, nil
	}

	if toks[0].Type != tokenizer.LPAREN {
		t, err := g.parseType(toks)
		if err != nil {
			return nil, err
		}
		return []GoParameter{{Type: t}}, nil
	}

	inner, pos, err := readUntilMatching(toks, 1, tokenizer.LPAREN, tokenizer.RPAREN)
	if err != nil {
		return nil, fmt.Errorf("%w: parsing return list: %w", ErrParsingDialect, err)
	}
	if pos != len(toks) {
		return nil, fmt.Errorf("%w: unexpected trailing tokens after return list", ErrParsingDialect)
	}

	chunks := splitTopLevelCommas(inner)

	// Detect named vs unnamed returns.
	// We only need to check the first chunk because Go syntax rules mandate that
	// either ALL return values are named or ALL are unnamed.
	var isNamed bool
	if len(chunks[0]) >= 2 && chunks[0][0].Type == tokenizer.IDENTIFIER {
		switch chunks[0][1].Type {
		case tokenizer.IDENTIFIER, tokenizer.LBRACKET, tokenizer.ASTERISK:
			isNamed = true
		default:
			isNamed = false
		}
	}

	var returns []GoParameter
	for _, chunk := range chunks {
		if isNamed {
			// Parse as name + type (reuse parseField logic)
			field, err := g.parseField(chunk, ast.VisibilityUnknown, nil)
			if err != nil {
				return nil, err
			}
			returns = append(returns, GoParameter{Name: field.Name, Type: field.Type})
		} else {
			// Parse as just a type
			t, err := g.parseType(chunk)
			if err != nil {
				return nil, err
			}
			returns = append(returns, GoParameter{Type: t})
		}
	}

	return returns, nil
}
