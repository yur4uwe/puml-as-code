package tokenizer

import (
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
				"MODIFIER:abstract",
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
			Class3 -right-> Class4`,
			expectedTokens: []string{
				"IDENTIFIER:Class1",
				"RELATIONSHIP:-left->",
				"IDENTIFIER:Class2",
				"NEWLINE:\n",
				"IDENTIFIER:Class3",
				"RELATIONSHIP:-right->",
				"IDENTIFIER:Class4",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lex := NewLexer(tc.input)

			out := []string{}
			for t := lex.NextToken(); t.Type != EOF; t = lex.NextToken() {
				out = append(out, t.Type.String()+":"+t.Literal)
			}

			require.Equal(t, tc.expectedTokens, out, "incorrect tokens for test case: %s", tc.name)
		})
	}
}
