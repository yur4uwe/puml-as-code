package parser

import (
	"testing"

	"yur4uwe/pac/internal/testutil"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/parser/keyword"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

type expectedBlockState byte

const (
	expectNoBlock expectedBlockState = iota
	expectBlockClosed
	expectBlockUnclosed
)

func TestParseEntity(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		bs     expectedBlockState
		kwType keyword.KeywordKind
		want   ast.Statement
	}{
		{
			name:   "simple class",
			input:  "class MyClass",
			kwType: keyword.Class,
			bs:     expectNoBlock,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Kind:       ast.EntityClass,
			},
		},
		{
			name:   "abstract class",
			input:  "abstract class MyAbstractClass",
			kwType: keyword.Abstract,
			want: &ast.Entity{
				Identifier: identRef("MyAbstractClass"),
				Kind:       ast.EntityAbstractClass,
			},
		},
		{
			name:   "class with alias",
			input:  "class MyClass as \"MC\"",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Alias:      tokString("MC"),
				Kind:       ast.EntityClass,
			},
		},
		{
			name:   "class with stereotype",
			input:  "class MyClass <<Service>>",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Kind:       ast.EntityClass,
				Stereotype: tokStereo("Service"),
			},
		},
		{
			name:   "class with color",
			input:  "class MyClass #FF0000",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Kind:       ast.EntityClass,
				Color:      "#FF0000",
			},
		},
		{
			name:   "class with empty body",
			input:  "class MyClass {\n}",
			kwType: keyword.Class,
			bs:     expectBlockClosed,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Kind:       ast.EntityClass,
			},
		},
		{
			name:   "class def with all modifiers",
			input:  "class MyClass as \"MC\" <T> <<Database>> #FF0000",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Alias:      tokString("MC"),
				Kind:       ast.EntityClass,
				Generic:    "T",
				Stereotype: tokStereo("Database"),
				Color:      "#FF0000",
			},
		},
		{
			name:   "class with field member",
			input:  "class MyClass {\n  +field int\n}",
			kwType: keyword.Class,
			bs:     expectBlockClosed,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Kind:       ast.EntityClass,
				Members: []ast.Member{
					&ast.FieldDeclaration{
						Visibility: ast.VisibilityPublic,
						Field: &dialect.GoField{
							Name: "field",
							Type: dialect.NamedRef("int"),
						},
					},
				},
			},
		},
		{
			name:   "class with package separator",
			input:  "class net.http.Client\n",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: identRef("Client", "net", "http"),
				Kind:       ast.EntityClass,
			},
		},
		{
			name:   "sketch field inside body",
			input:  "class MyClass {\n  untypedField\n}",
			kwType: keyword.Class,
			bs:     expectBlockClosed,
			want: &ast.Entity{
				Identifier: identRef("MyClass"),
				Kind:       ast.EntityClass,
				Members: []ast.Member{
					&ast.FieldDeclaration{
						Field: &dialect.GoField{
							Name: "untypedField",
							Type: nil,
						},
					},
				},
			},
		},
		{
			name:   "class with generic",
			input:  "class List<T>",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: identRef("List"),
				Kind:       ast.EntityClass,
				Generic:    "T",
			},
		},
		{
			name:   "class with multiple generics",
			input:  "class Map<K, V>",
			kwType: keyword.Class,
			want: &ast.Entity{
				Identifier: identRef("Map"),
				Kind:       ast.EntityClass,
				Generic:    "K, V",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			require.Equal(t, tc.kwType, keyword.Classify(tok.Literal))

			got := p.parseEntity(tok)
			require.NotNil(t, got)

			switch tc.bs {
			case expectNoBlock:
				assertBlockBodyless(t, got)
			case expectBlockClosed:
				assertBlockClosed(t, got)
			case expectBlockUnclosed:
				assertBlockUnclosed(t, got)
			}

			testutil.SemanticEqual(t, tc.want, got)
		})
	}
}

func TestParseEntity_Errors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		bs        expectedBlockState
		wantDiags []expectedDiag
	}{
		{
			name:  "class with invalid alias (not a string)",
			input: "class MyClass as MC",
			bs:    expectNoBlock,
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "entity name already set",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Line: 0, Col: 18, Offset: 17},
						End:   tokenizer.Pos{Line: 0, Col: 20, Offset: 19},
					},
				},
			},
		},
		{
			name:  "invalid alias",
			input: "class MyClass as",
			bs:    expectNoBlock,
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "expected token for entity identifier or alias",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Line: 0, Col: 17, Offset: 16},
						End:   tokenizer.Pos{Line: 0, Col: 17, Offset: 16},
					},
				},
			},
		},
		{
			name:  "unclosed stereotype",
			input: "class MyClass <<Stereo",
			bs:    expectNoBlock,
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unexpected EOF",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Line: 0, Col: 15, Offset: 14},
						End:   tokenizer.Pos{Line: 0, Col: 23, Offset: 22},
					},
				},
			},
		},
		{
			name:  "unclosed body",
			input: "class MyClass {",
			bs:    expectBlockUnclosed,
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unclosed class: expected }",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Line: 0, Col: 15, Offset: 14},
						End:   tokenizer.Pos{Line: 0, Col: 16, Offset: 15},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
			}
			tok := p.stream.Emit()
			require.Equal(t, keyword.Class, keyword.Classify(tok.Literal))

			got := p.parseEntity(tok)
			require.NotNil(t, got)

			switch tc.bs {
			case expectNoBlock:
				assertBlockBodyless(t, got)
			case expectBlockClosed:
				assertBlockClosed(t, got)
			case expectBlockUnclosed:
				assertBlockUnclosed(t, got)
			}

			assertDiagsEqual(t, tc.wantDiags, p.ast.Diagnostics, got)
		})
	}
}

func TestParseEntity_Recovery(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		bs          expectedBlockState
		wantMembers int
	}{
		{
			name:        "restore to opening brace on invalid alias with block",
			input:       "class MyClass as InvalidAlias {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on invalid alias with extra tokens before block",
			input:       "class MyClass as InvalidAlias extra junk tokens {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on bad name token with block",
			input:       "class 123 {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on unexpected member in target ref with block",
			input:       "class Foo.Bar() {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on unclosed stereotype with block",
			input:       "class MyClass <<unclosed {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on unclosed generic with block",
			input:       "class MyClass <unclosed {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on missing tag identifier with block",
			input:       "class MyClass $ {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on bad alias with generic, tags, and stereotype",
			input:       "class MyClass as InvalidAlias <T> $tag <<stereo>> {\n  +field int\n}\nclass NextClass",
			bs:          expectBlockClosed,
			wantMembers: 1,
		},
		{
			name:        "restore to opening brace on bad syntax with unclosed block",
			input:       "class MyClass as InvalidAlias {\n  +field int\n",
			bs:          expectBlockUnclosed,
			wantMembers: 1,
		},
		{
			name:        "restore to newline on invalid alias without block",
			input:       "class MyClass as InvalidAlias\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on invalid alias with extra tokens without block",
			input:       "class MyClass as InvalidAlias extra junk tokens\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on bad name token without block",
			input:       "class 123 extra junk\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on unexpected member in target ref without block",
			input:       "class Foo.Bar() extra junk\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on unclosed generic without block",
			input:       "class MyClass <unclosed generic\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on unclosed stereotype without block",
			input:       "class MyClass <<unclosed stereotype\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on missing tag identifier without block",
			input:       "class MyClass $\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on invalid tag token without block",
			input:       "class MyClass $123\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
		{
			name:        "restore to newline on bad alias with generic, tags, and stereotype",
			input:       "class MyClass as InvalidAlias <T> $tag <<stereo>>\nclass NextClass",
			bs:          expectNoBlock,
			wantMembers: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			require.Equal(t, keyword.Class, keyword.Classify(tok.Literal))

			got := p.parseEntity(tok)
			require.NotNil(t, got)
			require.True(t, p.ast.HasErrors(), "expected parser to record diagnostics on bad syntax")

			switch tc.bs {
			case expectNoBlock:
				assertBlockBodyless(t, got)
				require.Empty(t, got.Members)
			case expectBlockClosed:
				assertBlockClosed(t, got)
				require.Len(t, got.Members, tc.wantMembers)
			case expectBlockUnclosed:
				assertBlockUnclosed(t, got)
				require.Len(t, got.Members, tc.wantMembers)
			}

			// Verify the parser restored its state to the opening brace or newline
			if tc.bs == expectBlockUnclosed {
				require.True(t, p.stream.AssertType(tokenizer.EOF), "expected stream to reach EOF on unclosed block")
			} else {
				// The stream should now be positioned at the newline separating statements.
				require.Equal(t, tokenizer.NEWLINE, p.stream.PeekTokenAt(0).Type, "expected next token to be NEWLINE after recovery")

				// Consume newline and verify the subsequent statement parses cleanly
				p.stream.Emit()
				nextTok := p.stream.Emit()
				require.Equal(t, keyword.Class, keyword.Classify(nextTok.Literal))
				nextEnt := p.parseEntity(nextTok)
				require.NotNil(t, nextEnt)
				require.Equal(t, "NextClass", nextEnt.Identifier.EntityName())
			}
		})
	}
}

func TestParseFieldOrMethod(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		entryType   tokenizer.TokenType
		initialName string
		want        ast.Member
		expectErr   bool
	}{
		// Terminate fields with a newline because
		// parser will only stop collection when it sees a newline
		{
			name:        "simple field",
			input:       "name int\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: &ast.FieldDeclaration{
				Field: &dialect.GoField{
					Name: "name",
					Type: dialect.NamedRef("int"),
				},
			},
		},
		{
			name:        "pointer field",
			input:       "name *MyType\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: &ast.FieldDeclaration{
				Field: &dialect.GoField{
					Name: "name",
					Type: dialect.PointerTo(dialect.NamedRef("MyType")),
				},
			},
		},
		{
			name:        "slice field",
			input:       "name []int\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: &ast.FieldDeclaration{
				Field: &dialect.GoField{
					Name: "name",
					Type: dialect.SliceOf(dialect.NamedRef("int")),
				},
			},
		},
		{
			name:        "simple method",
			input:       "Method()\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want: &ast.MethodDeclaration{
				Method: &dialect.GoMethod{
					Name: "Method",
				},
			},
		},
		{
			name:        "method with parameters and return",
			input:       "Method(a int) error\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want: &ast.MethodDeclaration{
				Method: &dialect.GoMethod{
					Name: "Method",
					Parameters: []dialect.GoParameter{
						{Name: "a", Type: dialect.NamedRef("int")},
					},
					ReturnType: []dialect.GoParameter{
						{Type: dialect.NamedRef("error")},
					},
				},
			},
		},
		{
			name:        "method defined as method modifier but no params",
			input:       "{method} Method()\n",
			entryType:   tokenizer.LBRACE,
			initialName: "",
			want: &ast.MethodDeclaration{
				Modifiers: []string{"method"},
				Method: &dialect.GoMethod{
					Name: "Method",
				},
			},
		},
		{
			name:        "both field and method modifiers",
			input:       "{field} {method} Name()\n",
			entryType:   tokenizer.LBRACE,
			initialName: "",
			want:        nil,
			expectErr:   true,
		},
		{
			name:        "sketch field without type",
			input:       "name\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want: &ast.FieldDeclaration{
				Field: &dialect.GoField{
					Name: "name",
					Type: nil,
				},
			},
			expectErr: false,
		},
		{
			name:        "invalid field (invalid syntax)",
			input:       "name : int\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "name",
			want:        nil,
			expectErr:   true,
		},
		{
			name:        "invalid method (unclosed param parenthesis)",
			input:       "Method(a int\n",
			entryType:   tokenizer.IDENTIFIER,
			initialName: "Method",
			want:        nil,
			expectErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
			}

			var entryTok tokenizer.Token
			var mod *string
			vis := ast.VisibilityUnknown

			if tc.entryType == tokenizer.LBRACE {
				m, err := p.tryReadModifier()
				require.NoError(t, err)
				mod = &m
				entryTok = tokenizer.Token{Type: tokenizer.LBRACE}
			} else {
				entryTok = p.stream.Emit()
				require.Equal(t, tc.entryType, entryTok.Type)
				if entryTok.Type == tokenizer.IDENTIFIER {
					require.Equal(t, tc.initialName, entryTok.Literal)
				}
			}

			got, err := p.parseFieldOrMethod(mod, vis, entryTok)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				testutil.SemanticEqual(t, tc.want, got)
			}
		})
	}
}

func TestParseEntityMember(t *testing.T) {
	tt := []struct {
		name      string
		input     string
		want      ast.Member
		expectErr bool
	}{
		{
			name:  "class separator dots",
			input: ".. separator ..",
			want: &ast.ClassSeparator{
				Label:     "separator",
				Type:      '.',
				TypeCount: 4,
			},
		},
		{
			name:  "class separator hyphens",
			input: "-- section --",
			want: &ast.ClassSeparator{
				Label:     "section",
				Type:      '-',
				TypeCount: 4,
			},
		},
		{
			name:  "public field member",
			input: "+field int\n",
			want: &ast.FieldDeclaration{
				Visibility: ast.VisibilityPublic,
				Field: &dialect.GoField{
					Name: "field",
					Type: dialect.NamedRef("int"),
				},
			},
		},
		{
			name:  "private method member",
			input: "-Method()\n",
			want: &ast.MethodDeclaration{
				Visibility: ast.VisibilityPrivate,
				Method: &dialect.GoMethod{
					Name: "Method",
				},
			},
		},
		{
			name:  "field with modifier",
			input: "{field} myField int\n",
			want: &ast.FieldDeclaration{
				Modifiers: []string{"field"},
				Field: &dialect.GoField{
					Name: "myField",
					Type: dialect.NamedRef("int"),
				},
			},
		},
		{
			name:      "unclosed modifier",
			input:     "{field myField int",
			expectErr: true,
		},
		{
			name:      "unexpected token in body",
			input:     "@invalid",
			expectErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
			}
			got, err := p.parseEntityMember()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				testutil.SemanticEqual(t, tc.want, got)
			}
		})
	}
}

func TestParseInlineMember(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantTargetName string
		wantMethod     bool
		wantField      bool
		wantModifiers  []string
		wantVisibility ast.VisibilityKind
	}{
		{
			name:           "field with visibility",
			input:          "Foo : +id string\n",
			wantTargetName: "Foo",
			wantField:      true,
			wantVisibility: ast.VisibilityPublic,
		},
		{
			name:           "method with visibility and params",
			input:          "Foo : +DoWork(ctx Context) error\n",
			wantTargetName: "Foo",
			wantMethod:     true,
			wantVisibility: ast.VisibilityPublic,
		},
		{
			name:           "field with static modifier",
			input:          "Foo : {static} count int\n",
			wantTargetName: "Foo",
			wantField:      true,
			wantModifiers: []string{
				"static",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			for tok.Type == tokenizer.NEWLINE {
				tok = p.stream.Emit()
			}
			got, err := p.parseInlineMember(tok)
			require.NoError(t, err, "unexpected error: %s", err)
			member, ok := got.(*ast.InlineMember)
			require.True(t, ok, "expected inline member but got %T", got)
			require.Equal(t, tc.wantTargetName, member.Target.EntityName(), "expected entity %q but got %q", tc.wantTargetName, member.Target.EntityName())
			if tc.wantMethod {
				method, ok := member.Member.(*ast.MethodDeclaration)
				require.True(t, ok)
				require.Equal(t, tc.wantModifiers, method.Modifiers)
				require.Equal(t, tc.wantVisibility, method.Visibility)
			} else {
				field, ok := member.Member.(*ast.FieldDeclaration)
				require.True(t, ok)
				require.Equal(t, tc.wantModifiers, field.Modifiers)
				require.Equal(t, tc.wantVisibility, field.Visibility)
			}
		})
	}
}

func TestParseEntityAndContainerTags(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStereo *ast.Stereotype
		wantTags   []string
	}{
		{
			name:       "pre-stereotype tags",
			input:      "class Foo $tag1 $tag2 <<stereo>>\n",
			wantStereo: tokStereo("stereo"),
			wantTags:   []string{"tag1", "tag2"},
		},
		{
			name:       "post-stereotype tags",
			input:      "class Foo <<stereo>> $tag3\n",
			wantStereo: tokStereo("stereo"),
			wantTags:   []string{"tag3"},
		},
		{
			name:       "precedence: pre-stereotype tags override post-stereotype tags",
			input:      "class Foo $pre <<stereo>> $post\n",
			wantStereo: tokStereo("stereo"),
			wantTags:   []string{"pre"},
		},
		{
			name:       "standalone tag without stereotype",
			input:      "class Foo $alone\n",
			wantStereo: nil,
			wantTags:   []string{"alone"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			got := p.parseEntity(tok)
			require.NotNil(t, got)
			testutil.SemanticEqual(t, tc.wantStereo, got.Stereotype)
			require.Equal(t, tc.wantTags, got.Tags)
		})
	}
}
