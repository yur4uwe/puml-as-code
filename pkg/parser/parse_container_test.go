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

func TestParseContainer(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		kwType keyword.KeywordKind
		want   ast.Container
	}{
		{
			name:   "together empty newline",
			input:  "together\n",
			kwType: keyword.Together,
			want: ast.Container{
				Kind: ast.ContainerTogether,
			},
		},
		{
			name:   "together empty body",
			input:  "together {}",
			kwType: keyword.Together,
			want: ast.Container{
				Kind: ast.ContainerTogether,
			},
		},
		{
			name:   "together with single class",
			input:  "together { class A }",
			kwType: keyword.Together,
			want: ast.Container{
				Kind: ast.ContainerTogether,
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "package simple identifier",
			input:  "package mypkg\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("mypkg"),
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package string identifier and alias",
			input:  `package "My Package" as mypkg` + "\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("mypkg"),
				Alias:      tokString("My Package"),
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package identifier and string alias",
			input:  `package mypkg as "My Package"` + "\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("mypkg"),
				Alias:      tokString("My Package"),
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package identical identifier and alias",
			input:  "package mypkg as otherpkg\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("otherpkg"),
				Alias:      tokIdent("mypkg"),
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package with stereotype",
			input:  "package mypkg <<Service>>\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("mypkg"),
				Stereotype: tokStereo("Service"),
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package with stereotype and color",
			input:  "package mypkg <<Service>> #green\n",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("mypkg"),
				Stereotype: tokStereo("Service"),
				Color:      "#green",
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package body with relationship and nested package",
			input:  "package mypkg { class A together { class B } A -> B\n}",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("mypkg"),
				Kind:       ast.ContainerPackage,
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
					&ast.Container{
						Kind: ast.ContainerTogether,
						Statements: []ast.Statement{
							&ast.Entity{
								Identifier: identRef("B"),
								Kind:       ast.EntityClass,
							},
						},
					},
					&ast.Relationship{
						LHS:       identRef("A"),
						RHS:       identRef("B"),
						Body:      '-',
						RArrow:    '>',
						BodyCount: 1,
						TypeRHS:   ast.RelationAssociation,
					},
				},
			},
		},
		{
			name:   "package with the keyword as a name in relationship",
			input:  "package p { class folder {} folder --> p }",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("p"),
				Kind:       ast.ContainerPackage,
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("folder"),
						Kind:       ast.EntityClass,
					},
					&ast.Relationship{
						LHS:       identRef("folder"),
						RHS:       identRef("p"),
						Body:      '-',
						BodyCount: 2,
						RArrow:    '>',
						TypeRHS:   ast.RelationAssociation,
					},
				},
			},
		},
		{
			name:   "expect to correctly parse nested containers",
			input:  "package p.p {}",
			kwType: keyword.Package,
			want: ast.Container{
				Identifier: identRef("p", "p"),
				Kind:       ast.ContainerPackage,
			},
		},
		{
			name:   "package with same line body and color",
			input:  "package mypkg #red { class A }",
			kwType: keyword.Package,
			want: ast.Container{
				Kind:       ast.ContainerPackage,
				Identifier: identRef("mypkg"),
				Color:      "#red",
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		// --- Container keyword coverage ---
		{
			name:   "folder with body",
			input:  "folder myfolder { class A }",
			kwType: keyword.Folder,
			want: ast.Container{
				Kind:       ast.ContainerFolder,
				Identifier: identRef("myfolder"),
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "frame with body",
			input:  "frame myframe { class A }",
			kwType: keyword.Frame,
			want: ast.Container{
				Kind:       ast.ContainerFrame,
				Identifier: identRef("myframe"),
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "rectangle with body",
			input:  "rectangle myrect { class A }",
			kwType: keyword.Rectangle,
			want: ast.Container{
				Kind:       ast.ContainerRectangle,
				Identifier: identRef("myrect"),
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "cloud with body",
			input:  "cloud mycloud { class A }",
			kwType: keyword.Cloud,
			want: ast.Container{
				Kind:       ast.ContainerCloud,
				Identifier: identRef("mycloud"),
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "database with body",
			input:  "database mydb { class A }",
			kwType: keyword.Database,
			want: ast.Container{
				Kind:       ast.ContainerDatabase,
				Identifier: identRef("mydb"),
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "node with body",
			input:  "node mynode { class A }",
			kwType: keyword.Node,
			want: ast.Container{
				Kind:       ast.ContainerNode,
				Identifier: identRef("mynode"),
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "namespace with body",
			input:  "namespace myns { class A }",
			kwType: keyword.Namespace,
			want: ast.Container{
				Kind:       ast.ContainerNamespace,
				Identifier: identRef("myns"),
				Statements: []ast.Statement{
					&ast.Entity{
						Identifier: identRef("A"),
						Kind:       ast.EntityClass,
					},
				},
			},
		},
		{
			name:   "folder empty body",
			input:  "folder myfolder {}",
			kwType: keyword.Folder,
			want: ast.Container{
				Kind:       ast.ContainerFolder,
				Identifier: identRef("myfolder"),
			},
		},
		{
			name:   "namespace with alias and stereotype",
			input:  `namespace myns as "My Namespace" <<API>>` + "\n",
			kwType: keyword.Namespace,
			want: ast.Container{
				Kind:       ast.ContainerNamespace,
				Identifier: identRef("myns"),
				Alias:      tokString("My Namespace"),
				Stereotype: tokStereo("API"),
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

			got := p.parseContainer(tok)
			require.NotNil(t, got)
			testutil.SemanticEqual(t, tc.want, got)
		})
	}
}

func TestParseContainer_Errors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		kwType      keyword.KeywordKind
		bs          expectedBlockState
		errContains string
	}{
		{
			name:        "package incomplete alias",
			input:       "package mypkg as\n",
			kwType:      keyword.Package,
			bs:          expectNoBlock,
			errContains: "Expected identifier or string for container alias",
		},
		{
			name:        "package unclosed stereotype",
			input:       "package mypkg <<Service\n",
			kwType:      keyword.Package,
			bs:          expectNoBlock,
			errContains: "unexpected EOF",
		},
		{
			name:        "package unclosed body",
			input:       "package mypkg { class A",
			kwType:      keyword.Package,
			bs:          expectBlockUnclosed,
			errContains: "unclosed package: expected }",
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

			got := p.parseContainer(tok)
			require.NotNil(t, got)

			switch tc.bs {
			case expectNoBlock:
				assertBlockBodyless(t, got)
			case expectBlockClosed:
				assertBlockClosed(t, got)
			case expectBlockUnclosed:
				assertBlockUnclosed(t, got)
			}

			require.True(t, p.ast.HasErrors(), "expected parser to record diagnostics on bad syntax")
			require.NotEmpty(t, p.ast.Diagnostics)
			require.Contains(t, p.ast.Diagnostics[0].Message, tc.errContains)
		})
	}
}

func TestParseContainer_Recovery(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		kwType         keyword.KeywordKind
		bs             expectedBlockState
		wantStatements int
	}{
		// 1. Invalid identifier
		{
			name:           "restore to opening brace on bad name token with block",
			input:          "package 123 {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on bad name token without block",
			input:          "package 123\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},
		{
			name:           "restore to newline on bad name token with extra junk without block",
			input:          "package 123 extra junk\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},

		// 2. Invalid alias
		{
			name:           "restore to opening brace on invalid alias with block",
			input:          "package mypkg as 123 {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to opening brace on invalid alias with extra tokens before block",
			input:          "package mypkg as 123 extra junk tokens {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on invalid alias without block",
			input:          "package mypkg as 123\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},
		{
			name:           "restore to newline on invalid alias with extra tokens without block",
			input:          "package mypkg as 123 extra junk tokens\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},

		// 3. Extra junk before block
		{
			name:           "restore to opening brace on extra junk before block",
			input:          "package mypkg as \"alias\" extra junk {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on extra junk without block",
			input:          "package mypkg as \"alias\" extra junk\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},

		// 4. Unexpected member in target ref
		{
			name:           "restore to opening brace on unexpected member in target ref with block",
			input:          "package Foo.Bar() {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on unexpected member in target ref without block",
			input:          "package Foo.Bar() extra junk\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},

		// 5. Unclosed stereotype
		{
			name:           "restore to opening brace on unclosed stereotype with block",
			input:          "package mypkg <<unclosed {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on unclosed stereotype without block",
			input:          "package mypkg <<unclosed stereotype\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},

		// 6. Malformed tags
		{
			name:           "restore to opening brace on missing tag identifier with block",
			input:          "package mypkg $ {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on missing tag identifier without block",
			input:          "package mypkg $\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},
		{
			name:           "restore to opening brace on invalid tag token with block",
			input:          "package mypkg $123 {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on invalid tag token without block",
			input:          "package mypkg $123\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},

		// 7. Bad alias with tags and stereotype
		{
			name:           "restore to opening brace on bad alias with tags and stereotype",
			input:          "package mypkg as 123 $tag <<stereo>> {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on bad alias with tags and stereotype",
			input:          "package mypkg as 123 $tag <<stereo>>\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectNoBlock,
			wantStatements: 0,
		},

		// 8. Unclosed block
		{
			name:           "restore to opening brace on bad syntax with unclosed block",
			input:          "package mypkg as 123 {\n  class InnerClass\n",
			kwType:         keyword.Package,
			bs:             expectBlockUnclosed,
			wantStatements: 1,
		},
		{
			name:           "restore to opening brace on unclosed container block",
			input:          "package mypkg {\n  class InnerClass\n",
			kwType:         keyword.Package,
			bs:             expectBlockUnclosed,
			wantStatements: 1,
		},

		// 9. Other container keywords
		{
			name:           "restore to opening brace on bad name token for node with block",
			input:          "node 123 {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Node,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
		{
			name:           "restore to newline on invalid alias for folder without block",
			input:          "folder myfolder as 123\nclass NextClass",
			kwType:         keyword.Folder,
			bs:             expectNoBlock,
			wantStatements: 0,
		},
		{
			name:           "restore to opening brace on extra junk before block for database",
			input:          "database mydb extra junk {\n  class InnerClass\n}\nclass NextClass",
			kwType:         keyword.Database,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},

		// 10. Unexpected token in body
		{
			name:           "package unexpected token in body with block",
			input:          "package mypkg {\n  @invalid\n}\nclass NextClass",
			kwType:         keyword.Package,
			bs:             expectBlockClosed,
			wantStatements: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  tokenizer.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}

			// 1. Consume and validate keyword
			tok := p.stream.Emit()
			require.Equal(t, tc.kwType, keyword.Classify(tok.Literal))

			// 2. Parse unit under test
			got := p.parseContainer(tok)
			require.NotNil(t, got)
			require.True(t, p.ast.HasErrors(), "expected parser to record diagnostics on bad syntax")

			// 3. Verify node body state via AST helper assertions
			switch tc.bs {
			case expectNoBlock:
				assertBlockBodyless(t, got)
				require.Empty(t, got.Statements)
			case expectBlockClosed:
				assertBlockClosed(t, got)
				require.Len(t, got.Statements, tc.wantStatements)
			case expectBlockUnclosed:
				assertBlockUnclosed(t, got)
				require.Len(t, got.Statements, tc.wantStatements)
			}

			// 4. Verify stream position after recovery
			if tc.bs == expectBlockUnclosed {
				require.True(t, p.stream.AssertType(tokenizer.EOF), "expected stream to reach EOF on unclosed block")
			} else {
				// Next token must be the separating newline
				require.Equal(t, tokenizer.NEWLINE, p.stream.PeekTokenAt(0).Type, "expected next token to be NEWLINE after recovery")

				// 5. Canary assertion: consume newline and parse subsequent statement
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
