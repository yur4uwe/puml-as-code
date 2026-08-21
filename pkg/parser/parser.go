package parser

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"
)

type Parser struct {
	ast      *ast.Diagram
	stream   *tokenizer.TokenStream
	dialect  dialect.Dialect
	targetID string
}

func (p *Parser) moveToDiagStart() error {
	if p.isStartMarker() {
		return nil
	}
	// find the target diagram start marker
	for tok := p.stream.Emit(); tok.Type != tokenizer.EOF; tok = p.stream.Emit() {
		if tok.Type != tokenizer.NEWLINE {
			continue
		}
		if p.isStartMarker() {
			return nil
		}
	}

	return tokenizer.ErrUnexpectedEOF
}

func (p *Parser) Parse(input string) (*ast.Diagram, error) {
	if p.dialect == nil {
		return nil, errors.New("dialect not initialized")
	}

	p.ast = &ast.Diagram{}
	p.stream = tokenizer.NewTokenStream(input)

	var startBound ast.DiagramBound
	// Initial blockNum is -1 so that the first block is 0-indexed
	blockNum := -1
	for {
		err := p.moveToDiagStart()
		if err != nil {
			if blockNum == -1 {
				return nil, errors.New("no diagrams found")
			}
			return nil, fmt.Errorf("diagram block %s not found, file has %d blocks", p.targetID, blockNum+1)
		}

		blockNum++

		startBound, err = p.readDiagramBounds()
		if err != nil {
			return nil, err
		} else if !startBound.IsStart {
			return nil, NewParserError("Expected diagram start marker", p.stream.PeekTokenAt(0))
		}

		if p.targetID == "" {
			break
		}

		num, err := strconv.Atoi(p.targetID)
		if err == nil {
			// We have numerical ID which means order of the block in the file
			if num != blockNum {
				continue
			} else {
				break
			}
		}

		// Otherwise, we have a named ID which have to match
		if p.targetID == startBound.ID {
			break
		}
	}

	p.ast.Statements = append(p.ast.Statements, startBound)
	p.ast.Name = startBound.Name

	for {
		// End condition check should be before consuming a token to avoid swallowing '@'
		if p.isEndMarker() {
			boundLeading := p.stream.DumpCollectedTrivia()
			endBound, err := p.readDiagramBounds()
			if err != nil {
				return nil, err
			}
			if endBound.IsStart {
				return nil, NewParserError("Unexpected diagram end marker", p.stream.PeekTokenAt(0))
			}
			if endBound.Type != startBound.Type {
				return nil, NewParserError("Types of starting and ending markers don't match", p.stream.PeekTokenAt(0))
			}
			endBound.LeadingTrivia = boundLeading
			p.stream.EmitCommentToks()
			endBound.TrailingTrivia = p.stream.DumpCollectedTrivia()
			p.ast.Statements = append(p.ast.Statements, endBound)
			break
		}

		tok := p.stream.Emit()
		if tok.Type == tokenizer.EOF {
			return nil, NewParserError("Unexpected EOF", tok)
		} else if tok.Type == tokenizer.NEWLINE {
			// We can leave it like this for now
			// If the newline is relevant it will be consumed
			// by the statement parser in some function
			continue
		}

		// Here tokens that are first in line are handled
		// Parsing shuld be constructed to result in a single statement per iteration

		// Imports via !include
		// Styles via <style>
		// Keyword handling switch
		// Handle comments
		// Handle Identifiers

		stmts, err := p.parseDiagramOnlyStatement(tok)
		if err != nil {
			return nil, err
		}
		if len(stmts) > 0 {
			p.ast.Statements = append(p.ast.Statements, stmts...)
			continue
		}

		stmts, err = p.parseContainerStatement(tok)
		if err != nil {
			return nil, err
		}
		if len(stmts) > 0 {
			p.ast.Statements = append(p.ast.Statements, stmts...)
			continue
		}

	}

	if p.targetID == "" {
		// Keep searching for the next diagram start marker
		// User might not know about discarded blocks
		err := p.moveToDiagStart()
		if err == nil {
			log.Printf("warning: discarded diagram block at %s and all following blocks\n", p.stream.PeekTokenAt(0).Pos)
		}
	}

	return p.ast, nil
}

func (p *Parser) parseContainerStatement(tok tokenizer.Token) ([]ast.Statement, error) {
	if p.HasArrowOnLine() {
		rel, err := p.parseRelationship(tok)
		if err != nil {
			return nil, err
		}
		return []ast.Statement{rel}, nil
	}

	switch keyword.Classify(tok.Literal) {
	case keyword.Class,
		keyword.Interface,
		keyword.Struct,
		keyword.Abstract,
		keyword.Enum,
		keyword.Annotation,
		keyword.Record,
		keyword.Dataclass,
		keyword.Exception,
		keyword.Protocol,
		keyword.Entity:
		// Class-like Entities
		ent, err := p.parseEntity(tok)
		if err != nil {
			return nil, err
		}
		return []ast.Statement{ent}, nil
	// Containers
	case keyword.Package,
		keyword.Together,
		keyword.Folder,
		keyword.Frame,
		keyword.Rectangle,
		keyword.Cloud,
		keyword.Database,
		keyword.Namespace,
		keyword.Node:
		cont, err := p.parseContainer(tok)
		if err != nil {
			return nil, err
		}
		return []ast.Statement{cont}, nil
	case keyword.Circle, keyword.Diamond, keyword.Metaclass, keyword.Stereotype:
		return nil, errors.New("unimplemented entity keyword handling")
	// Special Keywords
	case keyword.Note:
		note, err := p.parseNote()
		if err != nil {
			return nil, err
		}
		return []ast.Statement{note}, nil
	}

	switch tok.Type {
	case tokenizer.IDENTIFIER:
		stmt, err := p.parseInlineMember(tok)
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	default:
		return nil, nil
	}
}

func (p *Parser) parseDiagramOnlyStatement(tok tokenizer.Token) ([]ast.Statement, error) {
	if tok.Type == tokenizer.LANGLE && p.stream.AssertSeq([]tokenizer.Token{amb(tokenizer.IDENTIFIER, "style"), unamb(tokenizer.RANGLE)}) {
		return p.parseStyleBlock(tok)
	}
	if tok.Type == tokenizer.EXCLAMATION {
		stmt, err := p.parseDirective(tok)
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmt}, nil
	}

	if p.HasArrowOnLine() {
		stmnt, err := p.parseRelationship(tok)
		if err != nil {
			return nil, err
		}
		return []ast.Statement{stmnt}, nil
	}

	var stmnt ast.Statement
	var err error
	switch keyword.Classify(tok.Literal) {
	// Multi-statement branches
	case keyword.Skinparam:
		return p.parseSkinparam()

	// Single-statement branches
	case keyword.Title:
		stmnt, err = p.parseTitle()
	case keyword.Hide, keyword.Show, keyword.Remove, keyword.Restore:
		stmnt, err = p.parseVisibilityCommand(tok)
	case keyword.Scale:
		stmnt, err = p.parseScale()
	case keyword.Direction:
		stmnt, err = p.parseDiagDirection(tok)
	case keyword.Set:
		stmnt, err = p.parseSetDirective()
	case keyword.Header,
		keyword.Footer,
		keyword.Legend,
		keyword.Caption,
		keyword.Newpage:
		return nil, errors.New("unimplemented layout directive handling")
	}
	if err != nil {
		return nil, err
	}
	if stmnt != nil {
		return []ast.Statement{stmnt}, nil
	}

	return nil, nil
}
