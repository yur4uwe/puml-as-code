package dialect

import (
	"fmt"
	"strconv"
	"strings"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/tokenizer"
)

type RefType int

const (
	Pointer RefType = iota
	Slice
	Array
	Named
)

type GoTypeRef struct {
	Typ       RefType
	ArraySize int
	Name      string
	Base      *GoTypeRef
}

// String implements [ast.TypeRef].

var _ ast.TypeRef = (*GoTypeRef)(nil)

// Lang implements [ast.TypeRef].
func (g *GoTypeRef) Lang() string {
	return "go"
}

// TypeName implements [ast.TypeRef].
func (g *GoTypeRef) String() string {
	var sb strings.Builder
	var hasName bool
	for curr := g; curr != nil; curr = curr.Base {
		switch curr.Typ {
		case Slice:
			sb.WriteString("[]")
		case Pointer:
			sb.WriteString("*")
		case Array:
			sb.WriteRune('[')
			if curr.ArraySize <= 0 {
				sb.WriteRune('?')
			} else {
				sb.WriteString(strconv.Itoa(curr.ArraySize))
			}
			sb.WriteRune(']')
		case Named:
			if hasName {
				sb.WriteRune('.')
			}
			hasName = true
			sb.WriteString(curr.Name)
		}
	}
	return sb.String()
}

func NewGoDialect() Dialect {
	return GoDialect{}
}

type GoDialect struct{}

var _ Dialect = GoDialect{}

func (g GoDialect) Name() string {
	return "go"
}

func (g GoDialect) ParseField(toks []tokenizer.Token) (ast.Parameter, error) {
	// expects this structure:
	// <name> <type>
	if len(toks) < 2 {
		return ast.Parameter{}, fmt.Errorf("%w: expected at least two tokens", ErrParsingDialect)
	}

	if toks[0].Type != tokenizer.IDENTIFIER {
		return ast.Parameter{}, fmt.Errorf("%w: expected identifier for a field name, got %s", ErrParsingDialect, toks[0].Type.String())
	}
	field := ast.Parameter{Name: toks[0].Literal}

	var err error
	field.Type, err = g.parseType(toks[1:])
	if err != nil {
		return ast.Parameter{}, err
	}

	return field, nil
}

func (g GoDialect) ParseMethod(toks []tokenizer.Token) (string, []ast.TypeRef, []ast.Parameter, error) {
	// expects this structure (no 'func' keyword, no receiver):
	// <name> '(' <params>? ')' <returns>?
	if len(toks) < 3 {
		return "", nil, nil, fmt.Errorf("%w: expected at least 3 tokens for a method", ErrParsingDialect)
	}

	if toks[0].Type != tokenizer.IDENTIFIER {
		return "", nil, nil, fmt.Errorf("%w: expected identifier for a method name, got %s", ErrParsingDialect, toks[0].Type.String())
	}
	methodName := toks[0].Literal
	pos := 1

	if toks[pos].Type != tokenizer.LPAREN {
		return "", nil, nil, fmt.Errorf("%w: expected '(' after method name, got %s", ErrParsingDialect, toks[pos].Type.String())
	}
	pos++

	paramToks, pos, err := readUntilMatching(toks, pos, tokenizer.LPAREN, tokenizer.RPAREN)
	if err != nil {
		return "", nil, nil, fmt.Errorf("%w: parsing parameter list: %w", ErrParsingDialect, err)
	}

	params, err := g.parseParamList(paramToks)
	if err != nil {
		return "", nil, nil, err
	}

	var returns []ast.TypeRef
	if pos < len(toks) {
		returns, err = g.parseReturnList(toks[pos:])
		if err != nil {
			return "", nil, nil, err
		}
	}

	return methodName, returns, params, nil
}

func (g GoDialect) parseType(toks []tokenizer.Token) (ast.TypeRef, error) {
	if len(toks) == 0 {
		return ast.TypeRef{}, fmt.Errorf("%w: expected a type, got end of tokens", ErrParsingDialect)
	}

	var sb strings.Builder

	for pos := range toks {
		sb.WriteString(toks[pos].Literal)
	}

	return ast.TypeRef{Name: sb.String()}, nil
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

func (g GoDialect) parseParamList(toks []tokenizer.Token) ([]ast.Parameter, error) {
	if len(toks) == 0 {
		return nil, nil
	}

	var params []ast.Parameter
	var sameTypeAmount int
	for _, chunk := range splitTopLevelCommas(toks) {
		if len(chunk) == 1 {
			// single token, presumably a name for a Parameter
			// with same type as the next one
			params = append(params, ast.Parameter{Name: chunk[0].Literal, Type: ast.TypeRef{}})
			sameTypeAmount++
			continue
		}
		field, err := g.ParseField(chunk)
		if err != nil {
			return nil, fmt.Errorf("parsing parameter: %w", err)
		}
		for i := 0; i < sameTypeAmount; i++ {
			// update the type of the previous parameters
			params[len(params)-1-i].Type = field.Type
		}
		params = append(params, field)
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

func (g GoDialect) parseReturnList(toks []tokenizer.Token) ([]ast.TypeRef, error) {
	if len(toks) == 0 {
		return nil, nil
	}

	if toks[0].Type == tokenizer.LPAREN {
		inner, pos, err := readUntilMatching(toks, 1, tokenizer.LPAREN, tokenizer.RPAREN)
		if err != nil {
			return nil, fmt.Errorf("%w: parsing return list: %w", ErrParsingDialect, err)
		}
		if pos != len(toks) {
			return nil, fmt.Errorf("%w: unexpected trailing tokens after return list", ErrParsingDialect)
		}

		var returns []ast.TypeRef
		for _, chunk := range splitTopLevelCommas(inner) {
			t, err := g.parseType(chunk)
			if err != nil {
				return nil, err
			}
			returns = append(returns, t)
		}
		return returns, nil
	}

	t, err := g.parseType(toks)
	if err != nil {
		return nil, err
	}
	return []ast.TypeRef{t}, nil
}
