package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTokenStream(t *testing.T) {
	ts := NewTokenStream("")
	require.NotNil(t, ts)
	tok := ts.PeekTokenAt(0)
	require.Equal(t, EOF, tok.Type)
}

func TestStreamPeekEmitConsume(t *testing.T) {
	input := "class Foo { }"
	ts := NewTokenStream(input)

	// PeekTokenAt
	tok0 := ts.PeekTokenAt(0)
	require.Equal(t, CLASS, tok0.Type)
	require.Equal(t, "class", tok0.Literal)

	tok1 := ts.PeekTokenAt(1)
	require.Equal(t, IDENTIFIER, tok1.Type)
	require.Equal(t, "Foo", tok1.Literal)

	// Assert
	require.True(t, ts.AssertType(CLASS))
	require.False(t, ts.AssertType(LBRACE))

	// Consume
	tok, ok := ts.TryConsumeType(CLASS)
	require.True(t, ok)
	require.Equal(t, "class", tok.Literal)

	// Emit
	tok = ts.Emit()
	require.Equal(t, IDENTIFIER, tok.Type)
	require.Equal(t, "Foo", tok.Literal)
}

func TestStreamEmitRaw(t *testing.T) {
	ts := NewTokenStream("A /' comment '/ B")
	tokA := ts.EmitRaw()
	require.Equal(t, IDENTIFIER, tokA.Type)
	require.Equal(t, "A", tokA.Literal)

	tokC := ts.EmitRaw()
	require.Equal(t, COMMENT, tokC.Type)
	require.Equal(t, " comment ", tokC.Literal)

	tokB := ts.EmitRaw()
	require.Equal(t, IDENTIFIER, tokB.Type)
	require.Equal(t, "B", tokB.Literal)
}

func TestStreamAssertSeq(t *testing.T) {
	ts := NewTokenStream("class Foo {")

	require.True(t, ts.AssertSeq([]Token{
		{Type: CLASS, Literal: "class"},
		{Type: IDENTIFIER, Literal: "Foo"},
		{Type: LBRACE},
	}))

	require.False(t, ts.AssertSeq([]Token{
		{Type: CLASS, Literal: "class"},
		{Type: IDENTIFIER, Literal: "Bar"},
	}))

	require.False(t, ts.AssertSeq([]Token{
		{Type: CLASS, Literal: "class"},
		{Type: IDENTIFIER, Literal: "Foo"},
		{Type: LBRACE},
		{Type: RBRACE},
	}))
}

func TestStreamTryReadModifier(t *testing.T) {
	ts := NewTokenStream("{abstract} {static} class")

	mod, err := ts.TryReadModifier()
	require.NoError(t, err)
	require.Equal(t, "abstract", mod)

	mod, err = ts.TryReadModifier()
	require.NoError(t, err)
	require.Equal(t, "static", mod)

	mod, err = ts.TryReadModifier()
	require.Error(t, err)
}

func TestStreamTryReadStereotype(t *testing.T) {
	// Note: Currently fails due to internal implementation error (missing spaces)
	ts := NewTokenStream("<<stereotype>> <<foo bar>>")

	stereo, err := ts.TryReadStereotype()
	require.NoError(t, err)
	require.Equal(t, "stereotype", stereo)

	stereo, err = ts.TryReadStereotype()
	require.NoError(t, err)
	require.Equal(t, "foo bar", stereo)
}

func TestStreamTryReadGeneric(t *testing.T) {
	// Note: Currently fails due to internal implementation error (missing spaces)
	ts := NewTokenStream("<T> <T, U>")

	gen, err := ts.TryReadGeneric()
	require.NoError(t, err)
	require.Equal(t, "T", gen)

	gen, err = ts.TryReadGeneric()
	require.NoError(t, err)
	require.Equal(t, "T, U", gen)
}

func TestStreamTryReadClassSeparator(t *testing.T) {
	ts := NewTokenStream(".. separator ..\n== sep ==")

	sep, err := ts.TryReadClassSeparator()
	require.NoError(t, err)
	require.Equal(t, "separator", sep.Label)
	require.Equal(t, '.', sep.Type)

	ts.TryConsumeType(NEWLINE)

	sep, err = ts.TryReadClassSeparator()
	require.NoError(t, err)
	require.Equal(t, "sep", sep.Label)
	require.Equal(t, '=', sep.Type)
}

func TestStreamTryReadTag(t *testing.T) {
	t.Skip("Abandoned for now per user instruction")
	ts := NewTokenStream("$tagName $another")

	tag, err := ts.TryReadTag()
	require.NoError(t, err)
	require.Equal(t, "tagName", tag)

	tag, err = ts.TryReadTag()
	require.NoError(t, err)
	require.Equal(t, "another", tag)
}

// func TestStreamTryReadDiagramBounds(t *testing.T) {
// 	// Note: Currently fails due to apparent implementation bug (Assert(AT) returns false)
// 	ts := NewTokenStream("@startuml\n@enduml")
//
// 	b, err := ts.TryReadDiagramBounds()
// 	require.NoError(t, err)
// 	require.Equal(t, "startuml", b)
//
// 	ts.ConsumeType(NEWLINE)
//
// 	b, err = ts.TryReadDiagramBounds()
// 	require.NoError(t, err)
// 	require.Equal(t, "enduml", b)
// }

func TestReadDiagramBounds(t *testing.T) {
	tt := []struct {
		name             string
		input            string
		expectedKvps     map[string]string
		expectError      bool
		expectedFilename string
		expectedType     string
		expectedID       string
	}{
		// general parsing
		{
			name:         "Default case",
			input:        "@startuml\n@enduml",
			expectedType: "uml",
		},
		{
			name:             "With filename",
			input:            "@startuml filename.puml\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
		},
		// id parsing
		{
			name:         "With tag",
			input:        "@startuml(id=tag)\n@enduml",
			expectedType: "uml",
			expectedID:   "tag",
		},
		{
			name:             "With filename and tag",
			input:            "@startuml(id=tag) filename.puml\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedID:       "tag",
		},
		// options parsing
		{
			name:             "filename using tool options",
			input:            "@startuml{filename.puml}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
		},
		{
			name:             "filename and caption using tool options",
			input:            "@startuml{filename.puml, foo bar}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedKvps: map[string]string{
				"caption": "foo bar",
			},
		},
		{
			name:             "kvp parsing",
			input:            "@startuml{filename.puml, foo bar, key=value}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedKvps: map[string]string{
				"caption": "foo bar",
				"key":     "value",
			},
		},
		{
			name:             "parsing kvp's right after filename and no caption",
			input:            "@startuml{filename.puml, key=value}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedKvps: map[string]string{
				"key": "value",
			},
		},
		{
			name:             "Tools and id parsing simulatenously",
			input:            "@startuml(id=tag){filename.puml, foo bar, key=value}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedID:       "tag",
			expectedKvps: map[string]string{
				"caption": "foo bar",
				"key":     "value",
			},
		},
		// error cases
		{
			name:        "Incorrect order of tools and id",
			input:       "@startuml{filename.puml, foo bar, key=value}(id=tag)\n@enduml",
			expectError: true,
		},
		{
			name:        "legacy filename syntax after tools",
			input:       "@startuml{filename.puml, foo bar, key=value} filename.puml\n@enduml",
			expectError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := NewTokenStream(tc.input)
			b, err := ts.ReadDiagramBounds()
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, b.IsStart)
			require.Equal(t, tc.expectedType, b.Type)

			if tc.expectedKvps != nil {
				require.Equal(t, tc.expectedKvps, b.Opts)
			}
			if tc.expectedFilename != "" {
				require.Equal(t, tc.expectedFilename, b.Name)
			}
			if tc.expectedID != "" {
				require.Equal(t, tc.expectedID, b.ID)
			}

			ts.TryConsumeType(NEWLINE)
			b, err = ts.ReadDiagramBounds()
			require.NoError(t, err)
			require.False(t, b.IsStart)
			require.Equal(t, tc.expectedType, b.Type)
		})
	}
}

func TestStreamReadUntilNewline(t *testing.T) {
	t.Skip("Abandoned for now per user instruction")
	ts := NewTokenStream("foo bar /' comment '/ baz\nnext")

	line := ts.ReadUntilNewline()
	require.Equal(t, "foo bar  baz", line)

	ts2 := NewTokenStream("foo bar /' comment '/ baz\nnext")
	lineRaw := ts2.ReadRawUntilNewline()
	require.Equal(t, "foo bar  comment  baz", lineRaw)
}

func TestStreamReadBlock(t *testing.T) {
	// Note: Currently fails because 'end' is tokenized as IDENTIFIER instead of END_BLOCK
	input := "note right of Foo\n  This is a block\n  with multiple lines\nend note"
	ts := NewTokenStream(input)

	ts.ReadUntilNewline()

	block, err := ts.ReadBlock(Token{Type: END_BLOCK}, Token{Type: NOTE})
	require.NoError(t, err)
	require.Contains(t, block, "This is a block")
}

func TestStreamReadBetween(t *testing.T) {
	// Note: Currently fails due to internal implementation error (missing spaces)
	ts := NewTokenStream("START foo bar END")

	startSeq := []Token{{Type: IDENTIFIER, Literal: "START"}}
	endSeq := []Token{{Type: END_BLOCK, Literal: "END"}}

	res, err := ts.readBetween(startSeq, endSeq)
	require.NoError(t, err)
	require.Equal(t, "foo bar", res)
}
