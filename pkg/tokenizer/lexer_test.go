package tokenizer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedTokens []string
	}{
		{
			name: "SimpleClassDeclaration",
			input: `class Foo {
				+id : int
  				-name : string
  				foo() : void
  			}`,
			expectedTokens: []string{
				"CLASS:class",
				"IDENTIFIER:Foo",
				"{:{",
				"NEWLINE:\n",
				"VISIBILITY:+",
				"IDENTIFIER:id",
				":::",
				"IDENTIFIER:int",
				"NEWLINE:\n",
				"VISIBILITY:-",
				"IDENTIFIER:name",
				":::",
				"IDENTIFIER:string",
				"NEWLINE:\n",
				"IDENTIFIER:foo",
				"(:(",
				"):)",
				":::",
				"IDENTIFIER:void",
				"NEWLINE:\n",
				"}:}",
			},
		},
		{
			name:  "RelationshipLexing",
			input: `User "1" -- "0..*" Order : places`,
			expectedTokens: []string{
				"IDENTIFIER:User",
				"STRING:\"1\"",
				"RELATIONSHIP:--",
				"STRING:\"0..*\"",
				"IDENTIFIER:Order",
				":::",
				"IDENTIFIER:places",
			},
		},
		{
			name: "EscapedName",
			input: `class Foo {
	  			\~bar()
			}`,
			expectedTokens: []string{
				"CLASS:class",
				"IDENTIFIER:Foo",
				"{:{",
				"NEWLINE:\n",
				"IDENTIFIER:~bar",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"}:}",
			},
		},
		{
			name: "ClassVisibility",
			input: `@startuml
			-class "private Class" {}
			#class "protected Class" {}
			~class "package private Class" {}
			+class "public Class" {}
			@enduml`,
			expectedTokens: []string{
				"@startuml:startuml",
				"VISIBILITY:-",
				"CLASS:class",
				"STRING:\"private Class\"",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"VISIBILITY:#",
				"CLASS:class",
				"STRING:\"protected Class\"",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"VISIBILITY:~",
				"CLASS:class",
				"STRING:\"package private Class\"",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"VISIBILITY:+",
				"CLASS:class",
				"STRING:\"public Class\"",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"@enduml:enduml",
			},
		},
		{
			name: "ClassModifiers",
			input: `abstract class Foo {
	  			{abstract} method()
				{static} method()
			}`,
			expectedTokens: []string{
				"ABSTRACT:abstract",
				"CLASS:class",
				"IDENTIFIER:Foo",
				"{:{",
				"NEWLINE:\n",
				"MODIFIER:{abstract}",
				"IDENTIFIER:method",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"MODIFIER:{static}",
				"IDENTIFIER:method",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"}:}",
			},
		},
		{
			name: "Enum",
			input: `enum {
				DAYS
				HOURS
				MINUTES
			}`,
			expectedTokens: []string{
				"ENUM:enum",
				"{:{",
				"NEWLINE:\n",
				"IDENTIFIER:DAYS",
				"NEWLINE:\n",
				"IDENTIFIER:HOURS",
				"NEWLINE:\n",
				"IDENTIFIER:MINUTES",
				"NEWLINE:\n",
				"}:}",
			},
		},
		{
			name: "Package",
			input: `package {
				class Foo {
					method()
				}
			}`,
			expectedTokens: []string{
				"PACKAGE:package",
				"{:{",
				"NEWLINE:\n",
				"CLASS:class",
				"IDENTIFIER:Foo",
				"{:{",
				"NEWLINE:\n",
				"IDENTIFIER:method",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"}:}",
				"NEWLINE:\n",
				"}:}",
			},
		},
		{
			name: "Alias",
			input: `class "Foo Bar" as Foo {
				method()
			}
			class Baz as "Baz Faz" {
				method()
			}`,
			expectedTokens: []string{
				"CLASS:class",
				"STRING:\"Foo Bar\"",
				"ALIAS:as",
				"IDENTIFIER:Foo",
				"{:{",
				"NEWLINE:\n",
				"IDENTIFIER:method",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"}:}",
				"NEWLINE:\n",
				"CLASS:class",
				"IDENTIFIER:Baz",
				"ALIAS:as",
				"STRING:\"Baz Faz\"",
				"{:{",
				"NEWLINE:\n",
				"IDENTIFIER:method",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"}:}",
			},
		},
		{
			name: "Relationship",
			input: `class Foo {
				method()
			}
			class Bar {
				method()
			}
			Foo -- Bar`,
			expectedTokens: []string{
				"CLASS:class",
				"IDENTIFIER:Foo",
				"{:{",
				"NEWLINE:\n",
				"IDENTIFIER:method",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"}:}",
				"NEWLINE:\n",
				"CLASS:class",
				"IDENTIFIER:Bar",
				"{:{",
				"NEWLINE:\n",
				"IDENTIFIER:method",
				"(:(",
				"):)",
				"NEWLINE:\n",
				"}:}",
				"NEWLINE:\n",
				"IDENTIFIER:Foo",
				"RELATIONSHIP:--",
				"IDENTIFIER:Bar",
			},
		},
		{
			name: "Stereotype",
			input: `class System << (S,#FF7700) Singleton >>
			class Date << (D,orchid) >>`,
			expectedTokens: []string{
				"CLASS:class",
				"IDENTIFIER:System",
				"STEREOTYPE:<< (S,#FF7700) Singleton >>",
				"NEWLINE:\n",
				"CLASS:class",
				"IDENTIFIER:Date",
				"STEREOTYPE:<< (D,orchid) >>",
			},
		},
		{
			name: "DecoratedRelationships",
			input: `Class21 #-- Class22
			Class23 x-- Class24
			Class25 }-- Class26
			Class27 +-- Class28
			Class29 ^-- Class30`,
			expectedTokens: []string{
				"IDENTIFIER:Class21",
				"RELATIONSHIP:#--",
				"IDENTIFIER:Class22",
				"NEWLINE:\n",
				"IDENTIFIER:Class23",
				"RELATIONSHIP:x--",
				"IDENTIFIER:Class24",
				"NEWLINE:\n",
				"IDENTIFIER:Class25",
				"RELATIONSHIP:}--",
				"IDENTIFIER:Class26",
				"NEWLINE:\n",
				"IDENTIFIER:Class27",
				"RELATIONSHIP:+--",
				"IDENTIFIER:Class28",
				"NEWLINE:\n",
				"IDENTIFIER:Class29",
				"RELATIONSHIP:^--",
				"IDENTIFIER:Class30",
			},
		},
		{
			name: "RelationsWithDirections",
			input: `Class1 -left-> Class2
			Class3 -r-> Class4`,
			expectedTokens: []string{
				"IDENTIFIER:Class1",
				"RELATIONSHIP:-left->",
				"IDENTIFIER:Class2",
				"NEWLINE:\n",
				"IDENTIFIER:Class3",
				"RELATIONSHIP:-r->",
				"IDENTIFIER:Class4",
			},
		},
		{
			name: "Keywords",
			input: `class Foo {}
			interface Bar {}
			enum Status {}
			struct Point {}
			record User {}
			dataclass Person {}
			exception Error {}
			protocol Sync {}
			package Pkg {}
			annotation Test
			note "test"
			hide methods
			show fields
			remove visibility
			restore defaults
			skinparam backgroundColor white
			set lineStyle ortho
			together {
			  class A
			  class B
			}`,
			expectedTokens: []string{
				"CLASS:class",
				"IDENTIFIER:Foo",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"INTERFACE:interface",
				"IDENTIFIER:Bar",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"ENUM:enum",
				"IDENTIFIER:Status",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"STRUCT:struct",
				"IDENTIFIER:Point",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"RECORD:record",
				"IDENTIFIER:User",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"DATACLASS:dataclass",
				"IDENTIFIER:Person",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"EXCEPTION:exception",
				"IDENTIFIER:Error",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"PROTOCOL:protocol",
				"IDENTIFIER:Sync",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"PACKAGE:package",
				"IDENTIFIER:Pkg",
				"{:{",
				"}:}",
				"NEWLINE:\n",
				"ANNOTATION:annotation",
				"IDENTIFIER:Test",
				"NEWLINE:\n",
				"NOTE:note",
				"STRING:\"test\"",
				"NEWLINE:\n",
				"HIDE:hide",
				"IDENTIFIER:methods",
				"NEWLINE:\n",
				"SHOW:show",
				"IDENTIFIER:fields",
				"NEWLINE:\n",
				"REMOVE:remove",
				"IDENTIFIER:visibility",
				"NEWLINE:\n",
				"RESTORE:restore",
				"IDENTIFIER:defaults",
				"NEWLINE:\n",
				"SKINPARAM:skinparam",
				"IDENTIFIER:backgroundColor",
				"IDENTIFIER:white",
				"NEWLINE:\n",
				"SET:set",
				"IDENTIFIER:lineStyle",
				"IDENTIFIER:ortho",
				"NEWLINE:\n",
				"TOGETHER:together",
				"{:{",
				"NEWLINE:\n",
				"CLASS:class",
				"IDENTIFIER:A",
				"NEWLINE:\n",
				"CLASS:class",
				"IDENTIFIER:B",
				"NEWLINE:\n",
				"}:}",
			},
		},
		{
			name: "Separators",
			input: `class Foo {
			  -- Sep with dashes --
			  .. Sep with dots ..
			  == Sep with equals ==
              __ Sep with lodashes __
              ' other seps:
              --
              ..
              ==
              __
			}`,
			expectedTokens: []string{
				"CLASS:class", "IDENTIFIER:Foo", "{:{", "NEWLINE:\n",
				"SEPARATOR:--", "IDENTIFIER:Sep with dashes", "SEPARATOR:--", "NEWLINE:\n",
				"SEPARATOR:..", "IDENTIFIER:Sep with dots", "SEPARATOR:..", "NEWLINE:\n",
				"SEPARATOR:==", "IDENTIFIER:Sep with equals", "SEPARATOR:==", "NEWLINE:\n",
				"SEPARATOR:__", "IDENTIFIER:Sep with lodashes", "SEPARATOR:__", "NEWLINE:\n",
				"COMMENT:other seps:", "NEWLINE:\n",
				"SEPARATOR:--", "NEWLINE:\n",
				"SEPARATOR:..", "NEWLINE:\n",
				"SEPARATOR:==", "NEWLINE:\n",
				"SEPARATOR:__", "NEWLINE:\n",
				"}:}",
			},
		},
		{
			name: "Note",
			input: `note left of Foo
				This is a note
				with multiline content
			end note
			note "String note with alias" as Alias
			class Baz
			note right : On Last defined class with colon`,
			expectedTokens: []string{
				"NOTE:note", "NOTE_DIRECTION:left", "NOTE_POSITION:of", "IDENTIFIER:Foo", "NEWLINE:\n",
				"IDENTIFIER:This is a note", "NEWLINE:\n",
				"IDENTIFIER:with multiline content", "NEWLINE:\n",
				"END_BLOCK:end note", "NEWLINE:\n",
				"NOTE:note", "STRING:\"String note with alias\"", "ALIAS:as", "IDENTIFIER:Alias", "NEWLINE:\n",
				"CLASS:class", "IDENTIFIER:Baz", "NEWLINE:\n",
				"NOTE:note", "NOTE_DIRECTION:right", ":::", "IDENTIFIER:On Last defined class with colon",
			},
		},
		{
			name: "AddMethodOrField",
			input: `Object : equals()
			ArrayList : Object[] elementData
			ArrayList : size()`,
			expectedTokens: []string{
				"IDENTIFIER:Object", ":::", "IDENTIFIER:equals", "(:(", "):)", "NEWLINE:\n",
				"IDENTIFIER:ArrayList", ":::", "IDENTIFIER:Object", "[:[", "]:]", "IDENTIFIER:elementData", "NEWLINE:\n",
				"IDENTIFIER:ArrayList", ":::", "IDENTIFIER:size", "(:(", "):)",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lex := NewLexer(tc.input)

			out := []string{}
			for tok := lex.NextToken(); tok.Type != EOF; tok = lex.NextToken() {
				out = append(out, tok.Type.String()+":"+tok.Literal)
				if tok.Literal == "" {
					t.Fatalf("%s: Empty literal, prone to infinite loops\npreceding tokens: %s", tc.name, strings.Join(out, ", "))
				}
			}

			require.Equal(t, tc.expectedTokens, out, "incorrect tokens for test case: %s", tc.name)
		})
	}
}
